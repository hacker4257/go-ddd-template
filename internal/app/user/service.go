package userapp

import (
	"context"
	"strings"

	"github.com/hacker4257/go-ddd-template/internal/domain/user"
)

type Service struct {
	repo user.Repo
}

func New(repo user.Repo) *Service {
	return &Service{repo: repo}
}

type CreateUserCmd struct {
	Name  string
	Email string
}

func (s *Service) Create(ctx context.Context, cmd CreateUserCmd) (user.User, error) {
	name := strings.TrimSpace(cmd.Name)
	email := strings.TrimSpace(strings.ToLower(cmd.Email))
	if name == "" || email == "" {
		return user.User{}, user.ErrInvalidInput
	}

	// 先查一下（DB 唯一键也会兜底）
	if _, err := s.repo.GetByEmail(ctx, email); err == nil {
		return user.User{}, user.ErrEmailExists
	} else if err != nil && err != user.ErrNotFound {
		return user.User{}, err
	}

	return s.repo.Create(ctx, name, email)
}

func (s *Service) Get(ctx context.Context, id uint64) (user.User, error) {
	return s.repo.GetByID(ctx, id)
}
