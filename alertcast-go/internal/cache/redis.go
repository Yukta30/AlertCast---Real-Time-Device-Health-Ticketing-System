package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	c *redis.Client
}

func New(addr string) *Redis {
	return &Redis{
		c: redis.NewClient(&redis.Options{Addr: addr}),
	}
}

func (r *Redis) Ping(ctx context.Context) error {
	return r.c.Ping(ctx).Err()
}

func (r *Redis) IncSeverity(ctx context.Context, severity string) error {
	return r.c.HIncrBy(ctx, "severity_counts", severity, 1).Err()
}

func (r *Redis) GetAllSeverity(ctx context.Context) (map[string]int64, error) {
	m, err := r.c.HGetAll(ctx, "severity_counts").Result()
	if err != nil {
		return nil, err
	}
	out := map[string]int64{"critical":0,"high":0,"medium":0,"low":0}
	for k, v := range m {
		// parse int64
		var n int64
		_, _ = fmt.Sscan(v, &n)
		out[k] = n
	}
	return out, nil
}
