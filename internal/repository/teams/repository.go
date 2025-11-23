package teams

import (
	"context"
	"reviewer-service/internal/models"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, team models.Team) error
	GetTeamByName(ctx context.Context, name string) (*models.Team, error)
	AddUsersToTeam(ctx context.Context, teamName string, users []models.User) error
}
