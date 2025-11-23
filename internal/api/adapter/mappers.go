package adapter

import (
	"avito-test-applicant/internal/api/adapter/apperrors"
	apigen "avito-test-applicant/internal/api/gen"
	"avito-test-applicant/internal/domain"

	"github.com/google/uuid"
)

// Domain → API

func ParseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, apperrors.ErrInvalidUUID
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

func MapAPIMemberToDomainUserInput(m apigen.TeamMember) (domain.UserInput, error) {
	userId, err := ParseUUID(m.UserId)
	if err != nil {
		return domain.UserInput{}, err
	}
	return domain.UserInput{
		UserId:   userId,
		Username: m.Username,
		IsActive: m.IsActive,
	}, nil
}

func MapAPIMembersToDomainUsersInput(members []apigen.TeamMember) ([]domain.UserInput, error) {
	users := make([]domain.UserInput, len(members))
	for i, m := range members {
		var err error
		users[i], err = MapAPIMemberToDomainUserInput(m)
		if err != nil {
			return nil, err
		}
	}
	return users, nil
}

func MapPullRequestShortToAPI(pr domain.PullRequestShort) apigen.PullRequestShort {
	return apigen.PullRequestShort{
		PullRequestId:   pr.PullRequestId.String(),
		AuthorId:        pr.AuthorId.String(),
		PullRequestName: pr.PullRequestName,
		Status:          apigen.PullRequestShortStatus(pr.Status),
	}
}

func MapPullRequestWithReviewersToAPI(pr domain.PullRequestWithReviewers) apigen.PullRequest {
	reviewers := make([]string, len(pr.Reviewers))
	for i, id := range pr.Reviewers {
		reviewers[i] = id.String()
	}

	return apigen.PullRequest{
		PullRequestId:     pr.PullRequestId.String(),
		PullRequestName:   pr.PullRequestName,
		AuthorId:          pr.AuthorId.String(),
		Status:            apigen.PullRequestStatus(pr.Status),
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
		AssignedReviewers: reviewers,
	}
}
