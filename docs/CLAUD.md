# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go microservice for authentication in the Chatty platform. Provides JWT-based auth with PostgreSQL persistence and Redis session storage. Part of a larger microservices architecture.

## Build & Run

```bash
# Build
go build ./cmd/main

# Run (requires config path)
go run ./cmd/main --config=config/local.yaml

# Docker
docker build -t auth-service .
docker run -p 8080:8080 auth-service
```

No test suite or linter is currently configured. Makefile exists but is empty.

## Architecture

Clean Architecture with four layers:

- **Entity** (`internal/entity/`) — Domain structs (User, Login)
- **Use Case** (`internal/usecase/`) — Business logic; interfaces defined in `usecase/interfaces/`
- **Infrastructure** (`internal/infrastructure/`) — HTTP/gRPC handlers, repository implementations
- **Packages** (`pkg/`) — Reusable wrappers for PostgreSQL, Redis, bcrypt, Cloudinary

Dependency flow: `infrastructure → usecase → entity`, with interfaces in `usecase/interfaces/` inverted so infrastructure implements them.

### Entry Point

`cmd/main/main.go` wires everything: loads config, connects DB/Redis, constructs repositories → services → use cases → handlers, registers routes under `/api/v1`, starts HTTP server with graceful shutdown.

### Key Services

- **TokenService** (`internal/service/token.go`) — RSA-signed JWT access tokens, Redis-stored refresh tokens (7-day TTL)
- **UserRepository** (`internal/infrastructure/repository/user.go`) — Raw SQL against PostgreSQL via `database/sql` + `lib/pq`
- **Auth middleware** (`internal/infrastructure/transport/http/middleware/`) — Validates JWT on protected routes

### API Routes

All under `/api/v1`:

| Method | Path | Auth |
|--------|------|------|
| POST | `/auth/register/` | No |
| POST | `/auth/login/` | No |
| DELETE | `/auth/logout/` | Yes |
| GET/PUT/DELETE | `/auth/profile/` | Yes |
| GET | `/auth/refresh/` | No |
| GET | `/auth/public-key` | No |
| GET | `/users/` | Yes |

## Configuration

YAML config loaded via `cleanenv`. Path set by `--config` flag or env var. Key env vars:

- `POSTGRES_URI`, `REDIS_ADDRESS`, `REDIS_USERNAME`, `REDIS_PASSWORD`
- `ACCESS_TOKEN_SECRET`, `REFRESH_TOKEN_SECRET`, `COOKIE_SECRET`
- `NAME`, `API_KEY`, `API_SECRET` (Cloudinary)

RSA keys expected at `keys/public.pem` and `keys/private.pem`.

## Key Dependencies

- HTTP: Go stdlib `net/http` (no framework)
- JWT: `golang-jwt/jwt/v5`
- DB: `lib/pq` (PostgreSQL), `redis/go-redis/v9`
- gRPC: `google.golang.org/grpc` with protobuf definitions in `internal/infrastructure/transport/rpc/proto/`
- Validation: `go-playground/validator`

## Conventions

- Constructor injection everywhere (`NewXxx` functions)
- HTTP handlers read/write JSON via utility functions in `transport/http/utils/`
- Repository auto-creates the `users` table on init
- Passwords hashed with bcrypt at default cost
