# Tech Leaders Examples

Architecture-pattern examples for the Tech Leaders talk. Each subfolder is a self-contained Go REST API that demonstrates one pattern applied to a realistic problem.

## Examples

| # | Pattern | Folder | Adds |
|---|---------|--------|------|
| 1 | MVC | [mvc/](mvc/) | Baseline: fat controllers do HTTP + business + DB all in one method |
| 2 | MVC + Service + Repository | [mvc-service-repo/](mvc-service-repo/) | Extracts business logic into services, GORM queries into repositories |
| 3 | Hexagonal (ports & adapters) | [hexagonal/](hexagonal/) | Services depend on repository *interfaces*; GORM is a swappable adapter; unit tests against mockery-generated mocks |
| 4 | Hexagonal — modular | [hexagonal-modular/](hexagonal-modular/) | Same architecture, reorganized by **module** (`internal/customer/`, `internal/purchase/`, …) with each module owning its own controller/service/repository/view/router and a shared `internal/shared/` for cross-cutting code |
| 5 | Hexagonal — modular + sidecars | [hexagonal-modular-sidecar/](hexagonal-modular-sidecar/) | Adds **Redis** (cache, plugged in as a decorator on the customer repo) and **Kafka** (async mail with a dedicated `cmd/mailworker` consumer). Ships a `docker-compose.yml` for the full stack |

