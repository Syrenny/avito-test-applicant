package handlers

import (
	"avito-test-applicant/internal/api/adapter"
	apigen "avito-test-applicant/internal/api/gen"
	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/service"
	"context"
	"errors"

	"github.com/google/uuid"
)

func (s *Server) PostPullRequestCreate(
	ctx context.Context,
	request apigen.PostPullRequestCreateRequestObject,
) (apigen.PostPullRequestCreateResponseObject, error) {
	if request.Body == nil {
		return nil, errors.New("empty body")
	}

	prID, err := adapter.ParseUUID(request.Body.PullRequestId)
	if err != nil {
		return nil, err
	}

	authorID, err := adapter.ParseUUID(request.Body.AuthorId)
	if err != nil {
		return nil, err
	}

	result, err := s.Services.PullRequest.CreateAndAssignPullRequest(
		ctx,
		prID,
		request.Body.PullRequestName,
		authorID,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAuthorNotFound):
			return apigen.PostPullRequestCreate404JSONResponse(makeAPIError(apigen.NOTFOUND, err.Error())), nil
		case errors.Is(err, service.ErrPullRequestExists):
			return apigen.PostPullRequestCreate409JSONResponse(makeAPIError(apigen.PREXISTS, err.Error())), nil
		default:
			return nil, err
		}
	}

	apiPullRequest := adapter.MapPullRequestWithReviewersToAPI(result)
	resp := apigen.PostPullRequestCreate201JSONResponse{
		Pr: &apiPullRequest,
	}

	return resp, nil
}

func (s *Server) PostPullRequestMerge(
	ctx context.Context,
	request apigen.PostPullRequestMergeRequestObject,
) (apigen.PostPullRequestMergeResponseObject, error) {
	if request.Body == nil {
		return nil, errors.New("empty body")
	}

	prID, err := adapter.ParseUUID(request.Body.PullRequestId)
	if err != nil {
		return nil, err
	}

	pr, err := s.Services.PullRequest.SetMerged(ctx, prID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPullRequestNotFound):
			return apigen.PostPullRequestMerge404JSONResponse(makeAPIError(apigen.NOTFOUND, err.Error())), nil
		default:
			return nil, err
		}
	}
	apiPullRequest := adapter.MapPullRequestWithReviewersToAPI(domain.PullRequestWithReviewers{
		PullRequest: pr,
		Reviewers:   []uuid.UUID{},
	})
	resp := apigen.PostPullRequestMerge200JSONResponse{
		Pr: &apiPullRequest,
	}

	return resp, nil
}

func (s *Server) PostPullRequestReassign(
	ctx context.Context,
	request apigen.PostPullRequestReassignRequestObject,
) (apigen.PostPullRequestReassignResponseObject, error) {
	if request.Body == nil {
		return nil, errors.New("empty body")
	}

	prID, err := adapter.ParseUUID(request.Body.PullRequestId)
	if err != nil {
		return nil, err
	}

	oldID, err := adapter.ParseUUID(request.Body.OldUserId)
	if err != nil {
		return nil, err
	}

	result, err := s.Services.PullRequest.Reassign(ctx, prID, oldID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPullRequestNotFound):
			return apigen.PostPullRequestReassign404JSONResponse(makeAPIError(apigen.NOTFOUND, err.Error())), nil
		case errors.Is(err, service.ErrPullRequestMerged):
			return apigen.PostPullRequestReassign409JSONResponse(makeAPIError(apigen.PRMERGED, err.Error())), nil
		case errors.Is(err, service.ErrUserNotFound):
			return apigen.PostPullRequestReassign404JSONResponse(makeAPIError(apigen.NOTFOUND, err.Error())), nil
		case errors.Is(err, service.ErrNoCandidate):
			return apigen.PostPullRequestReassign409JSONResponse(makeAPIError(apigen.NOCANDIDATE, err.Error())), nil
		default:
			return nil, err
		}
	}

	resp := apigen.PostPullRequestReassign200JSONResponse{
		Pr: adapter.MapPullRequestWithReviewersToAPI(result),
	}

	return resp, nil
}

func (s *Server) GetUsersGetReview(
	ctx context.Context,
	request apigen.GetUsersGetReviewRequestObject,
) (apigen.GetUsersGetReviewResponseObject, error) {
	userID, err := adapter.ParseUUID(request.Params.UserId)
	if err != nil {
		return nil, err
	}

	prs, err := s.Services.PullRequest.GetAssignedReviewsByUserId(ctx, userID)
	if err != nil {
		return nil, errors.New("internal error")
	}

	out := make([]apigen.PullRequestShort, len(prs))
	for i, pr := range prs {
		out[i] = adapter.MapPullRequestShortToAPI(pr)
	}

	resp := apigen.GetUsersGetReview200JSONResponse{
		UserId:       userID.String(),
		PullRequests: out,
	}

	return resp, nil
}
