package storage

import (
	"context"
	"database/sql"
	"fmt"
	"reviewer-service/internal/models"
	"strings"

	"github.com/google/uuid"
)

type PullRequestsRepo struct {
	db *sql.DB
}

func NewPullRequestsRepo(db *sql.DB) *PullRequestsRepo {
	return &PullRequestsRepo{db: db}
}

func (r *PullRequestsRepo) CreatePullRequest(ctx context.Context, pr models.PullRequest) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `, pr.ID, pr.Name, pr.AuthorID, pr.Status, pr.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("PR id already exists")
		}
	}
	return err
}

func (r *PullRequestsRepo) GetPullRequestByID(ctx context.Context, id string) (*models.PullRequest, error) {
	var pr models.PullRequest
	err := r.db.QueryRowContext(ctx, `
        SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
        FROM pull_requests
        WHERE pull_request_id = $1
    `, id).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pull request not found")
		}
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
        SELECT reviewer_id
        FROM pr_reviewers
        WHERE pull_request_id = $1
        ORDER BY order_index
    `, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var reviewerID uuid.UUID
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}
		pr.Reviewers = append(pr.Reviewers, reviewerID.String())
	}

	return &pr, nil
}

func (r *PullRequestsRepo) MergePullRequest(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE pull_requests
        SET status = 'MERGED', merged_at = COALESCE(merged_at, now())
        WHERE pull_request_id = $1
    `, id)
	return err
}

func (r *PullRequestsRepo) AssignReviewer(ctx context.Context, prID, reviewerID string, orderIndex int) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO pr_reviewers (pull_request_id, reviewer_id, order_index)
        VALUES ($1, $2, $3)
    `, prID, reviewerID, orderIndex)
	return err
}

func (r *PullRequestsRepo) ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	oldUUID, err := uuid.Parse(oldReviewerID)
	if err != nil {
		return fmt.Errorf("invalid old_reviewer_id: %w", err)
	}
	newUUID, err := uuid.Parse(newReviewerID)
	if err != nil {
		return fmt.Errorf("invalid new_reviewer_id: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
        UPDATE pr_reviewers
        SET reviewer_id = $1
        WHERE pull_request_id = $2 AND reviewer_id = $3
    `, newUUID, prID, oldUUID)
	return err
}

func (r *PullRequestsRepo) GetPRsByReviewer(ctx context.Context, userID string) ([]models.PullRequest, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}
	rows, err := r.db.QueryContext(ctx, `
        SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
        FROM pull_requests pr
        JOIN pr_reviewers rr ON pr.pull_request_id = rr.pull_request_id
        WHERE rr.reviewer_id = $1
    `, userUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []models.PullRequest
	for rows.Next() {
		var pr models.PullRequest
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
			return nil, err
		}

		revRows, _ := r.db.QueryContext(ctx, `
            SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id = $1 ORDER BY order_index
        `, pr.ID)
		for revRows.Next() {
			var revID uuid.UUID
			if err := revRows.Scan(&revID); err == nil {
				pr.Reviewers = append(pr.Reviewers, revID.String())
			}
		}
		revRows.Close()

		prs = append(prs, pr)
	}

	return prs, nil
}

func (r *PullRequestsRepo) GetOpenPRsWithInactiveReviewers(ctx context.Context, inactiveUserIDs []uuid.UUID) ([]models.PullRequest, error) {
	if len(inactiveUserIDs) == 0 {
		return []models.PullRequest{}, nil
	}

	query := `
		SELECT DISTINCT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
		FROM pull_requests pr
		JOIN pr_reviewers rev ON pr.pull_request_id = rev.pull_request_id
		WHERE pr.status = 'OPEN'
		  AND rev.reviewer_id = ANY($1::uuid[])
		ORDER BY pr.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, inactiveUserIDs)
	if err != nil {
		return nil, fmt.Errorf("query open PRs: %w", err)
	}
	defer rows.Close()

	var prs []models.PullRequest
	for rows.Next() {
		var pr models.PullRequest
		var mergedAt sql.NullTime
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &mergedAt); err != nil {
			return nil, fmt.Errorf("scan PR: %w", err)
		}
		if mergedAt.Valid {
			pr.MergedAt = &mergedAt.Time
		}

		revRows, _ := r.db.QueryContext(ctx, `
			SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id = $1 ORDER BY order_index
		`, pr.ID)
		for revRows.Next() {
			var revID uuid.UUID
			if err := revRows.Scan(&revID); err == nil {
				pr.Reviewers = append(pr.Reviewers, revID.String())
			}
		}
		revRows.Close()

		prs = append(prs, pr)
	}

	return prs, nil
}

func (r *PullRequestsRepo) BulkReassignReviewers(ctx context.Context, reassignments []models.ReviewerReassignment) error {
	if len(reassignments) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		UPDATE pr_reviewers
		SET reviewer_id = $1
		WHERE pull_request_id = $2 AND reviewer_id = $3
	`)
	if err != nil {
		return fmt.Errorf("prepare stmt: %w", err)
	}
	defer stmt.Close()

	for _, reassignment := range reassignments {
		_, err := stmt.ExecContext(ctx, reassignment.NewReviewerID, reassignment.PRID, reassignment.OldReviewerID)
		if err != nil {
			return fmt.Errorf("reassign reviewer: %w", err)
		}
	}

	return tx.Commit()
}
