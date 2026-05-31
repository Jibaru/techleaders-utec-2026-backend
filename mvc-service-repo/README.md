# MVC + Service Layer + Repositories — Coffee Shop Loyalty API

The second example in the Tech Leaders series. Same domain as [../mvc/](../mvc/), but the fat MVC controllers have been split into three layers:

- **Controller** parses the HTTP request, validates input format, calls the service, renders the JSON view, and maps domain errors to status codes.
- **Service** holds the business rules (point calculation, tier upgrades, redemption, refund reversal, webhook idempotency). It knows nothing about HTTP.
- **Repository** owns all GORM queries. It translates infrastructure errors (`gorm.ErrRecordNotFound`, `ErrDuplicatedKey`) into domain errors so the service never sees a GORM type.

## What changed vs. the MVC example

| Concern | MVC example | This example |
|---------|-------------|--------------|
| Business rules | scattered across controllers | live in `internal/service/*` |
| SQL / GORM queries | inline in every controller | live in `internal/repository/*` |
| Domain errors | inline strings + `gorm.ErrRecordNotFound` checks | declared in `internal/model/errors.go`, mapped to HTTP by `httpx.MapDomainError` |
| Webhook duplication | 30 lines copy-pasted from `purchase.Create` | one call to `purchase.Service.Record(...)` — same call the regular endpoint makes |
| Refund logic | a long inline `db.Transaction` in the controller | `purchase.Service.Refund(...)` — controller is 15 lines |
| Customer summary | three SQL queries + aggregation in the controller | `customer.Service.GetSummary(...)` — controller is 30 lines, mostly view assembly |

The point of this example is to show what the **service + repository extraction buys**. The next example will add multi-module organization; the example after that, hexagonal architecture + tests.

## Folder structure

```
internal/
├── controller/                   parse HTTP → call service → render view
│   ├── httpx/                    helpers + validators + domain-error → HTTP mapper
│   ├── customer/  purchase/  reward/  tier/  webhook/
├── service/                      business logic; knows nothing about HTTP
│   ├── customer/                 Create, Get, GetByEmail, Update, List, GetSummary
│   ├── purchase/                 Record (with idempotency), List, Refund
│   └── reward/                   Redeem, List
├── repository/                   GORM queries; translates infra errors → domain errors
│   ├── customer/                 + WithTx for transactional composition by services
│   ├── purchase/                 includes AggregateForCustomer
│   └── reward/                   includes AggregateForCustomer
├── model/                        entities + domain errors + tier helpers + reward costs
├── view/                         JSON shapes (unchanged from MVC example)
├── config/  db/  router/
```

**Dependency direction:** `controller → service → repository → model`. Services don't import GORM types except `*gorm.DB` for running transactions (`s.db.Transaction(func(tx) { repo.WithTx(tx).X(...) })`). The hexagonal example will hide even that behind an interface.

## What's deliberately NOT here

To keep this example focused on the **single new concept** (extracting services + repos), the following are deferred to later examples:

- **Repository interfaces.** Services depend on concrete repo structs. Without interfaces, services can't be unit-tested without a real database — that's the trade-off the hexagonal example fixes.
- **Tests.** Same reason. Adding tests here would force interfaces, which is the next lesson.
- **Multiple top-level modules / domains.** The customer/purchase/reward split is by *resource*, not by bounded context. Modular structure comes in example #3.

## Running it

```bash
cp .env.example .env       # adjust DB creds if needed
createdb coffee_loyalty
make run                   # AutoMigrate runs on startup
make seed                  # (optional) wipe + insert fixed-UUID fixtures
```

Endpoints are identical to the MVC example — same routes, same JSON shapes. The difference is entirely internal.

## The webhook punchline

Open [internal/controller/webhook/payment_confirmed.go](internal/controller/webhook/payment_confirmed.go) and [internal/controller/purchase/create.go](internal/controller/purchase/create.go) side-by-side. Both end with:

```go
result, err := c.svc.Record(r.Context(), purchaseservice.RecordInput{ ... })
```

That `purchaseservice.Record` is the SAME function. The 30-line duplicated block from the MVC example is gone. If the point-calculation rule changes, you change it in one place. If you add a CLI tool or a background job that needs to record purchases, it's another one-liner.

**That is what the service layer buys you.**
