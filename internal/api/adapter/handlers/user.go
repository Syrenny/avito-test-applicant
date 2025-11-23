package handlers

import (
	"avito-test-applicant/internal/api/adapter"
	apigen "avito-test-applicant/internal/api/gen"
	"avito-test-applicant/internal/service"
	"context"
	"errors"
)

func (s *Server) PostUsersSetIsActive(
	ctx context.Context,
	request apigen.PostUsersSetIsActiveRequestObject,
) (apigen.PostUsersSetIsActiveResponseObject, error) {
	if request.Body == nil {
		return nil, errors.New("request body is empty")
	}

	userId, err := adapter.ParseUUID(request.Body.UserId)
	if err != nil {
		return nil, err
	}

	updatedUser, err := s.Services.User.SetIsActive(ctx, userId, request.Body.IsActive)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return apigen.PostUsersSetIsActive404JSONResponse(
				makeAPIError(apigen.NOTFOUND, "user not found"),
			), nil
		}
		return nil, err
	}

	response := apigen.PostUsersSetIsActive200JSONResponse{
		User: &apigen.User{
			UserId:   updatedUser.UserId.String(),
			Username: updatedUser.Username,
			TeamName: updatedUser.TeamName,
			IsActive: updatedUser.IsActive,
		},
	}

	return response, nil
}
