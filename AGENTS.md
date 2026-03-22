## Learned Workspace Facts

### Auth Service Codebase Structure
- Clean Architecture: `cmd/main` (entry), `internal/usecase`, `internal/infrastructure/transport/http` (handlers, middleware), `testutil/` (mocks)
- `pkg/` for reusable libs (hash, validator, invitetoken, ratelimit, db, rdb); add only if usable outside this module
- `config/` YAML per environment — start from [`config/local.yaml.example`](config/local.yaml.example); Docker uses [`config/docker.yaml`](config/docker.yaml)
- Product/engineering assessment: [`docs/SAAS_AUTH_SERVICE_ASSESSMENT.md`](docs/SAAS_AUTH_SERVICE_ASSESSMENT.md)
- Human onboarding: [`README.md`](README.md)

### Authentication & Security Implementation
- RSA JWT access tokens (RS256); refresh tokens HS256 + Redis (7-day TTL)
- HttpOnly cookies (`access_token`, `refresh_token`); `Secure` when `ENV` != `dev`, `SameSite=Lax`
- Team routes under `/api/v1/team/*` require header **`X-Tenant-ID`** (business ID); see `middleware.TenantFromHeader`
- Rate limit ~5/min on register/login
- MFA/TOTP support via `github.com/pquerna/otp`; backup codes hashed with SHA256
- Google SSO via `golang.org/x/oauth2`; links GoogleID to existing users or creates new
- Multi-session management: Redis stores `session:{id}` JSON + `user_sessions:{userID}` SET

### Password & Email Flows
- Password reset: token hashed (SHA256), 1h expiry, stored in `password_reset_tokens` table
- Email verification: `email_verification_tokens` table; user fields `email_verified`, `email_verified_at`
- Email service interface: `SendInvite`, `SendPasswordReset`, `SendEmailVerification`; default impl is `NoopEmailService`
- Maileroo integration: `internal/service/maileroo.go` — POST to `smtp.maileroo.com/api/v2/emails`

### Code Quality Standards
- `gofmt`, `go vet`, `golangci-lint` in CI
- Required env vars: see [`internal/config/config.go`](internal/config/config.go) and [`.env.example`](.env.example)
- Structured logging: `slog` + request IDs

### Testing & Coverage
- `*_test.go`, `*_integration_test.go`; CI coverage floor (~65%) + integration job (Postgres + Redis)
- `make integration-test` runs tests matching `Integration`

### Infrastructure
- HTTP default `:8080` (YAML); gRPC **:9090** (`TokenService`, `PublicKeyService`)
- Health: `/health`, `/health/live`, `/health/ready` (503 when DB/Redis down on ready)
- Metrics: `/metrics` (Prometheus)
- Compose: [`compose.yaml`](compose.yaml) — `make compose-up`

### Audit Logging
- Expanded action types: `user.login`, `user.logout`, `user.mfa_enabled`, `user.session_revoked`, etc.
- API endpoint: `GET /api/v1/audit-logs` with filtering (user_id, action, from, to) and pagination

### Maintainability
- Validation: `pkg/validator` + usecase-owned rules
- Canonical OpenAPI: `api/openapi.yaml` — `docs/openapi.yaml` is deprecated reference only
