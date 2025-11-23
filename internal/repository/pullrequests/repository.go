package pullrequests

import (
	"context"
	"reviewer-service/internal/models"
)

type PullRequestRepository interface {
	CreatePullRequest(ctx context.Context, pr models.PullRequest) error
	GetPullRequestByID(ctx context.Context, id string) (*models.PullRequest, error)
	MergePullRequest(ctx context.Context, id string) error
	AssignReviewer(ctx context.Context, prID string, reviewerID string, orderIndex int) error
	GetReviewers(ctx context.Context, prID string) ([]string, error)
	ReassignReviewer(ctx context.Context, prID string, oldReviewerID string, newReviewerID string) error
	GetPRsByReviewer(ctx context.Context, userID string) ([]models.PullRequest, error)
}
