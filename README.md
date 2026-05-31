# Tech Leaders Examples

Architecture-pattern examples for the Tech Leaders talk. Each subfolder is a self-contained Go REST API that demonstrates one pattern applied to a realistic problem.

## Examples

| # | Pattern | Folder | Adds |
|---|---------|--------|------|
| 1 | MVC | [mvc/](mvc/) | Baseline: fat controllers do HTTP + business + DB all in one method |
| 2 | MVC + Service + Repository | [mvc-service-repo/](mvc-service-repo/) | Extracts business logic into services, GORM queries into repositories |

