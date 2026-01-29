package userapp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hacker4257/go-ddd-template/internal/domain/event"
	"github.com/hacker4257/go-ddd-template/internal/domain/user"
	"github.com/hacker4257/go-ddd-template/internal/pkg/trace"
)

type Service struct {
	repo  user.Repo
	cache user.Cache
	ttl   time.Duration
	
	pub       event.Publisher
	userTopic string


}

func New(repo user.Repo, cache user.Cache, ttl time.Duration, pub event.Publisher, userTopic string) *Service {
	return &Service{repo: repo, cache: cache, ttl: ttl, pub: pub, userTopic: userTopic}
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
	
	u, err := s.repo.Create(ctx, name, email)
	if err != nil {
		return user.User{}, err
	}
	
	if s.pub != nil && s.userTopic != "" {
		rid := trace.RequestID(ctx)
		_ = s.pub.Publish(ctx, s.userTopic, event.Event{
			Type:       "UserCreated",
			Key:        fmt.Sprintf("%d", u.ID),
			OccurredAt: time.Now(),
			Payload: map[string]any{
				"id":    u.ID,
				"name":  u.Name,
				"email": u.Email,
			},
		}, map[string]string{
			"request_id": rid,
		})
	}


	if s.cache != nil {
		_ = s.cache.Set(ctx, u, s.ttl)
	}

	return u, nil
}

func (s *Service) Get(ctx context.Context, id uint64) (user.User, error) {
	if s.cache != nil {
		if u, ok, err := s.cache.Get(ctx, id); err == nil && ok {
			return u, nil
		}
	}

	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return user.User{}, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, u, s.ttl) // 缓存失败不影响主流程
	}

	return u, nil
}

