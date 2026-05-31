# Hexagonal Architecture (Modular) — Coffee Shop Loyalty API

Same business behavior as [../hexagonal/](../hexagonal/), reorganized into **one folder per domain module**. Every module has its own layered structure (controller, service, repository, view, router); cross-cutting code lives under `internal/shared/`.

## What changes vs. the previous example

| Concern | hexagonal | hexagonal-modular |
|---------|-----------|-------------------|
| Grouping primary axis | by **layer** (`internal/controller/`, `internal/service/`, `internal/repository/`, …) | by **module** (`internal/customer/`, `internal/purchase/`, …) — each module *contains* its layers |
| Cross-cutting code | scattered under `internal/{model,mail,validate,config,db,httpx}` | gathered under `internal/shared/{...}` |
| Routing | one `internal/router/router.go` with a `Controllers` struct + 13 route lines | one `Register(mux, ctrl)` per module; `main.go` calls each module's `Register` |
| Package names | `customer`, `purchase`, `reward`, `tier`, `webhook` (one per layer-folder) | `controller`, `service`, `repository`, `view`, `router` — every module uses these same short names. Cross-module imports always use aliases (`customerctl`, `purchasesvc`, etc.). |

## Folder structure

```
internal/
├── customer/                          everything customer-related lives here
│   ├── controller/   package controller
│   │   ├── controller.go              struct + constructor
│   │   ├── create.go  get.go  get_by_email.go  list.go  update.go  summary.go
│   ├── service/      package service
│   │   ├── service.go
│   │   ├── create.go  get.go  ... (one file per method, same as before)
│   ├── repository/   package repository  (the PORT)
│   │   ├── repository.go              interface + ListFilter
│   │   ├── gorm/repository.go         GORM adapter
│   │   └── mock/repository.go         mockery-generated mock
│   ├── view/         package view
│   │   ├── request.go  response.go
│   └── router/       package router
│       └── router.go                  Register(mux, *controller.Controller)
├── purchase/                          same layout
├── reward/                            same layout
├── tier/                              controller/ + view/ + router/ only (no service/repo — pure static catalog)
├── webhook/                           controller/ + view/ + router/ (depends on customer + purchase services)
└── shared/                            cross-cutting concerns
    ├── model/                         entities, tier rules, reward costs, domain errors
    ├── mail/                          Sender port + templates + smtp/ + noop/ + mock/ adapters
    ├── validate/                      ValidateName, ValidateEmail, …
    ├── config/                        .env loader
    ├── db/                            GORM connection + slog adapter
    ├── repository/                    Transactor port + TxRepos + gorm/ + mock/ adapters
    ├── httpx/                         WriteJSON, MapDomainError, helpers
    └── view/                          ListMeta (cross-module response shape)
```

## Module dependency rules

- A module's `controller` may import its own `service` and shared `httpx` / cross-module `view` (for embedded shapes).
- A module's `service` may import its own and other modules' `repository` ports, plus `shared/repository.Transactor` and `shared/mail.Sender`.
- A module's `repository/gorm` imports its own `repository` port and `gorm.io/gorm`. Nothing else.
- A module's `repository/mock` imports its own `repository` port and `testify/mock`. Nothing else.
- A module's `router` imports its own `controller` and registers routes.
- `shared/` modules can be imported by anyone but never import a domain module — **except** `shared/repository/transactor.go`, which bundles the three repository interfaces into `TxRepos`. That single inward import is the price of having a typed transactor over three repos.

## Cross-module imports use aliases

Because every module's `controller` package is literally named `controller`, importing multiple modules' controllers into the same file requires aliases. The convention is `<module><layer>`:

```go
import (
    customerctl    "hexagonal-modular/internal/customer/controller"
    purchasesvc    "hexagonal-modular/internal/purchase/service"
    rewardgorm     "hexagonal-modular/internal/reward/repository/gorm"
    sharedview     "hexagonal-modular/internal/shared/view"
    customerrouter "hexagonal-modular/internal/customer/router"
    repogorm       "hexagonal-modular/internal/shared/repository/gorm"
)
```

[`cmd/api/main.go`](cmd/api/main.go) is the canonical example — it imports all five modules' controllers, services, gorm adapters, and routers.

## Per-module routers

Each module owns its URL surface:

```go
// internal/customer/router/router.go
func Register(mux *http.ServeMux, c *controller.Controller) {
    mux.HandleFunc("POST /customers",                c.Create)
    mux.HandleFunc("GET /customers",                 c.List)
    mux.HandleFunc("GET /customers/{id}",            c.Get)
    mux.HandleFunc("PATCH /customers/{id}",          c.Update)
    mux.HandleFunc("GET /customers/{id}/summary",    c.Summary)
}
```

`main.go` creates the shared mux, attaches `/health` directly, and calls every module's `Register`:

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /health", health)
customerrouter.Register(mux, customerController)
purchaserouter.Register(mux, purchaseController)
rewardrouter.Register(mux, rewardController)
tierrouter.Register(mux, tierController)
webhookrouter.Register(mux, webhookController)
```

## Mocks (mockery v3)

Same setup as before — `.mockery.yml` now points at the per-module repository ports:

```bash
make install-mockery   # one-time
make mocks             # regenerate every mock
make test              # offline; no DB required
```

Mocks land next to each port: `internal/customer/repository/mock/`, `internal/shared/mail/mock/`, etc.

## Running it

```bash
cp .env.example .env
createdb coffee_loyalty
make run               # AutoMigrate runs on startup
make seed              # (optional) fixed-UUID fixtures
make test              # 37+ subtests, all offline
```

API surface is identical to the previous example. The change is purely internal organization.

## What this layout buys you

- **Module ownership scales with team size.** A team that owns "purchases" works inside `internal/purchase/` and rarely touches sibling modules.
- **A new module is one new directory.** `internal/notifications/` with the same `controller/service/repository/view/router/` shape slots in cleanly.
- **Smaller blast radius.** Renaming a customer view field touches only files under `internal/customer/`. The previous layout would spread the same change across `internal/view/customer/`, `internal/controller/customer/`, `internal/service/customer/` — three top-level trees instead of one subtree.
- **The cost is import-alias verbosity** at the few seams that span modules (mostly `main.go`, the webhook controller, and the shared transactor). Acceptable in exchange for the locality wins.
