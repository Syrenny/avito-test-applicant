package domain

import "github.com/google/uuid"

// Team defines model for Team.
type Team struct {
	Members  []TeamMember `json:"members"`
	TeamName string       `json:"team_name"`
}

// TeamMember defines model for TeamMember.
type TeamMember struct {
	IsActive bool   `json:"is_active"`
	UserId   uuid.UUID `json:"user_id"`
	Username string `json:"username"`
}
