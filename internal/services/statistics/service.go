package statistics

import (
	"context"
	"reviewer-service/internal/storage"
)

type StatisticsRepository interface {
	GetUserAssignmentStats(ctx context.Context) ([]storage.UserAssignmentStats, error)
	GetPRStats(ctx context.Context) (*storage.PRStats, error)
}

type Service struct {
	repo StatisticsRepository
}

func New(repo StatisticsRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetUserAssignmentStats(ctx context.Context) ([]storage.UserAssignmentStats, error) {
	return s.repo.GetUserAssignmentStats(ctx)
}

func (s *Service) GetPRStats(ctx context.Context) (*storage.PRStats, error) {
	return s.repo.GetPRStats(ctx)
}

