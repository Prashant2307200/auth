# Auth Service — Operations & Deployment Runbook

## Prerequisites

| Dependency | Version | Purpose |
|------------|---------|---------|
| Go | 1.25+ | Build & run |
| PostgreSQL | 15+ | Primary datastore |
| Redis | 7+ | Refresh token storage (7-day TTL) |
| Cloudinary | — | Profile image uploads |
| Docker | 24+ | Container deployment (optional) |

RSA key pair required at `keys/private.pem` and `keys/public.pem` (relative to working directory).

Generate keys:
```bash
openssl genrsa -out keys/private.pem 2048
openssl rsa -in keys/private.pem -pubout -out keys/public.pem
```

---

## Configuration

Config is loaded from a YAML file. Path is set via `CONFIG_PATH` env var or `--config` flag.

```bash
CONFIG_PATH=config/local.yaml ./server
# or
./server --config=config/local.yaml
```

Copy the example config: `cp config/local.yaml.example config/local.yaml`

### Environment Variables

All secrets **must** be set via environment variables (they override YAML values).

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `ENV` | Yes | Runtime environment | `dev` / `staging` / `production` |
| `POSTGRES_URI` | Yes | PostgreSQL connection string | `postgresql://user:pass@localhost:5432/auth_db` |
| `REDIS_ADDRESS` | Yes | Redis host:port | `localhost:6379` |
| `REDIS_USERNAME` | Yes | Redis username (empty if none) | `` |
| `REDIS_PASSWORD` | Yes | Redis password (empty if none) | `` |
| `ACCESS_TOKEN_SECRET` | Yes | Secret for access token signing | `<random-64-byte-base64>` |
| `REFRESH_TOKEN_SECRET` | Yes | Secret for refresh token signing | `<random-64-byte-base64>` |
| `COOKIE_SECRET` | Yes | Secret for cookie signing | `<random-64-byte-base64>` |
| `NAME` | Yes | Cloudinary cloud name | `my-cloud` |
| `API_KEY` | Yes | Cloudinary API key | `123456789012345` |
| `API_SECRET` | Yes | Cloudinary API secret | `<cloudinary-secret>` |
| `CONFIG_PATH` | Yes* | Path to YAML config file | `config/local.yaml` |
| `SEED_ON_STARTUP` | No | Force DB seeding (auto-enabled in `dev`) | `true` |

> `http_server.address` is set in YAML only (e.g. `:8080`). No env var override.

Generate secrets:
```bash
openssl rand -base64 64
```

---

## Local Development

### Start infrastructure
```bash
docker-compose up -d postgres redis
```

### Run the service
```bash
make run                          # uses config/local.yaml
CONFIG_PATH=config/local.yaml make run
```

### Auto-reload (requires `air`)
```bash
make install-tools   # installs air
make dev
```

### Run tests
```bash
make test                # all tests, verbose
make test-short          # skip slow tests
make test-coverage       # generates coverage.html
```

### Code quality
```bash
make fmt    # go fmt ./...
make vet    # go vet ./...
make lint   # golangci-lint (requires install-tools)
make check  # fmt + vet + test
```

---

## Deployment

### Binary build
```bash
make build                        # outputs bin/server
make build-linux                  # cross-compile for Linux/Docker
go build -o auth-service ./cmd/main/
```

### Docker
```bash
# Build image
make docker-build                 # tags as auth-service:latest
docker build -t auth-service .

# Run container (supply all required env vars)
docker run -d \
  -p 8080:8080 \
  -e ENV=production \
  -e POSTGRES_URI="postgresql://..." \
  -e REDIS_ADDRESS="redis:6379" \
  -e REDIS_USERNAME="" \
  -e REDIS_PASSWORD="" \
  -e ACCESS_TOKEN_SECRET="..." \
  -e REFRESH_TOKEN_SECRET="..." \
  -e COOKIE_SECRET="..." \
  -e NAME="..." \
  -e API_KEY="..." \
  -e API_SECRET="..." \
  -e CONFIG_PATH="/config/production.yaml" \
  -v /path/to/config:/config \
  -v /path/to/keys:/keys \
  auth-service
```

### Docker Compose (full stack)
```bash
docker-compose up -d
```
Starts: PostgreSQL 15, Redis 7, MailHog (SMTP dev), and the app on `:8080`.

### Startup sequence
1. Load config (fatal if missing)
2. Connect PostgreSQL + run migrations (fatal on failure)
3. Connect Redis (fatal on failure)
4. Seed DB (only in `dev` or `SEED_ON_STARTUP=true`)
5. Start HTTP server on configured address
6. Start gRPC server on `:9090`
7. Graceful shutdown on `SIGINT`/`SIGTERM` (5s timeout)

---

## Ports

| Port | Protocol | Purpose |
|------|----------|---------|
| `8080` | HTTP | REST API (configurable via YAML) |
| `9090` | gRPC | Token verification & public key distribution |

---

## API Routes

| Prefix | Auth Required | Rate Limited |
|--------|--------------|--------------|
| `/api/v1/auth/register` | No | Yes — 5 req/min |
| `/api/v1/auth/login` | No | Yes — 5 req/min |
| `/api/v1/auth/*` | No | No |
| `/api/v1/users/*` | Yes | No |
| `/api/v1/business/*` | Yes | No |
| `/health`, `/health/live`, `/health/ready` | No | No |
| `/metrics` | No | No |

---

## Monitoring & Observability

### Health endpoints
```bash
GET /health          # application-level health
GET /health/live     # liveness: always 200 if process is up
GET /health/ready    # readiness: checks DB + Redis connectivity
```

### Prometheus metrics
```bash
GET /metrics         # Prometheus text format
```

Custom metrics exposed:

| Metric | Type | Description |
|--------|------|-------------|
| `auth_token_verifications_total` | Counter | Total token verification attempts |
| `auth_token_verification_duration_seconds` | Histogram | Token verification latency |
| `auth_invites_sent_total` | Counter | Total invites sent |
| `auth_invites_accepted_total` | Counter | Total invites accepted |
| `auth_invites_revoked_total` | Counter | Total invites revoked |

### Grafana dashboard suggestions
- **Auth rate**: `rate(auth_token_verifications_total[5m])`
- **Token latency p99**: `histogram_quantile(0.99, auth_token_verification_duration_seconds_bucket)`
- **Invite funnel**: sent vs accepted vs revoked counters
- **Error rate**: HTTP 4xx/5xx from standard Go metrics
- **Readiness**: alert on `/health/ready` returning non-200

### Structured logging
All logs use `slog` (JSON-compatible). Each request gets a unique `request_id` via middleware. Log level is not configurable at runtime — rebuild to change.

---

## Security Considerations

- **JWT**: RSA-2048 key pair. Private key signs tokens; public key verifies. Public key is distributed to other services via gRPC.
- **Cookies**: `HttpOnly`, `SameSite=Lax`. `Secure` flag is set in non-`dev` environments.
- **Rate limiting**: Token bucket — 0.083 rps (≈5/min) on `/register` and `/login`. `Retry-After: 60` header returned on 429.
- **Secrets**: All secrets via env vars. No hardcoded defaults. Service will not start if required vars are missing.
- **SQL**: Parameterized queries throughout — no string interpolation.
- **Headers**: Security headers middleware applied to all responses.

---

## Troubleshooting

### Service won't start — config error
```
Config path is not set
```
**Fix**: Set `CONFIG_PATH` env var or pass `--config=<path>` flag.

### Service won't start — config file missing
```
Config file does not exist: config/local.yaml
```
**Fix**: `cp config/local.yaml.example config/local.yaml` and fill in values.

### Database connection failure
```
Failed to initialize the storage: ...
```
**Check**:
- `POSTGRES_URI` is correct and reachable
- PostgreSQL is running: `docker-compose up -d postgres`
- DB user has CREATE TABLE permissions (needed for migrations)

### Redis connection failure
```
Failed to initialize the cache: ...
```
**Check**:
- `REDIS_ADDRESS` is correct (format: `host:port`)
- Redis is running: `docker-compose up -d redis`
- `REDIS_USERNAME`/`REDIS_PASSWORD` match Redis ACL config

### JWT key errors
```
Failed to initialize token service: ...
```
**Check**:
- `keys/private.pem` and `keys/public.pem` exist relative to the working directory
- Keys are valid RSA PEM format (generate with `openssl genrsa`)
- File permissions allow the process to read them

### Rate limiting — 429 responses
Clients hitting `/register` or `/login` more than 5 times/minute will receive HTTP 429 with `Retry-After: 60`. This is per-IP. Stale IP entries are cleaned up every 5 minutes (entries older than 1 hour are removed).

### DB seeding runs unexpectedly in production
Seeding runs automatically when `ENV=dev` or `SEED_ON_STARTUP=true`. Ensure `ENV` is set to `production` or `staging` in non-dev deployments.

### gRPC port conflict
```
Failed to listen on gRPC port: ...
```
**Fix**: Ensure port `9090` is not in use. gRPC port is not configurable without code change.
