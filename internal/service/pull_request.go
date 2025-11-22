package service

import (
	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/repo"
	"avito-test-applicant/internal/repo/repoerrors"
	"avito-test-applicant/pkg/postgres"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type PullRequestService struct {
	pullRequestRepo repo.PullRequest
	reviewerRepo    repo.Reviewer
	userRepo        repo.User
	trManager       postgres.TransactionManager
}

func NewPullRequestService(repos *repo.Repositories, trManager *postgres.TransactionManager) *PullRequestService {
	return &PullRequestService{
		pullRequestRepo: repos.PullRequest,
		reviewerRepo:    repos.Reviewer,
		userRepo:        repos.User,
		trManager:       *trManager,
	}
}

func (s *PullRequestService) CreatePullRequest(
	ctx context.Context, pullRequestId uuid.UUID, pullRequestName string, authorId uuid.UUID,
) (domain.PullRequest, error) {
	pr, err := s.pullRequestRepo.CreatePullRequest(ctx, pullRequestId, pullRequestName, authorId)
	if err != nil {
		if errors.Is(err, repoerrors.ErrAlreadyExists) {
			return domain.PullRequest{}, PrExistsError
		}
		return domain.PullRequest{}, err
	}

	return pr, nil
}

func (s *PullRequestService) GetPullRequestById(
	ctx context.Context, pullRequestId uuid.UUID,
) (domain.PullRequest, error) {
	pr, err := s.pullRequestRepo.GetPullRequestById(ctx, pullRequestId)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			return domain.PullRequest{}, ErrNotFound
		}
		return domain.PullRequest{}, err
	}

	return pr, nil
}

func (s *PullRequestService) SetMerged(
	ctx context.Context, pullRequestId uuid.UUID,
) (domain.PullRequest, error) {
	pr, err := s.pullRequestRepo.SetMerged(ctx, pullRequestId)
	if err != nil {
		if errors.Is(err, repoerrors.ErrNotFound) {
			return domain.PullRequest{}, ErrNotFound
		}
		return domain.PullRequest{}, err
	}

	return pr, nil
}

func (s *PullRequestService) Reassign(
	ctx context.Context,
	pullRequestId uuid.UUID,
	oldUserId uuid.UUID,
) (domain.PullRequest, error) {
	var pr domain.PullRequest

	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		p, err := s.pullRequestRepo.GetPullRequestById(ctx, pullRequestId)
		if err != nil {
			if errors.Is(err, repoerrors.ErrNotFound) {
				return ErrPullRequestNotFound
			}
			return err
		}

		if p.Status == domain.PullRequestStatusMERGED {
			return ErrPullRequestMerged
		}

		oldUser, err := s.userRepo.GetUserById(ctx, oldUserId)
		if err != nil {
			if errors.Is(err, repoerrors.ErrNotFound) {
				return ErrReviewerUserNotFound
			}
			return err
		}

		teamMembers, err := s.userRepo.GetUsersByTeam(ctx, oldUser.TeamId)
		if err != nil {
			return err
		}

		var candidates []uuid.UUID
		for _, u := range teamMembers {
			if u.IsActive && u.UserId != oldUserId {
				candidates = append(candidates, u.UserId)
			}
		}

		if len(candidates) == 0 {
			return ErrNoActiveUsersToReassign
		}

		newUserId := candidates[randomIndex(len(candidates))]

		if err := s.reviewerRepo.RemoveOne(ctx, pullRequestId, oldUserId); err != nil {
			return err
		}

		if err := s.reviewerRepo.AssignOne(ctx, pullRequestId, newUserId); err != nil {
			return err
		}

		pr = p
		return nil
	})

	return pr, err
}

func randomIndex(n int) int {
	return int(time.Now().UnixNano() % int64(n))
}
