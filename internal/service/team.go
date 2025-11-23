package service

import (
	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/repo"
	"avito-test-applicant/internal/repo/repoerrors"
	"avito-test-applicant/internal/utils/id"
	"avito-test-applicant/pkg/postgres"
	"context"
	"errors"
)

type TeamService struct {
	teamRepo  repo.Team
	userRepo  repo.User
	trManager postgres.TransactionManager
}

func NewTeamService(repos *repo.Repositories, trManager *postgres.TransactionManager) *TeamService {
	return &TeamService{
		teamRepo:  repos.Team,
		userRepo:  repos.User,
		trManager: *trManager,
	}
}

func (s *TeamService) CreateTeamWithUsers(
	ctx context.Context,
	teamName string,
	members []domain.UserInput,
) (domain.TeamWithUsers, error) {
	var result domain.TeamWithUsers

	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		team, err := s.teamRepo.GetTeamByName(ctx, teamName)
		if err != nil && !errors.Is(err, repoerrors.ErrNotFound) {
			return err
		}

		if err == nil {
			return ErrTeamAlreadyExists
		}

		teamId := id.NewUUID()
		team, err = s.teamRepo.CreateTeam(ctx, teamId, teamName)
		if err != nil {
			if errors.Is(err, repoerrors.ErrAlreadyExists) {
				return ErrTeamAlreadyExists
			}
			return err
		}
		var createdOrUpdatedUsers []domain.User

		for _, userInput := range members {
			user, err := s.userRepo.CreateUser(ctx, userInput.UserId, userInput.Username, userInput.IsActive, team.TeamId)
			if err != nil {
				if errors.Is(err, repoerrors.ErrAlreadyExists) {
					user = domain.User{
						UserId:   userInput.UserId,
						TeamId:   team.TeamId,
						IsActive: userInput.IsActive,
						Username: userInput.Username,
					}
					user, err = s.userRepo.UpdateUser(ctx, user)
					if err != nil {
						return err
					}
				} else {
					return err
				}
			}
			createdOrUpdatedUsers = append(createdOrUpdatedUsers, user)
		}

		result.Team = team
		result.Users = createdOrUpdatedUsers
		return nil
	})

	if err != nil {
		return domain.TeamWithUsers{}, err
	}

	return result, nil
}

func (s *TeamService) GetTeamByName(
	ctx context.Context, teamName string,
) (domain.TeamWithUsers, error) {
	team, err := s.teamRepo.GetTeamByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			return domain.TeamWithUsers{}, ErrNotFound
		}
		return domain.TeamWithUsers{}, err
	}

	users, err := s.userRepo.GetUsersByTeam(ctx, team.TeamId)
	if err != nil {
		return domain.TeamWithUsers{}, err
	}

	return domain.TeamWithUsers{
		Team:  team,
		Users: users,
	}, nil
}
