package storage

import (
	"context"
	"database/sql"
	"fmt"
)

type StatisticsRepo struct {
	db *sql.DB
}

func NewStatisticsRepo(db *sql.DB) *StatisticsRepo {
	return &StatisticsRepo{db: db}
}

type UserAssignmentStats struct {
	UserID   string
	Username string
	Count    int
}

type PRStats struct {
	TotalPRs      int
	OpenPRs       int
	MergedPRs     int
	TotalAssignments int
}

func (r *StatisticsRepo) GetUserAssignmentStats(ctx context.Context) ([]UserAssignmentStats, error) {
	query := `
		SELECT 
			u.user_id::text,
			u.username,
			COUNT(pr.reviewer_id) as assignment_count
		FROM users u
		LEFT JOIN pr_reviewers pr ON u.user_id = pr.reviewer_id
		GROUP BY u.user_id, u.username
		ORDER BY assignment_count DESC, u.username
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query user stats: %w", err)
	}
	defer rows.Close()

	var stats []UserAssignmentStats
	for rows.Next() {
		var stat UserAssignmentStats
		if err := rows.Scan(&stat.UserID, &stat.Username, &stat.Count); err != nil {
			return nil, fmt.Errorf("scan user stat: %w", err)
		}
		stats = append(stats, stat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return stats, nil
}

func (r *StatisticsRepo) GetPRStats(ctx context.Context) (*PRStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_prs,
			COUNT(*) FILTER (WHERE status = 'OPEN') as open_prs,
			COUNT(*) FILTER (WHERE status = 'MERGED') as merged_prs,
			(SELECT COUNT(*) FROM pr_reviewers) as total_assignments
		FROM pull_requests
	`

	var stats PRStats
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalPRs,
		&stats.OpenPRs,
		&stats.MergedPRs,
		&stats.TotalAssignments,
	)
	if err != nil {
		return nil, fmt.Errorf("query PR stats: %w", err)
	}

	return &stats, nil
}

