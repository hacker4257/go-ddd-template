package health

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Checker struct {
	DB      *sql.DB
	Redis   *goredis.Client
	Brokers []string

	Timeout time.Duration
}

func (c Checker) Ready(ctx context.Context) error {
	to := c.Timeout
	if to == 0 {
		to = 2 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, to)
	defer cancel()

	// MySQL
	if c.DB != nil {
		if err := c.DB.PingContext(ctx); err != nil {
			return fmt.Errorf("mysql not ready: %w", err)
		}
	}

	// Redis
	if c.Redis != nil {
		if err := c.Redis.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("redis not ready: %w", err)
		}
	}

	// Kafka：做一个 TCP 连通性检查（readiness 足够；后续可升级为 metadata）
	if len(c.Brokers) > 0 {
		b := c.Brokers[0]
		addr := strings.TrimSpace(b)
		d := net.Dialer{}
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err != nil {
			return fmt.Errorf("kafka not ready (tcp %s): %w", addr, err)
		}
		_ = conn.Close()
	}

	return nil
}
