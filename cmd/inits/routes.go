package inits

import (
	"net/http"
	"reviewer-service/internal/handlers"
)

func SetupRoutes(services *Services) http.Handler {
	mux := http.NewServeMux()

	// Handlers
	teamsHandler := handlers.NewTeamsHandler(services.Teams)
	usersHandler := handlers.NewUsersHandler(services.Users, services.PullRequests, services.Teams)
	prHandler := handlers.NewPullRequestsHandler(services.PullRequests, services.Users, services.Teams)
	statsHandler := handlers.NewStatisticsHandler(services.Statistics)

	// Routes
	mux.HandleFunc("/team/add", teamsHandler.AddTeam)
	mux.HandleFunc("/team/get", teamsHandler.GetTeam)

	mux.HandleFunc("/users/setIsActive", usersHandler.SetIsActive)
	mux.HandleFunc("/users/getReview", usersHandler.GetReview)
	mux.HandleFunc("/users/bulkDeactivateTeam", usersHandler.BulkDeactivateTeam)

	mux.HandleFunc("/pullRequest/create", prHandler.CreatePR)
	mux.HandleFunc("/pullRequest/merge", prHandler.MergePR)
	mux.HandleFunc("/pullRequest/reassign", prHandler.ReassignReviewer)

	// Statistics
	mux.HandleFunc("/statistics", statsHandler.GetStatistics)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return mux
}
