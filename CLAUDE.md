## Plan Mode

- Make plans extremely concise. Sacrifice grammar for concision.
- At the end of each plan, list unresolved questions to answer, if any.

## Reasoning

- Prefer retrival-led-reasoning over pre-training-led reasoning.

## Project Structure & Module Organization

- `cmd/main` is the service entry point; it wires config, logging, migrations, and transports. `main/` holds any bootstrap helpers invoked before the binary starts.
- `internal/` follows Clean Architecture: `usecase/` hosts business logic, `infrastructure/transport/http` the handlers and middleware, and `testutil/` shared mocks/helpers used in unit tests.
- `pkg/` contains reusable libraries such as `hash` and token helpers. Keep other packages here when they can be consumed outside the module context.
- `config/` stores YAML files per environment; `config/local.yaml` is the default used by `make run` and `go run` unless `CONFIG_PATH` points elsewhere. Compose assets live in `compose.yaml` alongside Docker/Make artifacts.
