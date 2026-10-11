package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	"github.com/tanjed/base-go/internal/config"
	"github.com/tanjed/base-go/internal/health"
)

// Module provides the Redis client (APP_REDIS_URL) and the Cache on it, and contributes the
// "redis" health probe.
var Module = fx.Module("cache",
	fx.Provide(
		newClient,
		fx.Annotate(New, fx.As(new(Cache))),
		fx.Annotate(probe, fx.ResultTags(health.ProbeGroup)),
	),
)

func New(c redis.UniversalClient, cfg config.Config) *Redis { return &Redis{C: c, Prefix: cfg.Name} }

func newClient(lc fx.Lifecycle, cfg config.Config) (redis.UniversalClient, error) {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("APP_REDIS_URL: %w", err)
	}
	c := redis.NewClient(opts)
	lc.Append(fx.StopHook(func(context.Context) error { return c.Close() }))
	return c, nil
}

// probe: Redis answers a ping.
func probe(c redis.UniversalClient) health.Probe {
	return health.Probe{Name: "redis", Check: func(ctx context.Context) error { return c.Ping(ctx).Err() }}
}
