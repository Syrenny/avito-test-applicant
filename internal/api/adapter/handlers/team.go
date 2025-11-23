package handlers

import (
	"avito-test-applicant/internal/api/adapter"
	apigen "avito-test-applicant/internal/api/gen"
	"avito-test-applicant/internal/service"
	"context"
	"errors"
)

func (s *Server) PostTeamAdd(
	ctx context.Context,
	request apigen.PostTeamAddRequestObject,
) (apigen.PostTeamAddResponseObject, error) {
	if request.Body == nil {
		return nil, errors.New("request body is empty")
	}

	domainUsers := adapter.MapAPIMembersToDomainUsersInput(request.Body.Members)

	teamWithUsers, err := s.Services.Team.CreateTeamWithUsers(ctx, request.Body.TeamName, domainUsers)
	if err != nil {
		if errors.Is(err, service.ErrTeamAlreadyExists) {
			return apigen.PostTeamAdd400JSONResponse(makeAPIError(apigen.TEAMEXISTS, err.Error())), nil
		}
		return nil, err
	}

	response := apigen.PostTeamAdd201JSONResponse{
		Team: adapter.MapDomainTeamWithUsersToAPITeam(teamWithUsers),
	}

	return response, nil
}

func (s *Server) GetTeamGet(
	ctx context.Context,
	request apigen.GetTeamGetRequestObject,
) (apigen.GetTeamGetResponseObject, error) {
	teamName := string(request.Params.TeamName)

	teamWithUsers, err := s.Services.Team.GetTeamByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return apigen.GetTeamGet404JSONResponse(makeAPIError(apigen.NOTFOUND, err.Error())), nil
		}
		return nil, err
	}

	response := apigen.GetTeamGet200JSONResponse{
		TeamName: teamWithUsers.Team.TeamName,
		Members:  adapter.MapDomainUsersToAPIMembers(teamWithUsers.Users),
	}

	return response, nil
}
