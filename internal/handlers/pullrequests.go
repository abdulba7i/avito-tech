package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"reviewer-service/internal/models"
	srvPR "reviewer-service/internal/services/pullrequests"
	srvTeams "reviewer-service/internal/services/teams"
	srvUsers "reviewer-service/internal/services/users"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PullRequestsHandler struct {
	prService    *srvPR.Service
	usersService *srvUsers.Service
	teamsService *srvTeams.Service
}

func NewPullRequestsHandler(prService *srvPR.Service, usersService *srvUsers.Service, teamsService *srvTeams.Service) *PullRequestsHandler {
	return &PullRequestsHandler{
		prService:    prService,
		usersService: usersService,
		teamsService: teamsService,
	}
}

type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type PRResponse struct {
	PR models.PullRequest `json:"pr"`
}

func (h *PullRequestsHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "INVALID_REQUEST", "Invalid request body", http.StatusBadRequest)
		return
	}

	authorID, err := uuid.Parse(req.AuthorID)
	if err != nil {
		respondError(w, "INVALID_REQUEST", "Invalid author_id", http.StatusBadRequest)
		return
	}

	author, err := h.usersService.GetUser(r.Context(), authorID)
	if err != nil || author == nil {
		respondError(w, "NOT_FOUND", "author not found", http.StatusNotFound)
		return
	}

	team, err := h.teamsService.GetTeam(r.Context(), author.TeamName)
	if err != nil || team == nil {
		respondError(w, "NOT_FOUND", "team not found", http.StatusNotFound)
		return
	}

	prID := req.PullRequestID
	if prID == "" {
		prID = uuid.New().String()
	}

	pr := models.PullRequest{
		ID:        prID,
		Name:      req.PullRequestName,
		AuthorID:  req.AuthorID,
		Status:    models.PullRequestStatusOpen,
		CreatedAt: time.Now(),
		Reviewers: []string{},
	}

	if err := h.prService.CreatePullRequest(r.Context(), pr); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "unique constraint") {
			respondError(w, "PR_EXISTS", "PR id already exists", http.StatusConflict)
			return
		}
		respondError(w, "INTERNAL_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	reviewers := h.selectReviewers(team.Members, req.AuthorID, 2)
	if len(reviewers) > 0 {
		if err := h.prService.AssignReviewers(r.Context(), prID, reviewers); err != nil {
		}
	}

	createdPR, err := h.prService.GetPullRequest(r.Context(), prID)
	if err != nil {
		respondError(w, "NOT_FOUND", "PR not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(PRResponse{PR: *createdPR})
}

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

func (h *PullRequestsHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "INVALID_REQUEST", "Invalid request body", http.StatusBadRequest)
		return
	}

	pr, err := h.prService.GetPullRequest(r.Context(), req.PullRequestID)
	if err != nil || pr == nil {
		respondError(w, "NOT_FOUND", "PR not found", http.StatusNotFound)
		return
	}

	if err := h.prService.Merge(r.Context(), req.PullRequestID); err != nil {
		respondError(w, "INTERNAL_ERROR", "Failed to merge PR", http.StatusInternalServerError)
		return
	}

	mergedPR, err := h.prService.GetPullRequest(r.Context(), req.PullRequestID)
	if err != nil {
		respondError(w, "NOT_FOUND", "PR not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PRResponse{PR: *mergedPR})
}

type ReassignRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

type ReassignResponse struct {
	PR         models.PullRequest `json:"pr"`
	ReplacedBy string             `json:"replaced_by"`
}

func (h *PullRequestsHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "INVALID_REQUEST", "Invalid request body", http.StatusBadRequest)
		return
	}

	pr, err := h.prService.GetPullRequest(r.Context(), req.PullRequestID)
	if err != nil || pr == nil {
		respondError(w, "NOT_FOUND", "PR not found", http.StatusNotFound)
		return
	}

	if pr.Status == models.PullRequestStatusMerged {
		respondError(w, "PR_MERGED", "cannot reassign on merged PR", http.StatusConflict)
		return
	}

	found := false
	for _, reviewerID := range pr.Reviewers {
		if reviewerID == req.OldUserID {
			found = true
			break
		}
	}
	if !found {
		respondError(w, "NOT_ASSIGNED", "reviewer is not assigned to this PR", http.StatusConflict)
		return
	}

	oldUserID, err := uuid.Parse(req.OldUserID)
	if err != nil {
		respondError(w, "INVALID_REQUEST", "Invalid old_user_id", http.StatusBadRequest)
		return
	}

	oldUser, err := h.usersService.GetUser(r.Context(), oldUserID)
	if err != nil || oldUser == nil {
		respondError(w, "NOT_FOUND", "old reviewer not found", http.StatusNotFound)
		return
	}

	team, err := h.teamsService.GetTeam(r.Context(), oldUser.TeamName)
	if err != nil || team == nil {
		respondError(w, "NOT_FOUND", "team not found", http.StatusNotFound)
		return
	}

	excludeIDs := map[string]bool{
		req.OldUserID: true,
		pr.AuthorID:   true,
	}
	for _, reviewerID := range pr.Reviewers {
		excludeIDs[reviewerID] = true
	}

	var candidates []string
	for _, member := range team.Members {
		if !excludeIDs[member.ID] && member.IsActive {
			candidates = append(candidates, member.ID)
		}
	}

	if len(candidates) == 0 {
		respondError(w, "NO_CANDIDATE", "no active replacement candidate in team", http.StatusConflict)
		return
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	newReviewerID := candidates[0]

	if err := h.prService.ReassignReviewer(r.Context(), req.PullRequestID, req.OldUserID, newReviewerID); err != nil {
		respondError(w, "INTERNAL_ERROR", "Failed to reassign reviewer", http.StatusInternalServerError)
		return
	}

	updatedPR, err := h.prService.GetPullRequest(r.Context(), req.PullRequestID)
	if err != nil {
		respondError(w, "NOT_FOUND", "PR not found", http.StatusNotFound)
		return
	}

	filteredReviewers := []string{}
	for _, reviewerID := range updatedPR.Reviewers {
		if reviewerID != updatedPR.AuthorID {
			filteredReviewers = append(filteredReviewers, reviewerID)
		}
	}
	updatedPR.Reviewers = filteredReviewers

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ReassignResponse{
		PR:         *updatedPR,
		ReplacedBy: newReviewerID,
	})
}

func (h *PullRequestsHandler) selectReviewers(members []models.User, excludeUserID string, maxCount int) []string {
	var candidates []string
	for _, member := range members {
		if member.ID != excludeUserID && member.IsActive {
			candidates = append(candidates, member.ID)
		}
	}

	if len(candidates) == 0 {
		return []string{}
	}

	if len(candidates) <= maxCount {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})
		return candidates
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	return candidates[:maxCount]
}
