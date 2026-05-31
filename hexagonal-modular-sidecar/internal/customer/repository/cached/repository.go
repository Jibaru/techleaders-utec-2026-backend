// Package cached holds a decorator for customer.Repository that adds
// cache-aside semantics on top of any inner adapter (typically the GORM one).
//
// The service layer never imports this package — it depends only on the
// customer.Repository port. main.go is the only place that knows to wrap the
// GORM repo with the cached decorator. That's the hexagonal-pure way to add
// caching: a new outbound port (cache.Cache) + a decorator that composes
// the two existing pieces.
package cached

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"hexagonal-modular-sidecar/internal/customer/model"
	customerrepo "hexagonal-modular-sidecar/internal/customer/repository"
	"hexagonal-modular-sidecar/internal/shared/cache"
)

const defaultTTL = 5 * time.Minute

type Repository struct {
	inner customerrepo.Repository
	cache cache.Cache
	ttl   time.Duration
}

func New(inner customerrepo.Repository, c cache.Cache) *Repository {
	return &Repository{inner: inner, cache: c, ttl: defaultTTL}
}

func key(id uuid.UUID) string {
	return fmt.Sprintf("customer:%s", id)
}

// FindByID applies cache-aside: try the cache, fall back to the inner repo
// on miss, then populate the cache. Cache errors are logged-and-ignored so a
// Redis hiccup never fails a read.
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (model.Customer, error) {
	k := key(id)

	if raw, err := r.cache.Get(ctx, k); err == nil {
		var c model.Customer
		if jerr := json.Unmarshal(raw, &c); jerr == nil {
			return c, nil
		}
		slog.WarnContext(ctx, "cache: bad payload, falling back to db", "key", k)
	} else if !errors.Is(err, cache.ErrMiss) {
		slog.WarnContext(ctx, "cache get failed", "key", k, "err", err)
	}

	customer, err := r.inner.FindByID(ctx, id)
	if err != nil {
		return customer, err
	}
	if raw, mErr := json.Marshal(customer); mErr == nil {
		if sErr := r.cache.Set(ctx, k, raw, r.ttl); sErr != nil {
			slog.WarnContext(ctx, "cache set failed", "key", k, "err", sErr)
		}
	}
	return customer, nil
}

// UpdatePoints invalidates the cache after the underlying write succeeds.
// Called from purchase + reward services via TxRepos, so this invalidation
// keeps the cache consistent across modules.
func (r *Repository) UpdatePoints(ctx context.Context, id uuid.UUID, points int) error {
	if err := r.inner.UpdatePoints(ctx, id, points); err != nil {
		return err
	}
	r.invalidate(ctx, id)
	return nil
}

// Update invalidates the cache after a successful write.
func (r *Repository) Update(ctx context.Context, id uuid.UUID, updates map[string]any) (model.Customer, error) {
	customer, err := r.inner.Update(ctx, id, updates)
	if err != nil {
		return customer, err
	}
	r.invalidate(ctx, id)
	return customer, nil
}

func (r *Repository) invalidate(ctx context.Context, id uuid.UUID) {
	if err := r.cache.Delete(ctx, key(id)); err != nil {
		slog.WarnContext(ctx, "cache delete failed", "key", key(id), "err", err)
	}
}

// ----- Pass-through methods (no cache effect) -------------------------------

func (r *Repository) Save(ctx context.Context, c *model.Customer) error {
	return r.inner.Save(ctx, c)
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (model.Customer, error) {
	return r.inner.FindByEmail(ctx, email)
}

func (r *Repository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	return r.inner.ExistsByID(ctx, id)
}

func (r *Repository) List(ctx context.Context, f customerrepo.ListFilter) ([]model.Customer, int64, error) {
	return r.inner.List(ctx, f)
}

// Trade-off note: UpdatePoints calls made INSIDE a transaction (via the
// Transactor's TxRepos) go straight to the gorm-tx-bound repo and bypass
// this decorator's invalidation. That means after a purchase or refund
// commits, the cache may serve a stale Points value until the TTL expires
// or someone calls the outer customer.Service.Update. Acceptable for a
// loyalty demo; production code would either use a Transactor that wraps
// tx-bound repos with the same cache, or invalidate via an outbox/event.
