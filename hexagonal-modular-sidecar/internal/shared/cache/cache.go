// Package cache holds the Cache port (interface) used by repository decorators
// to add caching without the service layer ever learning that a cache exists.
// Adapters live in sub-packages: cache/redis/ (real) and cache/noop/ (no-op
// for local dev without a Redis container).
package cache

import (
	"context"
	"errors"
	"time"
)

// ErrMiss is returned by Get when the key is not present in the cache.
// Callers use errors.Is(err, cache.ErrMiss) to distinguish a miss from a
// transport error.
var ErrMiss = errors.New("cache miss")

// Cache is the outbound port for a key/value cache. Implementations decide
// the wire format, TTL semantics, and serialization.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}
