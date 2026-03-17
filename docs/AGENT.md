# Repository Guidelines

## Project Structure & Module Organization
- `cmd/main` is the service entry point; it wires config, logging, migrations, and transports. `main/` holds any bootstrap helpers invoked before the binary starts.
- `internal/` follows Clean Architecture: `usecase/` hosts business logic, `infrastructure/transport/http` the handlers and middleware, and `testutil/` shared mocks/helpers used in unit tests.
- `pkg/` contains reusable libraries such as `hash` and token helpers. Keep other packages here when they can be consumed outside the module context.
- `config/` stores YAML files per environment; `config/local.yaml` is the default used by `make run` and `go run` unless `CONFIG_PATH` points elsewhere. Compose assets live in `compose.yaml` alongside Docker/Make artifacts.

## Build, Test, and Development Commands
- `make build` compiles a static `bin/server` binary; `make build-linux` targets Linux builds for Docker images.
- `make run` executes the service with `config/local.yaml` (override `CONFIG_PATH` to swap configs); `make dev` runs `air` for live reload.
- `make test`, `make test-short`, and `make test-coverage` wrap `go test` variants, while `make test-coverage-html` generates HTML reports for manual inspection.
- `make fmt`, `make vet`, and `make lint` enforce formatting, vetting, and linting (requires `golangci-lint`). `make check` chains fmt+vet+test.
- Docker helpers: `make docker-build`, `docker-run`, `docker-stop`, and `docker-clean` mirror the container lifecycle.

## Coding Style & Naming Conventions
- Follow Go conventions: tabs for indentation, exported symbols in CamelCase, private symbols in lowerCamel. Keep files short and focused per package.
- Clean Architecture naming: suffix `UseCase` for application services, `Repo` for persistence interfaces, `Service` for external integrations.
- Run `go fmt ./...`, `go vet ./...`, and `golangci-lint run` before committing to keep formatting and lint rules consistent.

## Testing Guidelines
- Tests live alongside the code they cover (e.g., `internal/usecase/*.go` with matching `_test.go`). Use table-driven test cases with descriptive names like `TestAuthUseCase_RegisterUser_Success`.
- Mock dependencies using `internal/testutil/mocks.go` and helpers (`CreateTestUserWithEmail`). Call `AssertExpectations(t)` on mocks.
- Coverage uses `make test-coverage`; aim to add handler, service, repository, and integration tests to raise the current baseline (see `TESTING.md`).

## Commit & Pull Request Guidelines
- Commit messages should be short, imperative, and descriptive (e.g., “Add JWT refresh endpoint”). Reference related issue IDs when available.
- Every PR should include a summary, testing steps (commands run), and link any relevant issue or ticket. Attach screenshots only if UI changes occur.
- Mention `Makefile` targets that were executed during manual verification; highlight any migrations run for PR reviewers.

## Configuration & Secrets
- Keep secrets out of version control; inject them via environment variables referenced in `config/*.yaml` and override with `export` before `make run`.
- Local development relies on `config/local.yaml`; do not commit sensitive data to that file. Use `CONFIG_PATH` or `--config` to point to sanitized overrides.
