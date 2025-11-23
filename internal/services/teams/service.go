package teams

import (
	"context"
	"reviewer-service/internal/models"
)

type TeamsRepository interface {
	CreateTeam(ctx context.Context, team models.Team) error
	GetTeamByName(ctx context.Context, name string) (*models.Team, error)
}

type Service struct {
	repo TeamsRepository
}

func New(repo TeamsRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTeam(ctx context.Context, team models.Team) error {
	return s.repo.CreateTeam(ctx, team)
}

func (s *Service) GetTeam(ctx context.Context, name string) (*models.Team, error) {
	return s.repo.GetTeamByName(ctx, name)
}
