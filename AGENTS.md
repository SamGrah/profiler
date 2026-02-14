# AGENTS.md

Guidance for autonomous coding agents working in this repository.

## Scope

- This is a Go service for `cars` CRUD backed by SQLite.
- Architecture is layered: API handlers -> service -> repository/DB.
- Keep edits minimal and consistent with existing patterns.

## Rule Files (Cursor/Copilot)

- `.cursorrules`: not present.
- `.cursor/rules/`: not present.
- `.github/copilot-instructions.md`: not present.
- If any of these appear later, treat them as higher-priority instructions.

## Project Layout

- `cmd/server/main.go`: application entrypoint and dependency wiring.
- `internal/api/`: HTTP handlers, routes, JSend response helpers.
- `internal/service/`: validation and domain/business logic.
- `internal/repository/`: repository interfaces + SQLite implementation.
- `internal/models/`: domain structs.
- `db/schema.sql`: SQLite schema (`inventory` + `cars`).
- `api/openapi.yaml`: OpenAPI contract.
- `tests/api_test.go`: integration tests against running server.

## Build and Run

- Install dependencies: `go mod tidy`
- Build all packages: `go build ./...`
- Build via make: `make build`
- Run server directly:
  - `go run ./cmd/server --addr :8080 --db-path cars.db --schema-path db/schema.sql`
- Run server via make: `make run`
- Optional binary flow:
  - `go build -o bin/server ./cmd/server`
  - `./bin/server --addr :8080 --db-path cars.db --schema-path db/schema.sql`

## Lint and Formatting

- Always run `gofmt` on changed Go files.
- Example formatting command:
  - `gofmt -w cmd/server/main.go internal/models/*.go internal/repository/*.go internal/service/*.go internal/api/*.go tests/*.go`
- Vet the code: `go vet ./...`
- Optional if installed: `staticcheck ./...`
- Make targets available:
  - `make fmt`
  - `make vet`

## Test Commands

- Recommended full test run in this environment:
  - `CGO_ENABLED=0 go test ./...`
  - `make test`
- Run a single package:
  - `CGO_ENABLED=0 go test ./internal/service -v`
- Run a single test by exact name (important):
  - `CGO_ENABLED=0 go test ./internal/service -run '^TestCarServiceUpdate$' -v`
  - `make test-one PKG=./internal/service NAME='^TestCarServiceUpdate$'`
- Run one API test:
  - `CGO_ENABLED=0 go test ./internal/api -run '^TestCreateCarHandler$' -v`
- Run one repository test:
  - `CGO_ENABLED=0 go test ./internal/repository -run '^TestSQLiteCarRepositoryDelete$' -v`
- Run integration test (server running locally):
  - `RUN_API_TESTS=1 API_BASE_URL=http://localhost:8080 CGO_ENABLED=0 go test ./tests -run '^TestCarsAPI_CRUDAgainstRunningServer$' -v`

## Environment Notes

- On some macOS setups, tests fail with `dyld: missing LC_UUID load command`.
- Use `CGO_ENABLED=0` for test commands in this repo.

## Testing Guidelines

- Use fakes for unit tests (no mocks/stubs in this codebase).
- Fakes should be deterministic and simple (map-backed in-memory state).
- Keep style aligned with current tests:
  - helper constructors like `newFakeX()`
  - direct named tests are acceptable
  - assertions via `t.Fatalf(...)` with explicit expected values
- Integration tests are gated by env var:
  - set `RUN_API_TESTS=1`
  - optional `API_BASE_URL` (default: `http://localhost:8080`)

## Architecture and DI Rules

- Dependency flow must remain one-way:
  - API layer depends on service interfaces.
  - Service layer depends on repository interfaces.
  - Repository layer depends on DB abstraction (`DB`, `Row`, `Rows`).
- Wire concrete implementations in `cmd/server/main.go` only.
- Prefer constructor injection (`NewX(...)`) over globals.
- Do not bypass layers in production code.

## API Conventions

- Use standard `net/http` (no framework).
- Register routes in `internal/api/routes.go`.
- JSend-like response format:
  - success: `{"status":"success","data":...}`
  - fail: `{"status":"fail","message":"..."}`
  - error: `{"status":"error","message":"..."}`
- Error/status mapping:
  - validation errors -> `400` + `fail`
  - not found -> `404` + `error`
  - unexpected -> `500` + `error`

## Service Layer Conventions

- Service methods take `context.Context`.
- Validate input before calling repository methods.
- Keep domain sentinel errors stable:
  - `ErrValidation`
  - `ErrCarNotFound`
- Wrap errors with `%w` when adding context.
- Map infrastructure errors (for example `sql.ErrNoRows`) to domain errors.

## Repository / SQL Conventions

- Keep SQL queries in named `const` blocks near usage.
- Use parameterized SQL (`?`) exclusively.
- Never concatenate user input into SQL strings.
- For update/delete, check `RowsAffected()` and map zero rows to `sql.ErrNoRows`.
- Always `defer rows.Close()` and check `rows.Err()` after iteration.

## Go Style Guidelines

- Formatting: rely on `gofmt`, do not hand-align code.
- Imports order:
  - standard library
  - blank line
  - internal module imports (`carsapi/...`)
  - blank imports only where required (driver registration)
- Naming:
  - exported: `CamelCase`
  - unexported: `camelCase`
  - acronyms: `ID`, `VIN`, `HTTP`
- Types:
  - entity IDs are `int64`
  - `year` is `int`
  - prefer `any` over `interface{}`
- JSON tags should use snake_case (for example `inventory_id`).
- Keep functions small; prefer early returns.
- Add comments only where behavior is non-obvious.

## Error Handling

- Do not ignore meaningful errors in production paths.
- If an error is intentionally ignored, make it explicit and minimal.
- Use `errors.Is` for comparisons.
- Keep client-facing messages stable and actionable.
- Avoid leaking DB internals in API responses.

## Schema and Contract Rules

- Keep the `inventory` table in schema for FK integrity.
- Do not add inventory API endpoints unless explicitly requested.
- Preserve FK behavior (`PRAGMA foreign_keys = ON`).
- Any API shape change must update `api/openapi.yaml`.

## Change Checklist

- Update only the necessary layer(s).
- Add or adjust unit tests with fakes for changed behavior.
- Run `gofmt`, then relevant tests.
- Run full suite when feasible: `make test`.
- Update OpenAPI when request/response contracts change.
