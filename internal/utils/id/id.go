package id

import "github.com/google/uuid"

// NewUUID returns a new UUIDv4.
func NewUUID() uuid.UUID {
	return uuid.New()
}
