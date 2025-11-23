package handlers

import (
	"encoding/json"
	"net/http"
	"reviewer-service/internal/models"
	srvTeams "reviewer-service/internal/services/teams"
	"strings"
)

type TeamsHandler struct {
	teamsService *srvTeams.Service
}

func NewTeamsHandler(teamsService *srvTeams.Service) *TeamsHandler {
	return &TeamsHandler{teamsService: teamsService}
}

type TeamRequest struct {
	TeamName string      `json:"team_name"`
	Members  []UserInput `json:"members"`
}

type UserInput struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamResponse struct {
	Team models.Team `json:"team"`
}

func (h *TeamsHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "INVALID_REQUEST", "Invalid request body", http.StatusBadRequest)
		return
	}

	members := make([]models.User, len(req.Members))
	for i, m := range req.Members {
		members[i] = models.User{
			ID:       m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	team := models.Team{
		Name:    req.TeamName,
		Members: members,
	}

	if err := h.teamsService.CreateTeam(r.Context(), team); err != nil {
		if strings.Contains(err.Error(), "team_name already exists") {
			respondError(w, "TEAM_EXISTS", "team_name already exists", http.StatusBadRequest)
			return
		}
		respondError(w, "INTERNAL_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(TeamResponse{Team: team})
}

func (h *TeamsHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		respondError(w, "INVALID_REQUEST", "team_name is required", http.StatusBadRequest)
		return
	}

	team, err := h.teamsService.GetTeam(r.Context(), teamName)
	if err != nil {
		respondError(w, "NOT_FOUND", "team not found", http.StatusNotFound)
		return
	}
	if team == nil {
		respondError(w, "NOT_FOUND", "team not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

