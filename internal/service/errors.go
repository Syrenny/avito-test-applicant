package service

import "errors"

var (
	ErrNotFound          = errors.New("entity not found")
	ErrTeamAlreadyExists = errors.New("team already exists")

	PrExistsError           = errors.New("pull request already exists")
	ErrPullRequestNotFound  = errors.New("pull request not found")
	ErrPullRequestMerged    = errors.New("cannot reassign reviewers on merged pull request")
	ErrReviewerUserNotFound = errors.New("user to replace not found")
	ErrNoActiveUsersToReassign    = errors.New("no active users available to reassign")
)
