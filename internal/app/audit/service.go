package auditapp

import (
	"context"

	"github.com/hacker4257/go-ddd-template/internal/domain/audit"
)

type Service struct {
	repo audit.Repo
}

func New(repo audit.Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, eventType, eventKey string, payload []byte) error {
	return s.repo.Insert(ctx, eventType, eventKey, payload)
}
