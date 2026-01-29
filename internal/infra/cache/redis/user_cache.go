package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/hacker4257/go-ddd-template/internal/domain/user"
)

type UserCache struct {
	rdb *goredis.Client
}

func NewUserCache(rdb *goredis.Client) *UserCache {
	return &UserCache{rdb: rdb}
}

func (c *UserCache) key(id uint64) string {
	return fmt.Sprintf("user:%d", id)
}

func (c *UserCache) Get(ctx context.Context, id uint64) (user.User, bool, error) {
	val, err := c.rdb.Get(ctx, c.key(id)).Result()
	if err == goredis.Nil {
		return user.User{}, false, nil
	}
	if err != nil {
		return user.User{}, false, err
	}

	var u user.User
	if err := json.Unmarshal([]byte(val), &u); err != nil {
		// 解析失败：当作 miss（也可以顺手删掉坏缓存）
		_ = c.rdb.Del(ctx, c.key(id)).Err()
		return user.User{}, false, nil
	}

	return u, true, nil
}

func (c *UserCache) Set(ctx context.Context, u user.User, ttl time.Duration) error {
	b, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, c.key(u.ID), b, ttl).Err()
}

func (c *UserCache) Del(ctx context.Context, id uint64) error {
	return c.rdb.Del(ctx, c.key(id)).Err()
}
