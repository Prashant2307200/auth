## Learned Workspace Facts

### Auth Service Codebase Structure
- Clean Architecture pattern: `cmd/main` (entry), `internal/usecase` (business logic), `infrastructure/transport/http` (handlers), `testutil/` (shared test mocks)
- `pkg/` contains reusable libraries like hash and token helpers; only add packages here if consumable outside module context
- `config/` stores YAML per environment; `config/local.yaml` is default for `make run`

### Authentication & Security Implementation
- Uses RSA key-based JWT with public/private key separation for access tokens
- Bcrypt password hashing with configurable cost factor
- Token storage: HttpOnly cookies with Secure flag (not in dev) and SameSite=Lax
- Parameterized queries for SQL injection protection
- Refresh tokens stored in Redis with 7-day TTL

### Code Quality Standards
- Formatting: `go fmt` must pass cleanly
- Linting: `go vet` has no warnings; no golangci-lint configured
- No hardcoded defaults for secrets (security risk if present)
- Structured logging uses `slog` package with request IDs via middleware

### Critical Issues Found
- Hardcoded secret defaults in `internal/config/config.go` must be removed—only use env vars
- ✅ FIXED: `TokenService` interface now includes `VerifyToken(ctx, token) (int64, error)` method; updated mock
- ✅ FIXED: Rate limiter cleanup runs every 5 minutes to prevent memory leak from stale IP entries
- ✅ FIXED: Context timeout in auth middleware increased from 1s to 5s for reliable DB operations

### Testing & Coverage
- Many test files exist (`*_test.go`) but test execution not fully visible
- Critical paths (auth flow, token validation) need higher coverage
- No explicit integration tests

### Infrastructure Components
- Postgres with connection pooling configured
- Redis for session/refresh token storage
- Cloudinary for image uploads
- Rate limiting: 0.083 rps (5/min) on register/login endpoints

### Maintainability Issues
- Validation logic duplicated: happens in both handler and usecase layers (consolidate to one)
- Response format inconsistent: `Response` struct has "Error" field used as message, `SuccessResponse` uses proper "message"
- ProfilePic field not validated before database insert
- Error handling mixes `log.Fatal` (cmd/main) with `slog` (rest of codebase)
