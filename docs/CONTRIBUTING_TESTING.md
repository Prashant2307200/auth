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
- Validation helpers: use `response.ValidationError()` and `response.FormatValidationErrors()` to return field-level errors.
- Seeding: startup seeding is gated. Use `SEED_ON_STARTUP=true` or run the dev seed endpoint when `ENV=dev` to seed the DB.

Verification
- Run tests: `go test ./...`
- Run lints for touched areas before merging.

