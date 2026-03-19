# Engineering Quality Improvements Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** Fix 5 engineering rough edges (config defaults, error responses, validation duplication, crypto tests, metrics exporter).

**Architecture:** 
- Fix config first (enables safe prod deployment)
- Add metrics exporter (enables monitoring)
- Standardize error responses (improves API consistency)
- Consolidate validation (reduces duplication)
- Increase crypto tests (ensures security)

**Tech Stack:** Go 1.25, Prometheus client, testify mocks, slog

---

## Task 1: Fix Config Defaults (Remove env-default for Redis)

**Files:**
- Modify: `internal/config/config.go:28-30`
- Test: `go build ./...` + verify startup requires env vars

**Step 1: Review current config**

```bash
cd /home/prashant-dobariya/engineering/chatty/auth
cat internal/config/config.go | grep -A 5 "type Redis"
```

**Expected output:**
```
type Redis struct {
    Addr string `yaml:"address" env:"REDIS_ADDRESS" env-required:"true" env-default:"localhost:6379"`
    User string `yaml:"username" env:"REDIS_USERNAME" env-required:"true" env-default:""`
    Pass string `yaml:"password" env:"REDIS_PASSWORD" env-required:"true" env-default:""`
}
```

**Step 2: Remove env-default values**

Edit `internal/config/config.go` lines 27-31:

```go
type Redis struct {
	Addr string `yaml:"address" env:"REDIS_ADDRESS" env-required:"true"`
	User string `yaml:"username" env:"REDIS_USERNAME" env-required:"true"`
	Pass string `yaml:"password" env:"REDIS_PASSWORD" env-required:"true"`
}
```

Changes:
- Line 28: Remove `env-default:"localhost:6379"` → Just `env-required:"true"`
- Line 29: Remove `env-default:""` → Just `env-required:"true"`
- Line 30: Remove `env-default:""` → Just `env-required:"true"`

**Step 3: Verify build still works with env vars set**

```bash
cd /home/prashant-dobariya/engineering/chatty/auth
export REDIS_ADDRESS="localhost:6379"
export REDIS_USERNAME=""
export REDIS_PASSWORD=""
go build ./cmd/main
echo $?
```

**Expected:** Exit code 0 (success)

**Step 4: Test that startup fails without env vars**

```bash
cd /home/prashant-dobariya/engineering/chatty/auth
unset REDIS_ADDRESS REDIS_USERNAME REDIS_PASSWORD
timeout 2 go run ./cmd/main -config config/local.yaml 2>&1 | head -20
```

**Expected:** Error mentioning missing REDIS_ADDRESS (fail fast before connecting)

**Step 5: Update example config**

Edit `.env.example`:
```bash
# Redis configuration (REQUIRED - no defaults)
REDIS_ADDRESS=localhost:6379
REDIS_USERNAME=
REDIS_PASSWORD=
```

**Step 6: Commit**

```bash
git add internal/config/config.go .env.example
git commit -m "security: remove env-default for Redis config to prevent prod credential leaks"
```

---

## Task 2: Add /metrics HTTP Endpoint (Prometheus Exporter)

**Files:**
- Create: `internal/infrastructure/transport/http/handler/metrics_handler.go`
- Modify: `cmd/main/main.go` - register metrics handler
- Test: `curl http://localhost:8080/metrics`

**Step 1: Create metrics handler**

Create file `internal/infrastructure/transport/http/handler/metrics_handler.go`:

```go
package handler

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RegisterMetricsHandler registers the Prometheus metrics exporter endpoint
func RegisterMetricsHandler(mux *http.ServeMux) {
	mux.Handle("GET /metrics", promhttp.Handler())
}
```

**Step 2: Register metrics handler in main**

Edit `cmd/main/main.go`. Find where `mux.HandleFunc` is called for other handlers, and add metrics:

Look for pattern:
```go
authHandler := handler.NewAuthHandler(...)
authHandler.RegisterRoutes(mux)
```

Add after all handler registrations:
```go
// Register metrics exporter
handler.RegisterMetricsHandler(mux)
```

**Step 3: Test metrics endpoint manually**

```bash
cd /home/prashant-dobariya/engineering/chatty/auth
make run &  # Start server in background
sleep 2
curl -s http://localhost:8080/metrics | head -20
pkill -f "go run ./cmd/main"
```

**Expected output:**
```
# HELP auth_invites_accepted_total Total number of invites accepted
# TYPE auth_invites_accepted_total counter
auth_invites_accepted_total 0
# HELP auth_invites_revoked_total Total number of invites revoked
# TYPE auth_invites_revoked_total counter
auth_invites_revoked_total 0
# HELP auth_invites_sent_total Total number of invites sent
# TYPE auth_invites_sent_total counter
auth_invites_sent_total 0
...
```

**Step 4: Write integration test**

Create file `internal/infrastructure/transport/http/handler/metrics_handler_test.go`:

```go
package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsHandler_ExposeMetrics(t *testing.T) {
	mux := http.NewServeMux()
	RegisterMetricsHandler(mux)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "# HELP")
	assert.Contains(t, w.Body.String(), "# TYPE")
	assert.Contains(t, w.Body.String(), "auth_")
}

func TestMetricsHandler_ExposesCustomMetrics(t *testing.T) {
	mux := http.NewServeMux()
	RegisterMetricsHandler(mux)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	metricsOutput := w.Body.String()
	customMetrics := []string{
		"auth_invites_sent_total",
		"auth_invites_accepted_total",
		"auth_invites_revoked_total",
		"auth_token_verifications_total",
		"auth_token_verification_duration_seconds",
	}

	for _, metric := range customMetrics {
		assert.True(t, strings.Contains(metricsOutput, metric),
			"Expected metric %s not found in output", metric)
	}
}

func TestMetricsHandler_ReturnsPrometheusFormat(t *testing.T) {
	mux := http.NewServeMux()
	RegisterMetricsHandler(mux)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8", w.Header().Get("Content-Type"))
}
```

**Step 5: Run tests**

```bash
cd /home/prashant-dobariya/engineering/chatty/auth
go test ./internal/infrastructure/transport/http/handler -run TestMetrics -v
```

**Expected:** All 3 tests pass

**Step 6: Verify all tests still pass**

```bash
cd /home/prashant-dobariya/engineering/chatty/auth
go test ./...
```

**Expected:** All 15 packages passing

**Step 7: Commit**

```bash
git add internal/infrastructure/transport/http/handler/metrics_handler.go
git add internal/infrastructure/transport/http/handler/metrics_handler_test.go
git add cmd/main/main.go
git commit -m "feat: add /metrics HTTP endpoint for Prometheus exporter"
```

---

## Task 3: Standardize Error Responses to JSON

**Files:**
- Create: `internal/infrastructure/transport/http/utils/error_response.go`
- Modify: All handlers in `internal/infrastructure/transport/http/handler/*.go`
- Test: Each handler test verifies JSON error format

**Step 1: Create error response helper**

Create file `internal/infrastructure/transport/http/utils/error_response.go`:

```go
package utils

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ErrorCode represents standard error codes
type ErrorCode string

const (
	ErrCodeBadRequest     ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden      ErrorCode = "FORBIDDEN"
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrCodeConflict       ErrorCode = "CONFLICT"
	ErrCodeInternalError  ErrorCode = "INTERNAL_ERROR"
	ErrCodeRateLimited    ErrorCode = "RATE_LIMITED"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

// SendErrorResponse writes a JSON error response
func SendErrorResponse(w http.ResponseWriter, statusCode int, code ErrorCode, message string) {
	SendErrorResponseWithDetails(w, statusCode, code, message, "")
}

// SendErrorResponseWithDetails writes a JSON error response with additional details
func SendErrorResponseWithDetails(w http.ResponseWriter, statusCode int, code ErrorCode, message string, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to write error response", slog.Any("error", err))
	}
}

// MustParseBody parses JSON body and returns error if invalid
func MustParseBody(w http.ResponseWriter, r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		SendErrorResponse(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return err
	}
	return nil
}
```

**Step 2: Update auth handler to use error responses**

Edit `internal/infrastructure/transport/http/handler/auth_handler.go`. Find all `http.Error` calls and replace:

Before:
```go
http.Error(w, "invalid email format", http.StatusBadRequest)
```

After:
```go
utils.SendErrorResponse(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid email format")
```

Replace all occurrences in auth_handler.go (approximately 10-15 error calls).

Add import at top:
```go
"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils"
```

**Step 3: Update team handler to use error responses**

Edit `internal/infrastructure/transport/http/handler/team_handler.go`. Same pattern as auth handler.

Replace all `http.Error` calls with `utils.SendErrorResponse`.

**Step 4: Update health handler to use error responses**

Edit `internal/infrastructure/transport/http/handler/health_handler.go` (if it exists) or any other handlers.

Replace all `http.Error` calls.

**Step 5: Write test for error response format**

Create file `internal/infrastructure/transport/http/utils/error_response_test.go`:

```go
package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendErrorResponse_ReturnsJSON(t *testing.T) {
	w := httptest.NewRecorder()
	SendErrorResponse(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid input")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, ErrCodeBadRequest, resp.Code)
	assert.Equal(t, "Invalid input", resp.Message)
}

func TestSendErrorResponseWithDetails_IncludesDetails(t *testing.T) {
	w := httptest.NewRecorder()
	SendErrorResponseWithDetails(w, http.StatusInternalServerError, ErrCodeInternalError,
		"Failed to process request", "database connection timeout")

	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "database connection timeout", resp.Details)
}

func TestErrorCodes_AreValid(t *testing.T) {
	codes := []ErrorCode{
		ErrCodeBadRequest,
		ErrCodeUnauthorized,
		ErrCodeForbidden,
		ErrCodeNotFound,
		ErrCodeConflict,
		ErrCodeInternalError,
		ErrCodeRateLimited,
	}

	for _, code := range codes {
		assert.NotEmpty(t, code)
	}
}
```

**Step 6: Run tests to verify error handling**

```bash
cd /home/prashant-dobariya/engineering/chatty/auth
go test ./internal/infrastructure/transport/http/utils -v
go test ./internal/infrastructure/transport/http/handler -v -run "Auth|Team" | head -50
```

**Expected:** All tests pass with consistent JSON error responses

**Step 7: Verify all tests still pass**

```bash
cd /home/prashant-dobariya/engineering/chatty/auth
go test ./...
```

**Expected:** All 15 packages passing

**Step 8: Commit**

```bash
git add internal/infrastructure/transport/http/utils/error_response.go
git add internal/infrastructure/transport/http/utils/error_response_test.go
git add internal/infrastructure/transport/http/handler/auth_handler.go
git add internal/infrastructure/transport/http/handler/team_handler.go
git add internal/infrastructure/transport/http/handler/health_handler.go
git commit -m "feat: standardize error responses to JSON format across all handlers"
```

---

## Task 4: Consolidate Validation Logic (Move to Usecase)

**Files:**
- Modify: `internal/usecase/team_usecase.go` - add validation methods
- Modify: `internal/infrastructure/transport/http/handler/team_handler.go` - use usecase validation
- Test: `internal/usecase/team_usecase_validation_test.go` (new)

**Step 1: Extract validation methods to usecase**

Edit `internal/usecase/team_usecase.go`. Add validation methods to `teamUsecase`:

```go
// ValidateInviteRequest validates the invite request
func (t *teamUsecase) validateInviteEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if !isValidEmail(email) {
		return errors.New("invalid email format")
	}
	return nil
}

// ValidateRole validates role is in acceptable range
func (t *teamUsecase) validateRole(role int) error {
	if role < 1 || role > 4 {
		return errors.New("role must be between 1 and 4")
	}
	return nil
}

// Helper function (add at package level if not exists)
func isValidEmail(email string) bool {
	// Simple email validation (already exists in pkg/validator)
	// For now, check basic format
	if !strings.Contains(email, "@") {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	return true
}
```

Add to `TeamUsecase` interface:
```go
type TeamUsecase interface {
	InviteUser(ctx context.Context, businessID int64, email string, role int) (string, error)
	AcceptInvitation(ctx context.Context, inviteToken string) error
	RevokeInvitation(ctx context.Context, inviteToken string) error
	ListMembers(ctx context.Context, businessID int64) ([]*entity.BusinessMember, error)
	RemoveMember(ctx context.Context, businessID int64, memberID int64) error
	UpdateMemberRole(ctx context.Context, businessID int64, memberID int64, newRole int) error
	
	// Validation methods (NEW)
	ValidateInviteEmail(email string) error
	ValidateRole(role int) error
}
```

**Step 2: Update handler to use usecase validation**

Edit `internal/infrastructure/transport/http/handler/team_handler.go`. Replace inline validation:

Before:
```go
func (h *TeamHandler) invite(w http.ResponseWriter, r *http.Request) {
	var req inviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	// Validate email
	if req.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}
	if !isValidEmail(req.Email) {
		http.Error(w, "invalid email format", http.StatusBadRequest)
		return
	}
	// Validate role (1-4)
	if req.Role < 1 || req.Role > 4 {
		http.Error(w, "role must be between 1 and 4", http.StatusBadRequest)
		return
	}
```

After:
```go
func (h *TeamHandler) invite(w http.ResponseWriter, r *http.Request) {
	var req inviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}
	
	// Use usecase validation
	if err := h.UC.ValidateInviteEmail(req.Email); err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}
	if err := h.UC.ValidateRole(req.Role); err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}
```

**Step 3: Write validation tests**

Create file `internal/usecase/team_usecase_validation_test.go`:

```go
package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/Prashant2307200/auth-service/internal/testutil"
)

func TestTeamUsecase_ValidateInviteEmail_ValidEmail(t *testing.T) {
	tu := NewTeamUsecase(&testutil.MockMemberRepository{}, &testutil.MockAuditRepository{}, nil, nil)

	err := tu.ValidateInviteEmail("user@example.com")
	assert.NoError(t, err)
}

func TestTeamUsecase_ValidateInviteEmail_EmptyEmail(t *testing.T) {
	tu := NewTeamUsecase(&testutil.MockMemberRepository{}, &testutil.MockAuditRepository{}, nil, nil)

	err := tu.ValidateInviteEmail("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestTeamUsecase_ValidateInviteEmail_InvalidFormat(t *testing.T) {
	tu := NewTeamUsecase(&testutil.MockMemberRepository{}, &testutil.MockAuditRepository{}, nil, nil)

	testCases := []string{
		"notanemail",
		"@example.com",
		"user@",
		"user@@example.com",
	}

	for _, email := range testCases {
		err := tu.ValidateInviteEmail(email)
		assert.Error(t, err, "Expected error for email: %s", email)
	}
}

func TestTeamUsecase_ValidateRole_ValidRoles(t *testing.T) {
	tu := NewTeamUsecase(&testutil.MockMemberRepository{}, &testutil.MockAuditRepository{}, nil, nil)

	for role := 1; role <= 4; role++ {
		err := tu.ValidateRole(role)
		assert.NoError(t, err, "Role %d should be valid", role)
	}
}

func TestTeamUsecase_ValidateRole_InvalidRoles(t *testing.T) {
	tu := NewTeamUsecase(&testutil.MockMemberRepository{}, &testutil.MockAuditRepository{}, nil, nil)

	invalidRoles := []int{0, -1, 5, 100}
	for _, role := range invalidRoles {
		err := tu.ValidateRole(role)
		assert.Error(t, err, "Role %d should be invalid", role)
	}
}
```

**Step 4: Run validation tests**

```bash
cd /home/prashant-dobariya/engineering/chatty/auth
go test ./internal/usecase -run "Validate" -v
```

**Expected:** All validation tests pass

**Step 5: Verify all tests still pass**

```bash
cd /home/prashant-dobariya/engineering/chatty/auth
go test ./...
```

**Expected:** All 15 packages passing

**Step 6: Commit**

```bash
git add internal/usecase/team_usecase.go
git add internal/usecase/team_usecase_validation_test.go
git add internal/infrastructure/transport/http/handler/team_handler.go
git commit -m "refactor: consolidate validation logic to usecase layer (DRY principle)"
```

---

## Task 5: Increase Test Coverage for pkg/hash (Crypto Package)

**Files:**
- Review: `pkg/hash/hash.go`, `pkg/hash/hash_utils.go`
- Modify/Create: `pkg/hash/hash_test.go`, `pkg/hash/hash_utils_test.go`
- Target: 95%+ coverage

**Step 1: Review current hash package**

```bash
cd /home/prashant-dobariya/engineering/chatty/auth
cat pkg/hash/hash.go
cat pkg/hash/hash_utils.go
cat pkg/hash/hash_test.go
```

**Step 2: Check current coverage**

```bash
cd /home/prashant-dobariya/engineering/chatty/auth
go test ./pkg/hash -cover
```

**Expected output:**
```
coverage: 33.3% of statements
```

**Step 3: Write comprehensive hash tests**

Edit or create `pkg/hash/hash_test.go` with complete coverage:

```go
package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword_CreatesValidHash(t *testing.T) {
	password := "securePassword123!"
	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestHashPassword_CreatesUniqueHashes(t *testing.T) {
	password := "samePassword"
	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	assert.NotEqual(t, hash1, hash2, "Same password should produce different hashes (bcrypt uses random salt)")
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	hash, err := HashPassword("")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestHashPassword_LongPassword(t *testing.T) {
	password := string(make([]byte, 100)) // 100 byte password
	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestVerifyPassword_CorrectPassword(t *testing.T) {
	password := "mySecurePassword123"
	hash, _ := HashPassword(password)

	err := VerifyPassword(hash, password)
	assert.NoError(t, err)
}

func TestVerifyPassword_IncorrectPassword(t *testing.T) {
	password := "correctPassword"
	hash, _ := HashPassword(password)

	err := VerifyPassword(hash, "wrongPassword")
	assert.Error(t, err)
	assert.Equal(t, bcrypt.ErrMismatchedHashAndPassword, err)
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	err := VerifyPassword("notavalidhash", "password")
	assert.Error(t, err)
}

func TestVerifyPassword_EmptyHash(t *testing.T) {
	err := VerifyPassword("", "password")
	assert.Error(t, err)
}

func TestVerifyPassword_EmptyPassword(t *testing.T) {
	hash, _ := HashPassword("")
	err := VerifyPassword(hash, "")
	assert.NoError(t, err)
}

func TestVerifyPassword_SensitiveToCase(t *testing.T) {
	password := "MyPassword"
	hash, _ := HashPassword(password)

	err := VerifyPassword(hash, "mypassword")
	assert.Error(t, err)
}

func TestHashPassword_ErrorOnExtremelyLongPassword(t *testing.T) {
	// Bcrypt has a 72-byte limit
	password := string(make([]byte, 73))
	hash, err := HashPassword(password)

	// Bcrypt will truncate at 72 bytes but still succeed
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestHashAndVerifyRoundtrip(t *testing.T) {
	testCases := []string{
		"simple",
		"with spaces",
		"with-special-chars!@#$%",
		"unicode-emoji-😀",
		"veryLongPasswordWithManyCharactersAndNumbers123456789",
	}

	for _, password := range testCases {
		hash, err := HashPassword(password)
		assert.NoError(t, err)

		err = VerifyPassword(hash, password)
		assert.NoError(t, err, "Failed for password: %s", password)
	}
}

func TestVerifyPassword_DoesNotEqual(t *testing.T) {
	// Ensure we're actually comparing hashes, not doing string comparison
	password := "test"
	hash, _ := HashPassword(password)

	// Hash should not equal password
	assert.NotEqual(t, hash, password)
	
	// But verification should still work
	assert.NoError(t, VerifyPassword(hash, password))
}

func TestHashPassword_CostFactor(t *testing.T) {
	password := "test"
	hash, _ := HashPassword(password)

	// Extract cost from hash and verify it's reasonable (10-14)
	cost, err := bcrypt.Cost([]byte(hash))
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, cost, 10)
	assert.LessOrEqual(t, cost, 14)
}
```

**Step 4: Write hash_utils tests (if hash_utils.go has functions)**

Check if `pkg/hash/hash_utils.go` exists and has functions:

```bash
cd /home/prashant-dobariya/engineering/chatty/auth
cat pkg/hash/hash_utils.go
```

If it has functions, create `pkg/hash/hash_utils_test.go` with similar comprehensive coverage.

**Step 5: Run hash tests and verify coverage**

```bash
cd /home/prashant-dobariya/engineering/chatty/auth
go test ./pkg/hash -v
go test ./pkg/hash -cover
```

**Expected output:**
```
coverage: 95%+ of statements
```

**Step 6: Verify all tests still pass**

```bash
cd /home/prashant-dobariya/engineering/chatty/auth
go test ./...
```

**Expected:** All 15 packages passing

**Step 7: Commit**

```bash
git add pkg/hash/hash_test.go
git add pkg/hash/hash_utils_test.go
git commit -m "test: increase pkg/hash coverage to 95%+ for crypto security"
```

---

## Summary

**Total effort:** ~6-7 hours spread across 5 tasks

**Execution order:**
1. Task 1: Config defaults (30 mins) — CRITICAL for prod safety
2. Task 2: Metrics endpoint (1 hour) — Enable monitoring
3. Task 3: Error responses (2 hours) — API consistency
4. Task 4: Validation consolidation (1.5 hours) — DRY principle
5. Task 5: Crypto tests (1.5 hours) — Security confidence

**Post-implementation:**
- All 15 test packages should pass
- Coverage: 80%+ across codebase
- Build: `go build ./...` clean, no warnings
- Ready for production deployment

**Commits will be:**
1. security: remove env-default for Redis config
2. feat: add /metrics HTTP endpoint for Prometheus exporter
3. feat: standardize error responses to JSON format
4. refactor: consolidate validation logic to usecase layer
5. test: increase pkg/hash coverage to 95%+

---

## Unresolved Questions

None — plan is complete and ready for implementation.
