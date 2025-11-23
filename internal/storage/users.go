package storage

import (
	"context"
	"database/sql"
	"errors"
	"reviewer-service/internal/models"
	"reviewer-service/internal/repository/users"

	"github.com/google/uuid"
)

type UsersRepository struct {
	db *sql.DB
}

func NewUsersRepo(db *sql.DB) users.Repository {
	return &UsersRepository{
		db: db,
	}
}

func (r *UsersRepository) CreateUser(ctx context.Context, user models.User) error {
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return err
	}

	const query = `
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET 
			username = EXCLUDED.username, 
			team_name = EXCLUDED.team_name, 
			is_active = EXCLUDED.is_active
	`

	_, err = r.db.ExecContext(ctx, query, userID, user.Username, user.TeamName, user.IsActive)
	return err
}

func (r *UsersRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	const query = `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE user_id = $1
	`

	u := &models.User{}
	var userID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&userID, &u.Username, &u.TeamName, &u.IsActive)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	u.ID = userID.String()
	return u, nil
}

func (r *UsersRepository) SetIsActive(ctx context.Context, id uuid.UUID, isActive bool) error {
	const query = `
		UPDATE users
		SET is_active = $1
		WHERE user_id = $2
	`

	res, err := r.db.ExecContext(ctx, query, isActive, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *UsersRepository) SetTeamInactive(ctx context.Context, teamName string) error {
	const query = `
		UPDATE users
		SET is_active = false
		WHERE team_name = $1 AND is_active = true
	`

	_, err := r.db.ExecContext(ctx, query, teamName)
	return err
}

func (r *UsersRepository) GetActiveByTeam(ctx context.Context, teamName string) ([]*models.User, error) {
	const query = `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1
		  AND is_active = true
	`

	rows, err := r.db.QueryContext(ctx, query, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.User

	for rows.Next() {
		u := &models.User{}
		var userID uuid.UUID
		if err := rows.Scan(&userID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		u.ID = userID.String()
		result = append(result, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *UsersRepository) GetAssignedPullRequests(ctx context.Context, userID uuid.UUID) ([]*models.PullRequest, error) {
	const query = `
		SELECT pr.pull_request_id,
		       pr.pull_request_name,
		       pr.author_id,
		       pr.status,
		       pr.created_at,
		       pr.merged_at
		FROM pr_reviewers rev
		JOIN pull_requests pr ON pr.pull_request_id = rev.pull_request_id
		WHERE rev.reviewer_id = $1
		ORDER BY pr.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []*models.PullRequest

	for rows.Next() {
		var mergedAt sql.NullTime

		pr := &models.PullRequest{}
		err = rows.Scan(
			&pr.ID,
			&pr.Name,
			&pr.AuthorID,
			&pr.Status,
			&pr.CreatedAt,
			&mergedAt,
		)
		if err != nil {
			return nil, err
		}

		if mergedAt.Valid {
			pr.MergedAt = &mergedAt.Time
		} else {
			pr.MergedAt = nil
		}

		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prs, nil
}
