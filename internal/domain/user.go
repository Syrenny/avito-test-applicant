package domain

import "github.com/google/uuid"

type User struct {
	IsActive bool      `json:"is_active"`
	TeamId   uuid.UUID `json:"team_name"`
	UserId   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
}

type UserInput struct {
	IsActive bool      `json:"is_active"`
	UserId   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
}

type UserWithTeamName struct {
	IsActive bool      `json:"is_active"`
	TeamName string    `json:"team_name"`
	UserId   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
}
