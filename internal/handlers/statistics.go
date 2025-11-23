package handlers

import (
	"encoding/json"
	"net/http"
	"reviewer-service/internal/services/statistics"
	"reviewer-service/internal/storage"
)

type StatisticsHandler struct {
	statsService *statistics.Service
}

func NewStatisticsHandler(statsService *statistics.Service) *StatisticsHandler {
	return &StatisticsHandler{statsService: statsService}
}

type StatisticsResponse struct {
	UserAssignments []storage.UserAssignmentStats `json:"user_assignments"`
	PRStats         *storage.PRStats             `json:"pr_stats"`
}

func (h *StatisticsHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userStats, err := h.statsService.GetUserAssignmentStats(r.Context())
	if err != nil {
		respondError(w, "INTERNAL_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	prStats, err := h.statsService.GetPRStats(r.Context())
	if err != nil {
		respondError(w, "INTERNAL_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(StatisticsResponse{
		UserAssignments: userStats,
		PRStats:         prStats,
	})
}

