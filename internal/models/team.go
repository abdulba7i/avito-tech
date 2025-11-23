package models

import "github.com/google/uuid"

type Team struct {
	Name    string `db:"team_name"`
	Members []User `db:"-"`
}

type MemberData struct {
	UserID   uuid.UUID
	Username string
	IsActive bool
}
