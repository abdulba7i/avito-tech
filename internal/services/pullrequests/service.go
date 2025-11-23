package pullrequests

import (
	"context"
	"errors"
	"fmt"
	"reviewer-service/internal/models"
	"time"

	"github.com/google/uuid"
)

type PullRequestsRepository interface {
	CreatePullRequest(ctx context.Context, pr models.PullRequest) error
	GetPullRequestByID(ctx context.Context, id string) (*models.PullRequest, error)
	MergePullRequest(ctx context.Context, id string) error
	AssignReviewer(ctx context.Context, prID, reviewerID string, orderIndex int) error
	ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
	GetOpenPRsWithInactiveReviewers(ctx context.Context, inactiveUserIDs []uuid.UUID) ([]models.PullRequest, error)
	BulkReassignReviewers(ctx context.Context, reassignments []models.ReviewerReassignment) error
}

type UsersRepository interface {
	GetUserByID(ctx context.Context, id string) (*models.User, error)
}

type Service struct {
	prRepo    PullRequestsRepository
	usersRepo UsersRepository
}

func New(prRepo PullRequestsRepository, usersRepo UsersRepository) *Service {
	return &Service{
		prRepo:    prRepo,
		usersRepo: usersRepo,
	}
}

func (s *Service) CreatePR(ctx context.Context, authorID, name string) (string, error) {
	pr := models.PullRequest{
		Name:      name,
		AuthorID:  authorID,
		Status:    "OPEN",
		CreatedAt: time.Now(),
	}

	if err := s.prRepo.CreatePullRequest(ctx, pr); err != nil {
		return "", err
	}

	return pr.ID, nil
}

func (s *Service) AssignReviewers(ctx context.Context, prID string, reviewers []string) error {
	if len(reviewers) == 0 || len(reviewers) > 2 {
		return errors.New("reviewers count must be 1 or 2")
	}

	for i, r := range reviewers {
		user, err := s.usersRepo.GetUserByID(ctx, r)
		if err != nil {
			return err
		}
		if !user.IsActive {
			return errors.New("reviewer is not active")
		}

		if err := s.prRepo.AssignReviewer(ctx, prID, r, i+1); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) Merge(ctx context.Context, prID string) error {
	return s.prRepo.MergePullRequest(ctx, prID)
}

func (s *Service) ReassignReviewer(ctx context.Context, prID, oldID, newID string) error {
	newUser, err := s.usersRepo.GetUserByID(ctx, newID)
	if err != nil {
		return err
	}
	if newUser == nil {
		return errors.New("new reviewer not found")
	}
	if !newUser.IsActive {
		return errors.New("new reviewer not active")
	}

	return s.prRepo.ReassignReviewer(ctx, prID, oldID, newID)
}

func (s *Service) CreatePullRequest(ctx context.Context, pr models.PullRequest) error {
	return s.prRepo.CreatePullRequest(ctx, pr)
}

func (s *Service) GetPullRequest(ctx context.Context, prID string) (*models.PullRequest, error) {
	return s.prRepo.GetPullRequestByID(ctx, prID)
}

func (s *Service) GetOpenPRsWithInactiveReviewers(ctx context.Context, inactiveUserIDs []string) ([]models.PullRequest, error) {
	uuids := make([]uuid.UUID, 0, len(inactiveUserIDs))
	for _, id := range inactiveUserIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("invalid user_id %s: %w", id, err)
		}
		uuids = append(uuids, uid)
	}
	return s.prRepo.GetOpenPRsWithInactiveReviewers(ctx, uuids)
}

func (s *Service) BulkReassignReviewers(ctx context.Context, reassignments []models.ReviewerReassignment) error {
	return s.prRepo.BulkReassignReviewers(ctx, reassignments)
}
