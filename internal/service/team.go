package service

import (
	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/repo"
	"avito-test-applicant/internal/repo/repoerrors"
	"context"
	"errors"
)

type TeamService struct {
	teamRepo repo.Team
}

func NewTeamService(repos *repo.Repositories) *TeamService {
	return &TeamService{
		teamRepo: repos.Team,
	}
}

func (s *TeamService) CreateTeam(
	ctx context.Context, teamName string,
) (domain.Team, error) {
	tm, err := s.teamRepo.CreateTeam(ctx, teamName)
	if err != nil {
		if errors.Is(err, repoerrors.ErrAlreadyExists) {
			return domain.Team{}, ErrTeamAlreadyExists
		}
		return domain.Team{}, err
	}
	return tm, nil
}

func (s *TeamService) GetTeamByName(
	ctx context.Context, teamName string,
) (domain.Team, error) {
	tm, err := s.teamRepo.GetTeamByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			return domain.Team{}, ErrNotFound
		}
		return domain.Team{}, err
	}
	return tm, nil
}
