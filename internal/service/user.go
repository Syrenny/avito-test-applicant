package service

import (
	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/repo"
	"avito-test-applicant/internal/repo/repoerrors"
	"avito-test-applicant/pkg/postgres"
	"context"
	"errors"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo  repo.User
	teamRepo  repo.Team
	trManager postgres.TransactionManager
}

func NewUserService(
	repos *repo.Repositories,
	trManager *postgres.TransactionManager,
) *UserService {
	return &UserService{
		userRepo:  repos.User,
		teamRepo:  repos.Team,
		trManager: *trManager,
	}
}

func (s *UserService) GetUserById(
	ctx context.Context, userId uuid.UUID,
) (domain.User, error) {
	user, err := s.userRepo.GetUserById(ctx, userId)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, err
	}
	return user, nil
}

func (s *UserService) SetIsActive(
	ctx context.Context, userId uuid.UUID, isActive bool,
) (domain.UserWithTeamName, error) {
	user, err := s.userRepo.SetIsActive(ctx, userId, isActive)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			return domain.UserWithTeamName{}, ErrNotFound
		}
		return domain.UserWithTeamName{}, err
	}
	team, err := s.teamRepo.GetTeamById(ctx, user.TeamId)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			return domain.UserWithTeamName{}, ErrNotFound
		}
		return domain.UserWithTeamName{}, err
	}

	return domain.UserWithTeamName{
		IsActive: user.IsActive,
		TeamName: team.TeamName,
		UserId:   user.UserId,
		Username: user.Username,
	}, nil
}
