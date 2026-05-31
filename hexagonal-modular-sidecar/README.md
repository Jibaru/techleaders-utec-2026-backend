# Hexagonal Modular + Sidecars — Coffee Shop Loyalty API

Same business behavior and module layout as [../hexagonal-modular/](../hexagonal-modular/), with two real sidecars added:

1. **Redis** — caching for customer reads, plugged in as a new outbound port (`shared/cache.Cache`) and a **decorator** wrapping the customer repository.
2. **Kafka + dedicated `mailworker` process** — the mail send becomes a queue-based pipeline. The API produces; a separate Go binary consumes and runs SMTP.

`docker-compose.yml` wires the whole stack: `postgres + redis + kafka + api + mailworker`.

## What changes vs. the previous example

| Concern | hexagonal-modular | hexagonal-modular-sidecar |
|---------|-------------------|----------------------------|
| Customer reads | hit Postgres every time | cache-aside via Redis (5-min TTL); invalidate on `Update` / `UpdatePoints` |
| Mail send timing | sync — the HTTP handler blocks on SMTP | async — handler returns when Kafka acks; mailworker drains the topic and runs SMTP |
| Outbound ports | `repository.*`, `mail.Sender` | + `cache.Cache` |
| Mail adapters | `smtp`, `noop`, `mock` | + `kafka` |
| Cache adapters | — | `redis`, `noop`, `mock` |
| Binaries | `cmd/api`, `cmd/seed` | + `cmd/mailworker` |
| Orchestration | none | `docker-compose.yml` with healthchecks |

## New ports & adapters

```
internal/shared/
├── cache/
│   ├── cache.go             port: Cache interface + ErrMiss sentinel
│   ├── redis/cache.go       adapter using github.com/redis/go-redis/v9
│   ├── noop/cache.go        adapter that always misses (for local dev)
│   └── mock/cache.go        mockery-generated
└── mail/
    ├── sender.go            port (unchanged from previous example)
    ├── smtp/sender.go       direct SMTP send
    ├── noop/sender.go       log-only
    ├── kafka/sender.go      NEW: produce JSON to a Kafka topic
    └── mock/sender.go

internal/customer/repository/
├── repository.go            port (unchanged)
├── gorm/repository.go       GORM adapter
├── cached/repository.go     NEW: decorator that wraps any inner Repository
└── mock/repository.go

cmd/
├── api/main.go              picks cache + mail adapters from env
├── seed/main.go             unchanged
└── mailworker/main.go       NEW: kafka consumer → SMTP
```

## The decorator pattern for caching

[`internal/customer/repository/cached/repository.go`](internal/customer/repository/cached/repository.go) wraps the GORM customer repository:

```go
func (r *Repository) FindByID(ctx, id) (model.Customer, error) {
    // Cache-aside
    if raw, err := r.cache.Get(ctx, key(id)); err == nil {
        var c model.Customer
        if json.Unmarshal(raw, &c) == nil {
            return c, nil
        }
    }
    customer, err := r.inner.FindByID(ctx, id)
    if err != nil { return customer, err }
    raw, _ := json.Marshal(customer)
    _ = r.cache.Set(ctx, key(id), raw, r.ttl)
    return customer, nil
}

func (r *Repository) UpdatePoints(ctx, id, points) error {
    if err := r.inner.UpdatePoints(ctx, id, points); err != nil { return err }
    _ = r.cache.Delete(ctx, key(id))   // write-through invalidate
    return nil
}
```

**The service layer never imports this package.** It depends on `customer.Repository`. `main.go` is the only place that wraps:

```go
customerGorm  := customergorm.New(gormDB)
customerRepo  := customercached.New(customerGorm, cacheAdapter)   // ← decoration
```

Set `REDIS_ADDR=""` → the noop cache adapter is selected → decorator passes everything straight through. The customer service can't tell the difference.

### Known consistency trade-off

Cache invalidation runs inside the `cached` decorator's `Update` and `UpdatePoints`. But `UpdatePoints` calls from `purchase.Service.Record` / `Refund` and `reward.Service.Redeem` go through `repos.Customers` inside a `Transactor.RunTx` — and that `repos.Customers` is the **gorm-tx-bound repo**, not the cached decorator. So after a purchase commits, the cache may serve a stale `Points` value until the TTL expires or an outer `customer.Service.Update` fires.

For a coffee-shop demo with a 5-minute TTL this is fine. The production fix is to either (a) have the Transactor wrap each tx-bound repo with the same cache decorator, or (b) move invalidation to a post-commit outbox/event. Deliberately out of scope here to keep the example focused.

## Async mail via Kafka

The same `mail.Sender` interface gets a new adapter:

```go
// internal/shared/mail/kafka/sender.go
func (s *Sender) Send(ctx, to, subject, body string) error {
    payload, _ := json.Marshal(Message{To: to, Subject: subject, Body: body})
    return s.writer.WriteMessages(ctx, kafka.Message{Key: []byte(to), Value: payload})
}
```

`cmd/mailworker/main.go` is the consumer side: reads from the `mail.outbox` topic, decodes the JSON, hands off to the existing SMTP sender, and only commits the Kafka offset after a successful send (so a transient SMTP failure causes redelivery).

**Adapter selection in main.go:**

```go
switch {
case len(cfg.Kafka.Brokers) > 0:
    mailer = mailkafka.New(cfg.Kafka.Brokers, cfg.Kafka.MailTopic)   // queue
case cfg.Mail.Host == "":
    mailer = mailnoop.New()                                          // log only
default:
    mailer = mailsmtp.New(...)                                       // sync SMTP
}
```

The service still calls `mailer.Send(ctx, to, subject, body)`. Three deployment modes, zero service-layer changes.

## Running the stack

### With docker-compose (the recommended path)

```bash
docker compose up --build
# or
make up
```

This brings up postgres, redis, kafka, the api (on `:8080`), and the mailworker. The api uses Kafka for mail; mailworker drains the topic.

To watch the pipeline:

```bash
make logs                                    # api + mailworker stdout
docker compose exec kafka kafka-console-consumer.sh \
    --bootstrap-server localhost:9092 \
    --topic mail.outbox --from-beginning     # raw messages
```

Then trigger a purchase against the seeded data:

```bash
make seed   # one-shot fixture inside the api container (run from host go)
ID=11111111-1111-1111-1111-111111111111
curl -sX POST localhost:8080/customers/$ID/purchases \
    -H 'content-type: application/json' \
    -d '{"amount_cents": 500}'
```

You'll see the api log "kafka mail sender" producing, then the mailworker log "sent" (or "would send" if `SMTP_HOST` is empty).

### Without docker (local dev)

Leave `REDIS_ADDR` and `KAFKA_BROKERS` empty in `.env`. The cache becomes noop; the mail sender becomes SMTP-or-noop. The example behaves identically to the previous module.

```bash
make run
```

## Tests

```bash
make test
```

Adds a new test file: [`internal/customer/repository/cached/repository_test.go`](internal/customer/repository/cached/repository_test.go) — 8 subtests covering cache hit / miss / transport error / corrupt payload / inner error / write-through invalidate. All offline; uses mockery-generated mocks for both the inner `Repository` and the `Cache` ports.

The existing 47 service subtests still pass unchanged — they depend on the `customer.Repository` interface, and they neither know nor care that the production wiring decorates it with caching.

## What this example shows

1. **Adding a new outbound concern (cache) is symmetric to the existing ones**: define a port (`cache.Cache`), write adapters (`redis/`, `noop/`, `mock/`), pick one in `main.go`. No service code changes.
2. **Decorators are the right tool when you want to wrap an existing port without modifying the consumer.** The customer service still depends on `customer.Repository`; the cache is one of many `Repository` implementations.
3. **Switching mail from sync SMTP to async Kafka is one new adapter + one new binary.** Service code is identical. The composition root chooses based on config.
4. **Sidecar architecture is just deployment.** From the Go code's perspective, Redis is "a Cache adapter" and Kafka is "a Sender adapter". Hexagonal makes the deployment topology a `main.go` concern, not a service-layer concern.
