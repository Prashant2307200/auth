# Auth Service

Go service for authentication, users, businesses, and team invitations. HTTP REST on `:8080` (configurable) and gRPC on `:9090`.

## Quick start (local)

1. **Dependencies:** Go 1.25+, PostgreSQL 16+, Redis 7+, RSA key pair, [Cloudinary](https://cloudinary.com/) account.

2. **Keys:**

   ```bash
   openssl genrsa -out keys/private.pem 2048
   openssl rsa -in keys/private.pem -pubout -out keys/public.pem
   ```

3. **Config:** Copy and edit YAML, then export env vars (see [`.env.example`](.env.example)):

   ```bash
   cp config/local.yaml.example config/local.yaml
   # Edit config/local.yaml; export POSTGRES_URI, REDIS_*, REFRESH_TOKEN_SECRET, NAME, API_KEY, API_SECRET, CONFIG_PATH
   ```

4. **Run:**

   ```bash
   make validate-config
   make run
   ```

5. **OpenAPI:** [`api/openapi.yaml`](api/openapi.yaml)

## Docker Compose

```bash
# Set REFRESH_TOKEN_SECRET, NAME, API_KEY, API_SECRET in environment or .env
export REFRESH_TOKEN_SECRET=... NAME=... API_KEY=... API_SECRET=...
make compose-up
```

Uses [`compose.yaml`](compose.yaml) and [`config/docker.yaml`](config/docker.yaml). Mounts `./keys` and `./config`.

## Documentation

| Doc | Purpose |
|-----|---------|
| [docs/OPERATIONS.md](docs/OPERATIONS.md) | Deploy, env vars, health, metrics |
| [docs/api-reference.md](docs/api-reference.md) | API overview |
| [docs/SAAS_AUTH_SERVICE_ASSESSMENT.md](docs/SAAS_AUTH_SERVICE_ASSESSMENT.md) | Product/engineering assessment |
| [docs/CONTRIBUTING_TESTING.md](docs/CONTRIBUTING_TESTING.md) | Tests and conventions |

## Makefile highlights

- `make test` — unit tests
- `make integration-test` — tests with name `Integration`
- `make proto` — regenerate gRPC/protobuf (requires `protoc`)
- `make compose-up` / `make compose-down` — Docker stack

## Team API

`POST /api/v1/team/invite` (and related routes) require **`X-Tenant-ID`** set to the business ID, plus a valid session (cookie or bearer).

## gRPC

Services: `authgrpc.TokenService` / `authgrpc.PublicKeyService` on port **9090** (see `internal/transport/grpc/proto`).
