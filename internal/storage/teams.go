package storage

import (
	"context"
	"database/sql"
	"fmt"
	"reviewer-service/internal/models"

	"github.com/google/uuid"
)

type TeamsRepo struct {
	db *sql.DB
}

func NewTeamsRepo(db *sql.DB) *TeamsRepo {
	return &TeamsRepo{db: db}
}

func (r *TeamsRepo) CreateTeam(ctx context.Context, team models.Team) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)", team.Name).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check team exists: %w", err)
	}
	if exists {
		return fmt.Errorf("team_name already exists")
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO teams (team_name) VALUES ($1)", team.Name)
	if err != nil {
		return fmt.Errorf("insert team: %w", err)
	}

	for _, u := range team.Members {
		userID, err := uuid.Parse(u.ID)
		if err != nil {
			return fmt.Errorf("invalid user_id %s: %w", u.ID, err)
		}
		_, err = tx.ExecContext(ctx, `
            INSERT INTO users (user_id, username, team_name, is_active)
            VALUES ($1, $2, $3, $4)
            ON CONFLICT (user_id) DO UPDATE SET username = EXCLUDED.username, team_name = EXCLUDED.team_name, is_active = EXCLUDED.is_active
        `, userID, u.Username, team.Name, u.IsActive)
		if err != nil {
			return fmt.Errorf("insert user %s: %w", u.ID, err)
		}
	}

	return tx.Commit()
}

func (r *TeamsRepo) GetTeamByName(ctx context.Context, name string) (*models.Team, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT user_id, username, is_active
        FROM users
        WHERE team_name = $1
    `, name)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	var members []models.User
	for rows.Next() {
		var u models.User
		var userID uuid.UUID
		if err := rows.Scan(&userID, &u.Username, &u.IsActive); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		u.ID = userID.String()
		u.TeamName = name
		members = append(members, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &models.Team{
		Name:    name,
		Members: members,
	}, nil
}
