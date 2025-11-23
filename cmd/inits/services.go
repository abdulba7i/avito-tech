package inits

import (
	"context"
	"database/sql"

	"reviewer-service/internal/models"
	"reviewer-service/internal/repository/users"
	srvPR "reviewer-service/internal/services/pullrequests"
	srvStats "reviewer-service/internal/services/statistics"
	srvTeams "reviewer-service/internal/services/teams"
	srvUsers "reviewer-service/internal/services/users"
	stPR "reviewer-service/internal/storage"

	"github.com/google/uuid"
)

type Services struct {
	PullRequests *srvPR.Service
	Teams        *srvTeams.Service
	Users        *srvUsers.Service
	Statistics   *srvStats.Service
}

type usersRepoAdapter struct {
	repo users.Repository
}

func (a *usersRepoAdapter) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	return a.repo.GetByID(ctx, userID)
}

func InitServices(db *sql.DB) *Services {
	// Storage
	prRepo := stPR.NewPullRequestsRepo(db)
	teamsRepo := stPR.NewTeamsRepo(db)
	usersRepo := stPR.NewUsersRepo(db)
	statsRepo := stPR.NewStatisticsRepo(db)

	// Адаптер для pullrequests service
	usersRepoAdapter := &usersRepoAdapter{repo: usersRepo}

	return &Services{
		PullRequests: srvPR.New(prRepo, usersRepoAdapter),
		Teams:        srvTeams.New(teamsRepo),
		Users:        srvUsers.New(usersRepo),
		Statistics:   srvStats.New(statsRepo),
	}
}
