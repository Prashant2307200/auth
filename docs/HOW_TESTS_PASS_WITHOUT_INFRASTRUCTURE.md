# How Tests Pass Without Real Infrastructure

## The Secret: Mock Objects (Dependency Injection + Testify)

When we write tests, we **replace real repositories, databases, and services with MOCK objects** that simulate their behavior without touching actual infrastructure.

---

## Concrete Example: Token Revocation Test

### What the test does:

```go
// Line 16-17: Create MOCK repositories (not real DB)
memberRepo := &testutil.MockMemberRepo{}
auditRepo := &testutil.MockAuditRepo{}

// Line 28: Tell the mock: "When GetByInviteToken is called with 'valid-token',
//          return this fake member object (no DB query)"
memberRepo.On("GetByInviteToken", mock.Anything, "valid-token").Return(member, nil)

// Line 39: Create usecase with MOCK repos (not real database)
tu := &teamUsecase{memberRepo: memberRepo, auditRepo: auditRepo, tokenGen: tokenGen}

// Line 41: Call the actual business logic
err := tu.RevokeInvitation(context.Background(), "valid-token")

// Line 42-44: Verify the mock was called correctly
assert.NoError(t, err)
memberRepo.AssertExpectations(t)  // Verify GetByInviteToken was called
auditRepo.AssertExpectations(t)   // Verify Log was called
```

---

## Diagram: How Mocking Works

```
┌─────────────────────────────────────────────────────────────────┐
│                     Unit Test (No Infrastructure)                │
└─────────────────────────────────────────────────────────────────┘

TEST CODE:
┌──────────────────────────────────────┐
│  TestRevokeInvitation                │
│                                      │
│  1. Create MockMemberRepo            │
│  2. Configure mock behavior:         │
│     "If you're asked for token X,    │
│      return this fake member"        │
│  3. Call real RevokeInvitation()     │
│  4. Verify mock was called correctly │
└──────────────────────────────────────┘
          ↓
    USECASE LAYER (REAL):
    ┌──────────────────────────────────────┐
    │  teamUsecase.RevokeInvitation()      │
    │  (This is REAL business logic)       │
    │                                      │
    │  1. call memberRepo.GetByInviteToken │
    │  2. Validate status                  │
    │  3. Update member status to revoked  │
    │  4. Log to auditRepo                 │
    └──────────────────────────────────────┘
          ↓
    DEPENDENCY LAYER (MOCK):
    ┌──────────────────────────────────────┐
    │  MockMemberRepo.GetByInviteToken()   │
    │  ← Returns pre-configured fake data  │
    │  (NOT hitting real PostgreSQL DB)    │
    │                                      │
    │  MockMemberRepo.Update()             │
    │  ← Simulates DB save (instant)       │
    │  (NOT hitting real PostgreSQL DB)    │
    │                                      │
    │  MockAuditRepo.Log()                 │
    │  ← Simulates audit log (instant)     │
    │  (NOT hitting real audit system)     │
    └──────────────────────────────────────┘

RESULT: ✅ Test passes without touching PostgreSQL, Redis, or any external service
```

---

## What's NOT Being Tested

Tests DON'T test:

- ❌ Database connectivity (no PostgreSQL running)
- ❌ Redis functionality (no Redis server)
- ❌ Cloudinary image uploads (no cloud service)
- ❌ Email sending (no SMTP service)
- ❌ Network latency or timeouts
- ❌ Authentication credentials

Tests DO test:

- ✅ Business logic correctness (RevokeInvitation updates status correctly)
- ✅ Error handling (what happens if token not found)
- ✅ Function calls (was Update() called with correct data)
- ✅ Happy path and edge cases

---

## Why This Works: Interfaces

The key is that repositories are **interfaces**, not concrete PostgreSQL code:

### In production (real code):

```go
// real_usecase.go
type teamUsecase struct {
    memberRepo repository.MemberRepository  // ← This is an INTERFACE
    auditRepo  repository.AuditRepository   // ← This is an INTERFACE
}

// When app starts, we inject real implementations:
postgres := &postgres.MemberRepository{}  // Talks to real DB
team := NewTeamUsecase(postgres, ...)     // Real DB injection
```

### In tests (mock code):

```go
// team_usecase_revoke_test.go
teamUsecase struct {
    memberRepo repository.MemberRepository  // ← Same INTERFACE
    auditRepo  repository.AuditRepository   // ← Same INTERFACE
}

// In tests, we inject mock implementations:
mock := &testutil.MockMemberRepo{}  // Fake, returns pre-set data
team := &teamUsecase{memberRepo: mock, ...}  // Mock injection
```

**Same interface, different implementations!**

---

## Testify Mock Library (What Powers This)

We use `github.com/stretchr/testify/mock` which provides:

### 1. Recording expectations:
```go
memberRepo.On("GetByInviteToken", mock.Anything, "token").Return(member, nil)
// "When GetByInviteToken is called with any context and 'token',
//  return this member object and nil error"
```

### 2. Verifying calls:
```go
memberRepo.AssertExpectations(t)
// "Check: Was GetByInviteToken actually called?"
// "Check: Was Update actually called with correct data?"
```

### 3. Custom matchers:
```go
auditRepo.On("Log", mock.Anything, mock.MatchedBy(func(al *entity.AuditLog) bool {
    return al.Action == "revoke_invitation"
})).Return(nil)
// "Accept any AuditLog where Action is 'revoke_invitation'"
```

---

## Real Infrastructure Tests (Integration Tests)

Our code ALSO has integration tests that use REAL infrastructure:

```bash
# Tests that use real DB/Redis (if configured):
internal/seeder/seeder_test.go          # Real PostgreSQL
internal/infrastructure/repository/*    # Real PostgreSQL
internal/infrastructure/transport/http  # Real HTTP handlers
```

These tests have **different setup** (database fixtures, containers):

```go
// integration test example
func TestUserRepository_CreateAndRetrieve(t *testing.T) {
    // Setup: Real PostgreSQL connection
    db := setupTestDatabase(t)  // Actually connects to test DB
    defer db.Close()
    
    repo := postgres.NewUserRepository(db)
    
    // This DOES query real database
    user, err := repo.GetById(ctx, 1)
    // ...
}
```

---

## Test Execution Flow (What We Just Did)

```bash
$ go test ./...

Running test: TestRevokeInvitation_SuccessfullyRevokes
  1. Create empty MockMemberRepo (0 DB calls)
  2. Configure: "GetByInviteToken('token') → returns member"
  3. Call real: tu.RevokeInvitation(ctx, "token")
  4. Real code internally calls: mockRepo.GetByInviteToken() → gets fake data
  5. Real code processes business logic
  6. Real code calls: mockRepo.Update() → mock records the call
  7. Verify: Was Update() called with correct data? ✅
  Result: PASS (0ms, no DB connection needed)

Running test: TestRevokeInvitation_TokenNotFound
  1. Create empty MockMemberRepo
  2. Configure: "GetByInviteToken('nonexistent') → return ErrNotFound"
  3. Call real: tu.RevokeInvitation(ctx, "nonexistent")
  4. Real code calls: mockRepo.GetByInviteToken() → gets error
  5. Real code error handling triggers
  6. Function returns error
  7. Verify: err contains "not found" ✅
  Result: PASS (0ms, no DB connection needed)

[... 8 more tests ...]

All tests: PASS in 0.004s
Total: No infrastructure needed, no network calls, blazingly fast ⚡
```

---

## Summary: Why Tests Pass Without Infrastructure

| Component | Test Behavior | Why |
|-----------|---------------|-----|
| PostgreSQL | Not used | `MockMemberRepo` returns fake data instantly |
| Redis | Not used | `MockRedisClient` not configured in tests |
| Cloudinary | Not used | Not tested; handler level only |
| Email Service | Not used | `MockEmailService` returns nil |
| HTTP calls | Not used | `httptest.NewRequest()` simulates requests |
| Database queries | Not used | Mocks intercept all repository calls |

**Key insight:** We test business logic in isolation, separated from infrastructure concerns.

---

## What Tests Would Fail Without This

If we removed the mocks and tried to use REAL repositories:

```go
// ❌ This would FAIL without real infrastructure:
memberRepo := &postgres.MemberRepository{}  // Needs real DB connection
tu := &teamUsecase{memberRepo: memberRepo}  // ← Fails: Can't connect to PostgreSQL

// ✅ This PASSES with mocks:
memberRepo := &testutil.MockMemberRepo{}    // Fake, no connection needed
tu := &teamUsecase{memberRepo: memberRepo}  // ← Works: Mock data is instant
```

---

## How to Run Tests and See This in Action

```bash
# Run just the revocation tests (no infrastructure):
go test ./internal/usecase -v -run RevokeInvitation

# Output shows they pass instantly with 0 infrastructure:
=== RUN   TestRevokeInvitation_SuccessfullyRevokes
--- PASS: TestRevokeInvitation_SuccessfullyRevokes (0.00s)
=== RUN   TestRevokeInvitation_TokenNotFound
--- PASS: TestRevokeInvitation_TokenNotFound (0.00s)
PASS
ok  	github.com/Prashant2307200/auth-service/internal/usecase	0.004s
```

Compare with integration tests that need infrastructure:

```bash
# Tests that actually use PostgreSQL
go test ./internal/infrastructure/repository/postgres -v

# These are slower (need DB):
=== RUN   TestPostgresUserRepository_Create
--- PASS: TestPostgresUserRepository_Create (0.23s)  ← Much slower!
PASS
ok	...repository/postgres	2.1s
```

**Unit tests (mocked) = 0.004s**
**Integration tests (real DB) = 2.1s**

Same code, different test strategy!
