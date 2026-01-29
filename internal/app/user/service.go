package userapp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hacker4257/go-ddd-template/internal/app/tx"
	"github.com/hacker4257/go-ddd-template/internal/domain/event"
	"github.com/hacker4257/go-ddd-template/internal/domain/user"
	"github.com/hacker4257/go-ddd-template/internal/pkg/trace"
)


type Service struct {
	repo  user.Repo
	cache user.Cache
	ttl   time.Duration

	tx     tx.Transactor
	outbox event.Outbox
	topic  string
}


func New(repo user.Repo, cache user.Cache, ttl time.Duration, tx tx.Transactor, outbox event.Outbox, topic string) *Service {
	return &Service{repo: repo, cache: cache, ttl: ttl, tx: tx, outbox: outbox, topic: topic}
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
	
	var created user.User
	err := s.tx.WithinTx(ctx, func(tctx context.Context) error {
		u, err := s.repo.Create(tctx, name, email)
		if err != nil {
			return err
		}
		created = u

		rid := trace.RequestID(tctx)
		return s.outbox.Add(tctx, event.OutboxMessage{
			Topic: s.topic,
			Key: fmt.Sprintf("%d", u.ID),
			Type: "UserCreated",
			Payload: map[string]any{
				"id": u.ID, "name": u.Name, "email": u.Email,
			},
			Headers: map[string]string{
				"request_id": rid,
			},
		})
	})
	if err != nil {
		return user.User{}, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, created, s.ttl)
	}

	return created, nil
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

