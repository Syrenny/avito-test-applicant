package adapter

import (
	"fmt"

	apigen "avito-test-applicant/internal/api/gen"
	"avito-test-applicant/internal/domain"

	"github.com/google/uuid"
)

// Domain → API

func ParseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid %q: %w", s, err)
	}
	return id, nil
}

func MapDomainUserToAPIMember(u domain.User) apigen.TeamMember {
	return apigen.TeamMember{
		UserId:   u.UserId.String(),
		Username: u.Username,
		IsActive: u.IsActive,
	}
}

func MapDomainUsersToAPIMembers(users []domain.User) []apigen.TeamMember {
	members := make([]apigen.TeamMember, len(users))
	for i, u := range users {
		members[i] = MapDomainUserToAPIMember(u)
	}
	return members
}

func MapDomainTeamWithUsersToAPITeam(t domain.TeamWithUsers) *apigen.Team {
	return &apigen.Team{
		TeamName: t.Team.TeamName,
		Members:  MapDomainUsersToAPIMembers(t.Users),
	}
}

// API → Domain

func MapAPIMemberToDomainUserInput(m apigen.TeamMember) domain.UserInput {
	userId, _ := uuid.Parse(m.UserId)
	return domain.UserInput{
		UserId:   userId,
		Username: m.Username,
		IsActive: m.IsActive,
	}
}

func MapAPIMembersToDomainUsersInput(members []apigen.TeamMember) []domain.UserInput {
	users := make([]domain.UserInput, len(members))
	for i, m := range members {
		users[i] = MapAPIMemberToDomainUserInput(m)
	}
	return users
}
