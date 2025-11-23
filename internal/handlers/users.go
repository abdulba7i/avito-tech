package handlers

import (
	"encoding/json"
	"net/http"
	"reviewer-service/internal/models"
	srvPR "reviewer-service/internal/services/pullrequests"
	srvTeams "reviewer-service/internal/services/teams"
	srvUsers "reviewer-service/internal/services/users"

	"github.com/google/uuid"
)

type UsersHandler struct {
	usersService        *srvUsers.Service
	pullRequestsService *srvPR.Service
	teamsService        *srvTeams.Service
}

func NewUsersHandler(usersService *srvUsers.Service, prService *srvPR.Service, teamsService *srvTeams.Service) *UsersHandler {
	return &UsersHandler{
		usersService:        usersService,
		pullRequestsService: prService,
		teamsService:        teamsService,
	}
}

type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type UserResponse struct {
	User models.User `json:"user"`
}

func (h *UsersHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SetIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "INVALID_REQUEST", "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		respondError(w, "INVALID_REQUEST", "Invalid user_id", http.StatusBadRequest)
		return
	}

	if err := h.usersService.SetIsActive(r.Context(), userID, req.IsActive); err != nil {
		respondError(w, "NOT_FOUND", "user not found", http.StatusNotFound)
		return
	}

	user, err := h.usersService.GetUser(r.Context(), userID)
	if err != nil {
		respondError(w, "NOT_FOUND", "user not found", http.StatusNotFound)
		return
	}
	if user == nil {
		respondError(w, "NOT_FOUND", "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UserResponse{User: *user})
}

type GetReviewResponse struct {
	UserID       string               `json:"user_id"`
	PullRequests []models.PullRequest `json:"pull_requests"`
}

func (h *UsersHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		respondError(w, "INVALID_REQUEST", "user_id is required", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondError(w, "INVALID_REQUEST", "Invalid user_id", http.StatusBadRequest)
		return
	}

	prs, err := h.usersService.GetAssignedPullRequests(r.Context(), userID)
	if err != nil {
		respondError(w, "NOT_FOUND", "user not found", http.StatusNotFound)
		return
	}

	shortPRs := make([]models.PullRequest, len(prs))
	for i, pr := range prs {
		shortPRs[i] = models.PullRequest{
			ID:        pr.ID,
			Name:      pr.Name,
			AuthorID:  pr.AuthorID,
			Status:    pr.Status,
			CreatedAt: pr.CreatedAt,
			MergedAt:  pr.MergedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetReviewResponse{
		UserID:       userIDStr,
		PullRequests: shortPRs,
	})
}
