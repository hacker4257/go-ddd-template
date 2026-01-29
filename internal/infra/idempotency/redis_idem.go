package idempotency

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Store struct {
	rdb *goredis.Client
}

func New(rdb *goredis.Client) *Store {
	return &Store{rdb: rdb}
}

// TryMarkProcessed：返回 true 表示首次处理；false 表示重复（应跳过）
func (s *Store) TryMarkProcessed(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	ok, err := s.rdb.SetNX(ctx, fmt.Sprintf("inbox:%s", key), "1", ttl).Result()
	return ok, err
}
