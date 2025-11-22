package domain

import "github.com/google/uuid"

type User struct {
	IsActive bool   `json:"is_active"`
	TeamId uuid.UUID `json:"team_name"`
	UserId   uuid.UUID `json:"user_id"`
	Username string `json:"username"`
}
