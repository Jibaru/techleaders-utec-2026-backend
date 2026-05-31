// Package noop holds a Cache adapter that always misses, never stores.
// Used when REDIS_ADDR is empty so the app still runs without a Redis sidecar.
package noop

import (
	"context"
	"time"

	"hexagonal-modular-sidecar/internal/shared/cache"
)

type Cache struct{}

func New() *Cache { return &Cache{} }

func (Cache) Get(_ context.Context, _ string) ([]byte, error) {
	return nil, cache.ErrMiss
}

func (Cache) Set(_ context.Context, _ string, _ []byte, _ time.Duration) error {
	return nil
}

func (Cache) Delete(_ context.Context, _ string) error {
	return nil
}
