package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"reviewer-service/internal/models"
	"time"

	"github.com/google/uuid"
)

type BulkDeactivateRequest struct {
	TeamName string `json:"team_name"`
}

type BulkDeactivateResponse struct {
	DeactivatedUsers []string             `json:"deactivated_users"`
	ReassignedPRs    []PRReassignmentInfo `json:"reassigned_prs"`
	DurationMs       int64                `json:"duration_ms"`
}

type PRReassignmentInfo struct {
	PRID         string   `json:"pr_id"`
	Replaced     []string `json:"replaced"`
	NewReviewers []string `json:"new_reviewers"`
}

func (h *UsersHandler) BulkDeactivateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startTime := time.Now()

	var req BulkDeactivateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "INVALID_REQUEST", "Invalid request body", http.StatusBadRequest)
		return
	}

	teamModel, err := h.teamsService.GetTeam(r.Context(), req.TeamName)
	if err != nil || teamModel == nil {
		respondError(w, "NOT_FOUND", "team not found", http.StatusNotFound)
		return
	}
	team := teamModel.Members

	var activeUserIDs []string
	for _, user := range team {
		if user.IsActive {
			activeUserIDs = append(activeUserIDs, user.ID)
		}
	}

	if len(activeUserIDs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(BulkDeactivateResponse{
			DeactivatedUsers: []string{},
			ReassignedPRs:    []PRReassignmentInfo{},
			DurationMs:       time.Since(startTime).Milliseconds(),
		})
		return
	}

	inactiveUUIDs := make([]uuid.UUID, len(activeUserIDs))
	for i, id := range activeUserIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			respondError(w, "INVALID_REQUEST", "Invalid user_id", http.StatusBadRequest)
			return
		}
		inactiveUUIDs[i] = uid
	}

	openPRs, err := h.pullRequestsService.GetOpenPRsWithInactiveReviewers(r.Context(), activeUserIDs)
	if err != nil {
		respondError(w, "INTERNAL_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	var reassignments []models.ReviewerReassignment
	var reassignedPRs []PRReassignmentInfo

	for _, pr := range openPRs {
		authorID, err := uuid.Parse(pr.AuthorID)
		if err != nil {
			continue
		}
		author, err := h.usersService.GetUser(r.Context(), authorID)
		if err != nil || author == nil {
			continue
		}

		authorTeamModel, err := h.teamsService.GetTeam(r.Context(), author.TeamName)
		if err != nil {
			continue
		}
		authorTeam := authorTeamModel.Members

		var inactiveReviewers []string
		for _, reviewerID := range pr.Reviewers {
			for _, inactiveID := range activeUserIDs {
				if reviewerID == inactiveID {
					inactiveReviewers = append(inactiveReviewers, reviewerID)
					break
				}
			}
		}

		if len(inactiveReviewers) == 0 {
			continue
		}

		excludeIDs := map[string]bool{
			pr.AuthorID: true,
		}
		for _, reviewerID := range pr.Reviewers {
			excludeIDs[reviewerID] = true
		}

		var candidates []string
		for _, member := range authorTeam {
			if !excludeIDs[member.ID] && member.IsActive {
				candidates = append(candidates, member.ID)
			}
		}

		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})

		var replaced []string
		var newReviewers []string
		for i, inactiveReviewerID := range inactiveReviewers {
			if i >= len(candidates) {
				break
			}

			oldUUID, _ := uuid.Parse(inactiveReviewerID)
			newUUID, _ := uuid.Parse(candidates[i])

			reassignments = append(reassignments, models.ReviewerReassignment{
				PRID:          pr.ID,
				OldReviewerID: oldUUID,
				NewReviewerID: newUUID,
			})
			replaced = append(replaced, inactiveReviewerID)
			newReviewers = append(newReviewers, candidates[i])
		}

		if len(replaced) > 0 {
			reassignedPRs = append(reassignedPRs, PRReassignmentInfo{
				PRID:         pr.ID,
				Replaced:     replaced,
				NewReviewers: newReviewers,
			})
		}
	}

	if len(reassignments) > 0 {
		if err := h.pullRequestsService.BulkReassignReviewers(r.Context(), reassignments); err != nil {
			respondError(w, "INTERNAL_ERROR", "Failed to reassign reviewers", http.StatusInternalServerError)
			return
		}
	}

	if err := h.usersService.SetTeamInactive(r.Context(), req.TeamName); err != nil {
		respondError(w, "INTERNAL_ERROR", "Failed to deactivate users", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime).Milliseconds()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BulkDeactivateResponse{
		DeactivatedUsers: activeUserIDs,
		ReassignedPRs:    reassignedPRs,
		DurationMs:       duration,
	})
}
