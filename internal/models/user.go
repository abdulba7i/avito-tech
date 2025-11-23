package models

type User struct {
	ID       string `db:"user_id"`
	Username string `db:"username"`
	TeamName string `db:"team_name"`
	IsActive bool   `db:"is_active"`
}
