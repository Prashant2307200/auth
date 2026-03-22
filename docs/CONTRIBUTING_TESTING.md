# Testing & Test-data Conventions

This project follows a "balanced hybrid" testing approach:

- High-level behavioral tests (usecase and HTTP handler tests)
  - Use seeder-backed `testutil` helpers for realistic fixtures and convenience.
  - Tests live under `internal/usecase` and `internal/infrastructure/transport/http/handler`.

- Low-level repository tests (SQL / sqlmock)
  - Use explicit, minimal inline fixtures tailored to SQL assertions.
  - Avoid coupling repository tests to the seeder/catalog.

Guidelines
- Use `[]*entity.User{}` or `[]*entity.Business{}` for empty-list expectations (no EmptyX helpers).
- Handlers should map domain errors to HTTP status codes using `response.ErrorToStatus(err)` and return safe messages with `response.GeneralError(err)`.
- Validation ownership: keep business/input rules in usecases (`pkg/validator`), not duplicated in HTTP handlers.
- Validation helpers: use `response.FormatValidationErrors()` to return field-level errors from usecase validation failures.
- Seeding: startup seeding is gated; production startup will not seed even if `SEED_ON_STARTUP=true`. Use `SEED_ON_STARTUP=true` only for non-prod maintenance environments or run the dev seed endpoint when `ENV=dev` to seed the DB.
- Secrets policy: never commit real credentials in config or compose defaults; inject via environment variables only.

Verification
- Run tests: `go test ./...`
- Run formatting and static checks before merging: `gofmt -w . && go vet ./... && go build ./...`
- Run linters locally: `golangci-lint run` (CI runs golangci-lint in the lint job)
- Coverage: CI enforces a minimum total coverage (see `.github/workflows/ci.yml`); aim for coverage > 65% across critical packages.

Integration tests (CI parity)
- CI runs auth integration tests with Postgres + Redis services via `.github/workflows/ci.yml`.
- Local run requires:
  - `INTEGRATION_POSTGRES_URI`
  - `INTEGRATION_REDIS_ADDRESS`
  - optional `INTEGRATION_REDIS_USERNAME`, `INTEGRATION_REDIS_PASSWORD`
- Example local command:
  - `INTEGRATION_POSTGRES_URI='postgresql://auth_user:auth_password@localhost:5432/auth_db?sslmode=disable' INTEGRATION_REDIS_ADDRESS='localhost:6379' go test ./internal/infrastructure/transport/http/handler -run TestAuthFlow_Integration -count=1 -v`
