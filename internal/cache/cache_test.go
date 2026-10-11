package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	ctx := context.Background()
	m := miniredis.RunT(t)
	c := &Redis{C: redis.NewClient(&redis.Options{Addr: m.Addr()}), Prefix: "base"}

	var got map[string]int
	require.ErrorIs(t, c.Get(ctx, "k", &got), ErrMiss)

	require.NoError(t, c.Set(ctx, "k", map[string]int{"a": 1}, time.Minute))
	require.NoError(t, c.Get(ctx, "k", &got))
	require.Equal(t, map[string]int{"a": 1}, got)
	require.True(t, m.Exists("base:k"), "keys carry the prefix")

	m.FastForward(2 * time.Minute)
	require.ErrorIs(t, c.Get(ctx, "k", &got), ErrMiss, "expired")

	require.NoError(t, c.Set(ctx, "k", 1, 0))
	require.NoError(t, c.Delete(ctx, "k"))
	require.ErrorIs(t, c.Get(ctx, "k", &got), ErrMiss)
}

// The readiness probe fails while Redis is down.
func TestProbe(t *testing.T) {
	m := miniredis.RunT(t)
	p := probe(redis.NewClient(&redis.Options{Addr: m.Addr(), MaxRetries: -1}))
	require.NoError(t, p.Check(context.Background()))
	m.Close()
	require.Error(t, p.Check(context.Background()))
}
