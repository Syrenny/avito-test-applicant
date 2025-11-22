package handlers

import (
	apigen "avito-test-applicant/internal/api/gen"
	"context"
)

func (s *Server) PostPullRequestCreate(
	ctx context.Context,
	request apigen.PostPullRequestCreateRequestObject,
) (apigen.PostPullRequestCreateResponseObject, error) {

}

func (s *Server) PostPullRequestMerge(
	ctx context.Context,
	request apigen.PostPullRequestMergeRequestObject,
) (apigen.PostPullRequestMergeResponseObject, error) {

}

func (s *Server) PostPullRequestReassign(
	ctx context.Context,
	request apigen.PostPullRequestReassignRequestObject,
) (apigen.PostPullRequestReassignResponseObject, error) {

}

func (s *Server) GetUsersGetReview(
	ctx context.Context,
	request apigen.GetUsersGetReviewRequestObject,
) (apigen.GetUsersGetReviewResponseObject, error) {

}
