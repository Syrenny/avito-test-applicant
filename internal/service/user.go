package service

import (
	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/repo"
	"context"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo repo.User
}

func NewUserService(repos *repo.Repositories) *UserService {
	return &UserService{
		userRepo: repos.User,
	}
}

func (s *UserService) CreateUser(
	ctx context.Context, username string, team_name string,
) (domain.User, error) {
}

func (s *UserService) GetUserById(
	ctx context.Context, user_id uuid.UUID,
) (domain.User, error) {
}

func (s *UserService) SetIsActive(
	ctx context.Context, user_id uuid.UUID, is_active bool,
) (domain.User, error) {
}
