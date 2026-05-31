# Hexagonal Architecture — Coffee Shop Loyalty API

The final example in the Tech Leaders series. Same domain as [../mvc-service-repo/](../mvc-service-repo/), but the dependency on GORM has been pushed entirely to the edge. Services depend only on **ports** (interfaces) — never on adapters — and every port has both a real GORM adapter and a generated mock.

## What hexagonal adds vs. the previous example

| Concern | mvc-service-repo | This example |
|---------|------------------|--------------|
| Service ↔ Repository | Service depends on the concrete `*Repository` struct | Service depends on the `Repository` *interface*. The GORM struct is just one of many possible implementations. |
| Transactions | Service takes `*gorm.DB` and calls `.Transaction(...)` directly | Service depends on the `Transactor` port. The GORM adapter wraps `db.Transaction(...)`. |
| Testability | Untestable without a real Postgres | Every method has unit tests against mocks. `go test ./...` runs offline. |
| GORM blast radius | Imported by the service layer for transactions | Imported only by `cmd/api/main.go` and the `*/gorm/` adapter packages. Swap them out and nothing else moves. |

## Folder structure

```
internal/
├── controller/                       (unchanged)
├── service/
│   ├── customer/   <-- now depends on the repository.Repository INTERFACES, not the GORM structs
│   ├── purchase/   <-- depends on repository.Transactor too
│   └── reward/     <-- depends on repository.Transactor too
└── repository/
    ├── transactor.go                 PORT: Transactor + TxRepos
    ├── gorm/transactor.go            ADAPTER: GORM impl of Transactor
    ├── mock/transactor.go            mockery-generated mock
    ├── customer/
    │   ├── repository.go             PORT: Repository interface + ListFilter
    │   ├── gorm/repository.go        ADAPTER: GORM impl
    │   └── mock/repository.go        mockery-generated mock
    ├── purchase/
    │   ├── repository.go  gorm/  mock/
    └── reward/
        ├── repository.go  gorm/  mock/
```

The single most important property: **the only Go file in the codebase that imports `gorm.io/gorm` and also wires repositories is `cmd/api/main.go`**. Everything else either lives in a `gorm/` adapter package (which knows it's GORM) or depends only on a port (which doesn't).

## Mocks (mockery v3)

`.mockery.yml` declares the mocks; the `mock/` subpackage next to each port holds the generated file.

```bash
make install-mockery   # one-time: go install github.com/vektra/mockery/v3@latest
make mocks             # regenerate every mock listed in .mockery.yml
```

The generated mocks use the v3 `EXPECT()` pattern with the testify template:

```go
mocks.Customers.EXPECT().FindByID(ctx, id).Return(customer, nil)
```

## Tests

```bash
make test
```

Layout: one `*_test.go` file per service method (`create_test.go`, `update_test.go`, etc.) with subtests inside (`t.Run("happy path", ...)`, `t.Run("invalid email", ...)`) — **no test tables**. Each subtest sets up a fresh service via `newService(t)`, declares its expectations on the mocks, runs the call, and the mock's auto-cleanup asserts every expectation was met.

For methods that go through the Transactor, `expectTxRun(ctx)` arranges for the mocked `RunTx` to invoke the callback against the same mock repos the test already configured.

## What's in each test file

| File | Scenarios |
|------|-----------|
| `customer/create_test.go` | happy path with trim/lowercase, invalid name, invalid email, email-already-exists |
| `customer/get_test.go` | happy path, not found |
| `customer/get_by_email_test.go` | happy path with normalization, invalid format, not found |
| `customer/update_test.go` | both fields, name only, invalid name, invalid email, empty input, not found, email collision |
| `customer/list_test.go` | defaults, tier→points-range, invalid sort, unknown tier, limit cap + pagination math |
| `customer/summary_test.go` | full aggregate, gold customer (no next tier), customer not found |
| `purchase/record_test.go` | bronze→bronze, tier-upgrade, invalid amount, invalid external id, customer not found, idempotency hit, race during tx, OccurredAt respected |
| `purchase/list_test.go` | happy path, customer not found |
| `purchase/refund_test.go` | happy path, tier demotion, purchase not found, already refunded, points already spent |
| `reward/redeem_test.go` | happy path, unknown reward, customer not found, insufficient points |
| `reward/list_test.go` | happy path, customer not found |

## Running it

Same as the previous example:

```bash
cp .env.example .env
createdb coffee_loyalty
make run          # AutoMigrate runs on startup
make seed         # (optional) fixed-UUID fixtures
make test         # offline; no DB required
```
