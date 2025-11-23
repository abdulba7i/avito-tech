package models

import "time"

type PullRequestStatus string

const (
	PullRequestStatusOpen   PullRequestStatus = "OPEN"
	PullRequestStatusMerged PullRequestStatus = "MERGED"
)

type PullRequest struct {
	ID        string            `db:"pull_request_id"`
	Name      string            `db:"pull_request_name"`
	AuthorID  string            `db:"author_id"`
	Status    PullRequestStatus `db:"status"`
	CreatedAt time.Time         `db:"created_at"`
	MergedAt  *time.Time        `db:"merged_at"`
	Reviewers []string          `db:"-"`
}
