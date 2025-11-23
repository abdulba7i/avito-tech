package users

import (
	"context"
	"reviewer-service/internal/models"
	"reviewer-service/internal/repository/users"

	"github.com/google/uuid"
)

type Service struct {
	repo users.Repository
}

func New(repo users.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(ctx context.Context, user models.User) error {
	return s.repo.CreateUser(ctx, user)
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetTeamUsers(ctx context.Context, team string) ([]models.User, error) {
	users, err := s.repo.GetActiveByTeam(ctx, team)
	if err != nil {
		return nil, err
	}
	result := make([]models.User, len(users))
	for i, u := range users {
		result[i] = *u
	}
	return result, nil
}

func (s *Service) SetIsActive(ctx context.Context, id uuid.UUID, isActive bool) error {
	return s.repo.SetIsActive(ctx, id, isActive)
}

func (s *Service) GetAssignedPullRequests(ctx context.Context, userID uuid.UUID) ([]*models.PullRequest, error) {
	return s.repo.GetAssignedPullRequests(ctx, userID)
}

func (s *Service) SetTeamInactive(ctx context.Context, teamName string) error {
	return s.repo.SetTeamInactive(ctx, teamName)
}
