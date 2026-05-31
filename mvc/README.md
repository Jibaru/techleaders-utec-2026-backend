# MVC Example — Coffee Shop Loyalty API

A small Go REST API that demonstrates the **Model–View–Controller** pattern — and *only* MVC. It is the first example in a progressive series of architecture patterns for the Tech Leaders talk.

> This example is intentionally minimal. There is **no service layer**, **no repository abstraction**, **no tests**. Every controller talks to the database directly and contains its own business rules. The smells you will notice are the motivation for the next example in the series.

## What the API does

A coffee shop loyalty program:

- Customers earn points on every purchase (1 point per $1, multiplied by their current tier).
- Tiers auto-upgrade as points grow: **Bronze** (0–499) → **Silver** (500–1999, ×1.25) → **Gold** (2000+, ×1.5).
- Customers can redeem points for rewards: `free_drink` (100 pts), `free_pastry` (75 pts).

## How MVC maps to this project

| Layer | Folder | What it contains |
|-------|--------|------------------|
| **Model** | [internal/model/](internal/model/) | Plain data structs (`Customer`, `Purchase`, `Reward`) and constants (tier ladder, reward costs). No behavior. |
| **View** | [internal/view/](internal/view/) | JSON request/response shapes, one sub-package per resource (`customer/`, `purchase/`, `reward/`, `tier/`, `webhook/`) with `request.go` + `response.go` inside each. Cross-cutting shapes (pagination metadata) live in [internal/view/shared/](internal/view/shared/). In a REST API the "view" is the serialized payload returned to the client. |
| **Controller** | [internal/controller/](internal/controller/) | HTTP handlers, one sub-package per resource (`customer/`, `purchase/`, `reward/`, `tier/`, `webhook/`). Each package has a `Controller` struct in `controller.go` and one file per method. Shared HTTP helpers live in [internal/controller/httpx/](internal/controller/httpx/). **Each controller does everything**: parses the request, runs queries via GORM, computes the business rules, handles errors, and renders the JSON view. |

Supporting packages (not part of the MVC discussion — they would exist in any architecture):

- [internal/config/](internal/config/) — loads `.env` into a `Config` struct.
- [internal/db/](internal/db/) — opens a [GORM](https://gorm.io) connection to Postgres and routes GORM's logs through `slog` so everything ends up on the same handler. Schema is created on startup via `db.AutoMigrate(...)` over the model structs (see [cmd/api/main.go](cmd/api/main.go)).
- [internal/router/](internal/router/) — owns the HTTP route table. `router.New(controllers)` returns an `http.Handler`; middleware (request logging, panic recovery) is composed in `main.go`.

## Smells to notice (and what the next example will fix)

Look at [internal/controller/purchase_controller.go](internal/controller/purchase_controller.go) and [internal/controller/webhook_controller.go](internal/controller/webhook_controller.go):

- One controller method knows about GORM queries, point math, tier rules, transactions, idempotency, **and** HTTP status codes. Pick any of `PurchaseController.Create`, `PurchaseController.Refund`, or `WebhookController.PaymentConfirmed` — each one is doing 4–5 jobs.
- The "load customer → mutate points → save" dance is duplicated across `PurchaseController.Create`, `RewardController.Redeem`, `PurchaseController.Refund`, and `WebhookController.PaymentConfirmed`. Four controllers, same business invariant, four implementations.
- **The webhook is the smoking gun**: `WebhookController.PaymentConfirmed` is a second entry point that has to apply the exact same "calculate points based on current tier and update the customer" rule as `PurchaseController.Create`. The duplication is literal — go look. If the rule ever changes, we have to remember to change it in both places.
- The aggregation in `CustomerController.Summary` ("how many points to the next tier?") encodes business knowledge that the controller has no business knowing.
- Testing any of this requires a real database; the rules cannot be tested in isolation.
- Using an ORM does **not** fix any of this. The ORM hides the SQL, but the mixed responsibilities remain.

The **next example** in the series will introduce a **service layer** (extracts the business rules) and a **repository pattern** (extracts persistence). The punchline will be: both `PurchaseController.Create` and `WebhookController.PaymentConfirmed` end up calling the **same** `loyaltyService.RecordPurchase(...)` — and *that* is why the abstraction is worth it.

## Running it

### 1. Set up environment

```bash
cp .env.example .env
# edit .env if your Postgres credentials differ
```

### 2. Create the database

```bash
createdb coffee_loyalty
```

The schema is created automatically on startup via GORM's `AutoMigrate` — no manual migration step.

### 3. Run the server

```bash
make run
```

The server listens on `http://localhost:8080`.

### 4. (Optional) Seed sample data

```bash
make seed
```

Wipes the three tables and inserts four fixed-UUID customers with realistic purchase and reward history. Re-running is safe — it always truncates first.

| Name | UUID | Tier | Points |
|------|------|------|--------|
| Ana Torres   | `11111111-1111-1111-1111-111111111111` | Bronze | 120 |
| Bruno Diaz   | `22222222-2222-2222-2222-222222222222` | Silver | 750 |
| Carla Reyes  | `33333333-3333-3333-3333-333333333333` | Gold   | 2100 |
| Diego Vega   | `44444444-4444-4444-4444-444444444444` | Bronze | 40 |

Quick check after seeding:

```bash
curl -s localhost:8080/customers/33333333-3333-3333-3333-333333333333/summary
```

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET    | `/health` | Health check |
| POST   | `/customers` | Register a customer |
| GET    | `/customers` | List customers — supports `?page=&limit=&tier=&min_points=&max_points=&sort=` |
| GET    | `/customers/{id}` | Get one customer (with tier + points) |
| PATCH  | `/customers/{id}` | Partial update (name and/or email) |
| GET    | `/customers/{id}/summary` | Aggregated stats: spend, points earned/spent, avg purchase, next tier |
| POST   | `/customers/{id}/purchases` | Record a purchase, earn points, maybe upgrade tier |
| GET    | `/customers/{id}/purchases` | List a customer's purchases |
| POST   | `/purchases/{id}/refund` | Refund a purchase (transactional: reverts points, may demote tier) |
| POST   | `/customers/{id}/rewards` | Redeem points for a reward |
| GET    | `/customers/{id}/rewards` | List a customer's redemptions |
| GET    | `/tiers` | Tier definitions |
| GET    | `/rewards/catalog` | Available rewards and their cost |
| POST   | `/webhooks/payment-confirmed` | Idempotent webhook from a payment processor — creates a purchase by email + external payment ID |

## Example session

```bash
# Register Ana
curl -sX POST localhost:8080/customers \
  -H 'content-type: application/json' \
  -d '{"name":"Ana","email":"ana@example.com"}'

# Save her ID (use the value from the previous response)
ID=...

# Record a $5.00 purchase
curl -sX POST localhost:8080/customers/$ID/purchases \
  -H 'content-type: application/json' \
  -d '{"amount_cents":500}'

# Check her balance and tier
curl -s localhost:8080/customers/$ID

# Redeem a free pastry (needs 75 pts — buy more coffee first if needed!)
curl -sX POST localhost:8080/customers/$ID/rewards \
  -H 'content-type: application/json' \
  -d '{"type":"free_pastry"}'
```
