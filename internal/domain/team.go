package domain

import "github.com/google/uuid"

type Team struct {
	TeamId   uuid.UUID `json:"team_id"`
	TeamName string    `json:"team_name"`
}

type TeamWithUsers struct {
	Team  Team   `json:"team"`
	Users []User `json:"users,omitempty"`
}
