// Package cache is a small key-value cache on Redis: JSON values with a time to live. Use it for
// data that can be recomputed; Postgres stays the truth.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrMiss is returned by Get when the key is absent or expired.
var ErrMiss = errors.New("cache miss")

// Cache is what services use.
type Cache interface {
	// Get decodes the value at key into dst, or returns ErrMiss.
	Get(ctx context.Context, key string, dst any) error
	// Set stores v as JSON at key for ttl (0: no expiry).
	Set(ctx context.Context, key string, v any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

var _ Cache = (*Redis)(nil)

// Redis is the Cache on a Redis client. Every key is prefixed with Prefix (the service's name),
// so services can share a Redis.
type Redis struct {
	C      redis.UniversalClient
	Prefix string
}

func (r *Redis) Get(ctx context.Context, key string, dst any) error {
	b, err := r.C.Get(ctx, r.key(key)).Bytes()
	if errors.Is(err, redis.Nil) {
		return ErrMiss
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

func (r *Redis) Set(ctx context.Context, key string, v any, ttl time.Duration) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return r.C.Set(ctx, r.key(key), b, ttl).Err()
}

func (r *Redis) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	full := make([]string, len(keys))
	for i, k := range keys {
		full[i] = r.key(k)
	}
	return r.C.Del(ctx, full...).Err()
}

func (r *Redis) key(k string) string { return r.Prefix + ":" + k }
