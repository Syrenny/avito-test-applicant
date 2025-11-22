package service

import (
	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/repo"
	"context"
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
	ctx context.Context, team_name string,
) (domain.Team, error) {
}

func (s *TeamService) GetTeamByName(
	ctx context.Context, team_name string,
) (domain.Team, error) {
}
