package service

import (
	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/repo"
	"avito-test-applicant/internal/repo/repoerrors"
	"avito-test-applicant/pkg/postgres"
	"context"
	"errors"
	"math/rand"
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

func (s *PullRequestService) selectFromTeamExcludeAuthor(
	ctx context.Context,
	teamId uuid.UUID,
	authorId uuid.UUID,
	n int,
) ([]uuid.UUID, error) {
	users, err := s.userRepo.GetUsersByTeam(ctx, teamId)
	if err != nil {
		return nil, err
	}

	// 1) collect active non-author users
	candidates := make([]uuid.UUID, 0, len(users))
	for _, u := range users {
		if u.UserId == authorId {
			continue
		}
		if !u.IsActive {
			continue
		}
		candidates = append(candidates, u.UserId)
	}
	if len(candidates) == 0 {
		return []uuid.UUID{}, nil
	}

	// 2) shuffle
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	// 3) take first n (or fewer)
	if len(candidates) > n {
		candidates = candidates[:n]
	}
	return candidates, nil
}

func (s *PullRequestService) selectReplacement(
	ctx context.Context,
	teamId uuid.UUID,
	authorId uuid.UUID,
	assigned []uuid.UUID,
	oldUserId uuid.UUID,
) (uuid.UUID, error) {
	users, err := s.userRepo.GetUsersByTeam(ctx, teamId)
	if err != nil {
		return uuid.Nil, err
	}

	assignedSet := make(map[uuid.UUID]struct{}, len(assigned))
	for _, id := range assigned {
		assignedSet[id] = struct{}{}
	}

	candidates := make([]uuid.UUID, 0, len(users))
	for _, u := range users {
		if u.UserId == authorId {
			continue
		}
		if u.UserId == oldUserId {
			continue
		}
		if !u.IsActive {
			continue
		}
		if _, exists := assignedSet[u.UserId]; exists {
			continue
		}
		candidates = append(candidates, u.UserId)
	}

	if len(candidates) == 0 {
		return uuid.Nil, ErrNoCandidate
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return candidates[r.Intn(len(candidates))], nil
}

func (s *PullRequestService) assignReviewers(
	ctx context.Context,
	prID uuid.UUID,
	reviewers []uuid.UUID,
) error {
	for _, rid := range reviewers {
		if err := s.reviewerRepo.AssignOne(ctx, prID, rid); err != nil {
			return err
		}
	}
	return nil
}

func (s *PullRequestService) CreateAndAssignPullRequest(
	ctx context.Context, pullRequestId uuid.UUID, pullRequestName string, authorId uuid.UUID,
) (domain.PullRequestWithReviewers, error) {
	var result domain.PullRequestWithReviewers

	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		// 1) ensure author exists
		author, err := s.userRepo.GetUserById(ctx, authorId)
		if err != nil {
			// propagate NotFound as-is
			if errors.Is(err, repoerrors.ErrNotFound) {
				return ErrAuthorNotFound
			}
			return err
		}

		// 2) create PR
		pr, err := s.pullRequestRepo.CreatePullRequest(ctx, pullRequestId, pullRequestName, authorId)
		if err != nil {
			// PR exists -> bubble up as already exists
			if errors.Is(err, repoerrors.ErrAlreadyExists) {
				return ErrPullRequestExists
			}
			return err
		}
		// 3) select up to 2 reviewers
		reviewers, err := s.selectFromTeamExcludeAuthor(ctx, author.TeamId, authorId, 2)
		if err != nil {
			return err
		}

		// 4) assign reviewers
		if len(reviewers) > 0 {
			if err := s.assignReviewers(ctx, pr.PullRequestId, reviewers); err != nil {
				return err
			}
		}

		// 5) prepare result
		result.PullRequest = pr
		result.Reviewers = reviewers
		return nil
	})

	if err != nil {
		return domain.PullRequestWithReviewers{}, err
	}

	return result, nil
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
	var pr domain.PullRequest

	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		// 1) get current PR
		current, err := s.pullRequestRepo.GetPullRequestById(ctx, pullRequestId)
		if err != nil {
			if errors.Is(err, repoerrors.ErrNotFound) {
				return ErrPullRequestNotFound
			}
			return err
		}

		// 2) if already merged — idempotent, return current state
		if current.Status == domain.PullRequestStatusMERGED {
			pr = current
			return nil
		}

		// 3) otherwise set merged and return updated row
		updated, err := s.pullRequestRepo.SetMerged(ctx, pullRequestId)
		if err != nil {
			if errors.Is(err, repoerrors.ErrNotFound) {
				return ErrNotFound
			}
			return err
		}

		pr = updated
		return nil
	})

	if err != nil {
		return domain.PullRequest{}, err
	}
	return pr, nil
}

func (s *PullRequestService) Reassign(
	ctx context.Context,
	pullRequestId uuid.UUID,
	oldUserId uuid.UUID,
) (domain.PullRequestWithReviewers, error) {
	var result domain.PullRequestWithReviewers

	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		// 1) получить PR
		pr, err := s.pullRequestRepo.GetPullRequestById(ctx, pullRequestId)
		if err != nil {
			if errors.Is(err, repoerrors.ErrNotFound) {
				return ErrPullRequestNotFound
			}
			return err
		}

		// 2) проверить статус PR (MERGED нельзя менять)
		if pr.Status == domain.PullRequestStatusMERGED {
			return ErrPullRequestMerged
		}

		// 3) проверить, что oldUserId действительно назначен
		assignedReviewers, err := s.reviewerRepo.ListReviewers(ctx, pullRequestId)
		if err != nil {
			return err
		}
		found := false
		for _, uid := range assignedReviewers {
			if uid == oldUserId {
				found = true
				break
			}
		}
		if !found {
			return ErrUserNotFound
		}

		oldUser, err := s.userRepo.GetUserById(ctx, pr.AuthorId)
		if err != nil {
			return err
		}

		// 4) выбрать кандидата на замену из команды автора
		replacement, err := s.selectReplacement(ctx, oldUser.TeamId, pr.AuthorId, assignedReviewers, oldUserId)
		if err != nil {
			return err
		}

		// 5) снять oldUserId
		if err := s.reviewerRepo.RemoveOne(ctx, pullRequestId, oldUserId); err != nil {
			return err
		}

		// 6) назначить нового ревьювера
		if err := s.reviewerRepo.AssignOne(ctx, pullRequestId, replacement); err != nil {
			return err
		}

		// 7) собрать результат
		updatedReviewers, err := s.reviewerRepo.ListReviewers(ctx, pullRequestId)
		if err != nil {
			return err
		}

		result.PullRequest = pr
		result.Reviewers = updatedReviewers
		return nil
	})

	if err != nil {
		return domain.PullRequestWithReviewers{}, err
	}

	return result, nil
}

func (s *PullRequestService) GetAssignedReviewsByUserId(
	ctx context.Context,
	userId uuid.UUID,
) ([]domain.PullRequestShort, error) {
	// 1) получить список PR, где user назначен
	prIDs, err := s.reviewerRepo.ListByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	if len(prIDs) == 0 {
		return []domain.PullRequestShort{}, nil
	}

	// 2) получить краткую информацию о PR
	prs, err := s.pullRequestRepo.GetPullRequestsByIds(ctx, prIDs)
	if err != nil {
		return nil, err
	}

	// 3) собрать результат
	result := make([]domain.PullRequestShort, len(prs))
	for i, pr := range prs {
		result[i] = domain.PullRequestShort{
			PullRequestId:   pr.PullRequestId,
			PullRequestName: pr.PullRequestName,
			AuthorId:        pr.AuthorId,
			Status:          pr.Status,
		}
	}

	return result, nil
}
