# Testing Guide

## Overview

This project now includes comprehensive unit tests for critical components. The test suite uses [testify](https://github.com/stretchr/testify) for assertions and mocking.

## Test Coverage

Current test coverage: **10.4%** (improving from 0%)

### Coverage by Package

| Package | Coverage | Status |
|---------|----------|--------|
| `pkg/hash` | 100% | ✅ Complete |
| `internal/infrastructure/transport/http/middleware` | 100% | ✅ Complete |
| `internal/usecase` | 52.6% | 🟡 Good |
| Other packages | 0% | ⚠️ Needs tests |

## Running Tests

### Run all tests
```bash
make test
# or
go test ./...
```

### Run tests with coverage
```bash
make test-coverage
# Generates coverage.out and coverage.html
```

### Run specific package tests
```bash
go test ./pkg/hash/... -v
go test ./internal/usecase/... -v
go test ./internal/infrastructure/transport/http/middleware/... -v
```

### Run tests in short mode (skip slow tests)
```bash
make test-short
```

## Test Structure

### Test Files Created

1. **`pkg/hash/hash_test.go`** - Tests for password hashing
   - `TestHashPassword` - Tests password hashing with various inputs
   - `TestHashPassword_UniqueHashes` - Ensures unique hashes for same password
   - `TestCheckPassword` - Tests password verification
   - `TestHashPassword_CheckPassword_RoundTrip` - End-to-end password flow

2. **`internal/usecase/auth_test.go`** - Tests for authentication use cases
   - `TestAuthUseCase_RegisterUser` - User registration scenarios
   - `TestAuthUseCase_LoginUser` - Login scenarios
   - `TestAuthUseCase_LogoutUser` - Logout functionality
   - `TestAuthUseCase_GetAuthUserProfile` - Profile retrieval
   - `TestAuthUseCase_UpdateAuthUserProfile` - Profile updates

3. **`internal/usecase/user_test.go`** - Tests for user management
   - `TestUserUseCase_GetUsers` - List users
   - `TestUserUseCase_GetUserById` - Get user by ID
   - `TestUserUseCase_CreateUser` - Create user
   - `TestUserUseCase_UpdateUserById` - Update user
   - `TestUserUseCase_DeleteUserById` - Delete user
   - `TestUserUseCase_SearchUsers` - Search functionality

4. **`internal/infrastructure/transport/http/middleware/auth_test.go`** - Tests for auth middleware
   - `TestGetUserIDFromContext` - Context extraction
   - `TestAuthenticate` - Authentication middleware with various scenarios

### Test Utilities

**`internal/testutil/`** - Shared test utilities
- `mocks.go` - Mock implementations of interfaces
  - `MockUserRepo` - Mock user repository
  - `MockTokenService` - Mock token service
  - `MockCloudService` - Mock cloud service
- `helpers.go` - Test helper functions
  - `CreateTestUser()` - Create test user entities
  - `CreateTestUserWithID()` - Create user with specific ID
  - `CreateTestUserWithEmail()` - Create user with specific email

## Writing New Tests

### Example: Testing a Use Case

```go
func TestMyUseCase_MyMethod(t *testing.T) {
    tests := []struct {
        name       string
        input      string
        setupMocks func(*testutil.MockUserRepo)
        wantErr    bool
        wantResult string
    }{
        {
            name:  "successful case",
            input: "test",
            setupMocks: func(repo *testutil.MockUserRepo) {
                repo.On("SomeMethod", mock.Anything, "test").Return(result, nil)
            },
            wantErr:    false,
            wantResult: "expected",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repo := new(testutil.MockUserRepo)
            tt.setupMocks(repo)

            uc := NewMyUseCase(repo)
            result, err := uc.MyMethod(context.Background(), tt.input)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.wantResult, result)
            }

            repo.AssertExpectations(t)
        })
    }
}
```

## Test Best Practices

1. **Use table-driven tests** - Makes tests easy to read and extend
2. **Mock external dependencies** - Use `testutil` mocks for repositories/services
3. **Test error cases** - Don't just test happy paths
4. **Use descriptive test names** - `TestFunctionName_ScenarioName`
5. **Assert expectations** - Always call `AssertExpectations()` on mocks
6. **Keep tests isolated** - Each test should be independent

## Next Steps

To improve test coverage:

1. **Add handler tests** - Test HTTP handlers with mocked use cases
2. **Add service tests** - Test token service and cloud service
3. **Add repository tests** - Integration tests with test database
4. **Add integration tests** - End-to-end API tests
5. **Add benchmark tests** - Performance testing for critical paths

## CI/CD Integration

Tests can be integrated into CI/CD:

```yaml
# Example GitHub Actions
- name: Run tests
  run: make test

- name: Generate coverage
  run: make test-coverage

- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    file: ./coverage.out
```

## Test Statistics

- **Total test files**: 4
- **Total test cases**: 17+
- **Test execution time**: ~2 seconds
- **Coverage target**: 80%+ (currently 10.4%)
