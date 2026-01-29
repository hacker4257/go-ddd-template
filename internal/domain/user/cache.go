package user

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, id uint64) (User, bool, error) // bool = hit?
	Set(ctx context.Context, u User, ttl time.Duration) error
	Del(ctx context.Context, id uint64) error
}
