package users

import (
	"context"
	"reviewer-service/internal/models"

	"github.com/google/uuid"
)

type Repository interface {
	CreateUser(ctx context.Context, user models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	SetIsActive(ctx context.Context, id uuid.UUID, isActive bool) error
	SetTeamInactive(ctx context.Context, teamName string) error
	GetActiveByTeam(ctx context.Context, teamName string) ([]*models.User, error)
	GetAssignedPullRequests(ctx context.Context, userID uuid.UUID) ([]*models.PullRequest, error)
}
