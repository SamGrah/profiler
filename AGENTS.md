# AGENTS.md

Guidance for autonomous coding agents working in this repository.

## Scope

- Go service for `cars` CRUD backed by SQLite.
- Layered architecture: API handlers -> service -> repository/DB.
- Keep edits minimal and consistent with existing patterns.

## Rule Files (Cursor/Copilot)

- `.cursorrules`: not present.
- `.cursor/rules/`: not present.
- `.github/copilot-instructions.md`: not present.
- If any appear later, treat them as higher-priority instructions.

## Project Layout

```
cmd/server/main.go           # Entrypoint and dependency wiring
internal/api/                # HTTP handlers, routes, JSend responses
internal/service/            # Validation and business logic
internal/repository/         # Repository interfaces + SQLite impl
internal/models/             # Domain structs
db/schema.sql                # SQLite schema (inventory + cars tables)
api/openapi.yaml             # OpenAPI contract
tests/api_test.go            # Integration tests against running server
```

## Build and Run

```bash
go mod tidy
go build ./...
go run ./cmd/server --addr :8080 --db-path cars.db --schema-path db/schema.sql
```

### Make Targets (via Dagger)

```bash
make build    # dagger call build
make test     # dagger call test
make run      # dagger call run up --ports=8080:8080
make fmt      # dagger call fmt export --path=.
make vet      # dagger call vet
```

## Test Commands

### Full Suite

```bash
CGO_ENABLED=0 go test ./...
make test
```

### Single Test (use `^...$` for exact match)

```bash
CGO_ENABLED=0 go test ./internal/service -run '^TestCarServiceUpdate$' -v
CGO_ENABLED=0 go test ./internal/api -run '^TestCreateCarHandler$' -v
make test-one PKG=./internal/service NAME='^TestCarServiceUpdate$'
```

### Integration Tests

```bash
RUN_API_TESTS=1 API_BASE_URL=http://localhost:8080 CGO_ENABLED=0 go test ./tests -run '^TestCarsAPI_CRUDAgainstRunningServer$' -v
```

**Note:** Use `CGO_ENABLED=0` for all test commands (avoids macOS dyld issues).

## Lint and Formatting

```bash
gofmt -w cmd/server/main.go internal/**/*.go tests/*.go
go vet ./...
```

## Go Style Guidelines

### Formatting

- Rely on `gofmt`; do not hand-align code.
- Keep functions small; prefer early returns.
- Comments only where behavior is non-obvious.

### Imports

Order: standard library, blank line, internal module (`carsapi/...`), blank imports for drivers.

```go
import (
    "context"
    "database/sql"
    "errors"

    "carsapi/internal/models"

    _ "modernc.org/sqlite"
)
```

### Naming

- Exported: `CamelCase` (e.g., `CarService`, `GetByID`)
- Unexported: `camelCase` (e.g., `carService`, `validateCar`)
- Acronyms: `ID`, `VIN`, `HTTP` (not `Id`, `Vin`, `Http`)
- Interfaces: noun/verb + `er` (e.g., `CarRepository`)
- Constructors: `NewX()` returns the interface type

### Types

- Entity IDs: `int64`; Year: `int`
- Prefer `any` over `interface{}`
- JSON tags: snake_case (e.g., `json:"inventory_id"`)

## Error Handling

- Do not ignore meaningful errors in production paths.
- Use `errors.Is` for comparisons, not `==`.
- Wrap with `%w`: `fmt.Errorf("%w: context", err)`
- Map infrastructure errors (e.g., `sql.ErrNoRows`) to domain errors.
- Keep client-facing messages stable; avoid leaking DB internals.

## Testing Guidelines

- Use fakes for unit tests (no mocks/stubs).
- Fakes are map-backed, deterministic, simple.
- Helper constructors: `newFakeCarRepository()`
- Assertions via `t.Fatalf(...)` with explicit expected values.
- Integration tests gated by: `RUN_API_TESTS=1`

## Architecture and DI Rules

- Dependency flow: API -> Service -> Repository -> DB (one-way).
- Wire implementations only in `cmd/server/main.go`.
- Prefer constructor injection (`NewX(...)`) over globals.
- Do not bypass layers in production code.

## API Conventions

- Use standard `net/http` (no framework).
- Routes in `internal/api/routes.go`.
- JSend format:
  - Success: `{"status":"success","data":...}`
  - Fail: `{"status":"fail","message":"..."}` (client error)
  - Error: `{"status":"error","message":"..."}` (server error)
- Status mapping: validation -> 400/fail, not found -> 404/error, unexpected -> 500/error

## Service Layer Conventions

- Methods take `context.Context` first.
- Validate input before calling repository.
- Sentinel errors: `ErrValidation`, `ErrCarNotFound` (in `internal/service/errors.go`)

## Repository / SQL Conventions

- SQL queries in named `const` blocks near usage.
- Parameterized SQL (`?`) only; never concatenate user input.
- Check `RowsAffected()` for update/delete; map zero to `sql.ErrNoRows`.
- Always `defer rows.Close()` and check `rows.Err()`.

## Schema and Contract Rules

- Keep `inventory` table for FK integrity; no inventory endpoints.
- Preserve FK: `PRAGMA foreign_keys = ON`.
- API shape changes must update `api/openapi.yaml`.

## Change Checklist

1. Update only the necessary layer(s).
2. Add/adjust unit tests with fakes.
3. Run `gofmt`, then relevant tests.
4. Run full suite: `make test` or `CGO_ENABLED=0 go test ./...`
5. Update OpenAPI for contract changes.
