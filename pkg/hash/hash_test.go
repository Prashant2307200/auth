package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "testpassword123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "verylongpasswordthatexceedsnormallength123456789",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashed, err := HashPassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, hashed)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hashed)
				assert.NotEqual(t, tt.password, hashed, "hashed password should not equal plain password")
			}
		})
	}
}

func TestHashPassword_UniqueHashes(t *testing.T) {
	password := "testpassword123"
	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEqual(t, hash1, hash2, "each hash should be unique due to salt")
}

func TestCheckPassword(t *testing.T) {
	password := "testpassword123"
	hashedPassword, err := HashPassword(password)
	require.NoError(t, err)

	tests := []struct {
		name           string
		hashedPassword string
		plainPassword  string
		wantErr        bool
	}{
		{
			name:           "correct password",
			hashedPassword: hashedPassword,
			plainPassword:  password,
			wantErr:        false,
		},
		{
			name:           "incorrect password",
			hashedPassword: hashedPassword,
			plainPassword:  "wrongpassword",
			wantErr:        true,
		},
		{
			name:           "empty password",
			hashedPassword: hashedPassword,
			plainPassword:  "",
			wantErr:        true,
		},
		{
			name:           "invalid hash",
			hashedPassword: "invalidhash",
			plainPassword:  password,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPassword(tt.hashedPassword, tt.plainPassword)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHashPassword_CheckPassword_RoundTrip(t *testing.T) {
	passwords := []string{
		"password123",
		"P@ssw0rd!",
		"verylongpassword123456789",
		"short",
		"",
	}

	for _, password := range passwords {
		t.Run(password, func(t *testing.T) {
			hashed, err := HashPassword(password)
			require.NoError(t, err)

			err = CheckPassword(hashed, password)
			assert.NoError(t, err, "should be able to verify the original password")
		})
	}
}

func TestNeedsRehashAndHashWithCost(t *testing.T) {
	// Empty hash should return error and indicate rehash needed
	need, err := NeedsRehash("")
	assert.True(t, need, "empty hash should need rehash")
	assert.Error(t, err, "empty hash should return an error")

	// Invalid (non-bcrypt) hash should indicate rehash but not return an error
	need, err = NeedsRehash("not-a-bcrypt-hash")
	assert.True(t, need, "invalid hash should need rehash")
	assert.NoError(t, err, "invalid bcrypt hash should not return an error from NeedsRehash")

	// Create a hash with lower cost and ensure NeedsRehash notices it
	lowCost := CurrentCost - 1
	if lowCost < bcrypt.MinCost {
		lowCost = bcrypt.MinCost
	}
	lowHash, err := HashPasswordWithCost("pw-for-low-cost", lowCost)
	require.NoError(t, err)
	need, err = NeedsRehash(lowHash)
	assert.True(t, need, "hash with lower cost should need rehash")
	assert.NoError(t, err)

	// Create a hash with exactly CurrentCost and ensure NeedsRehash returns false
	exactHash, err := HashPasswordWithCost("pw-for-exact-cost", CurrentCost)
	require.NoError(t, err)
	need, err = NeedsRehash(exactHash)
	assert.False(t, need, "hash with current cost should NOT need rehash")
	assert.NoError(t, err)

	// HashPasswordWithCost should return error for wildly invalid cost values (too large)
	_, err = HashPasswordWithCost("pw", 1000)
	assert.Error(t, err, "very large cost should return an error")
}

func TestHashPassword_BcryptCostAndLargeInputs(t *testing.T) {
	// Ensure HashPassword uses a sane cost (between 10 and 14 inclusive)
	hashed, err := HashPassword("cost-check")
	require.NoError(t, err)
	cost, err := bcrypt.Cost([]byte(hashed))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, cost, 10, "bcrypt cost should be >= 10")
	assert.LessOrEqual(t, cost, 14, "bcrypt cost should be <= 14")

	// Very large password within bcrypt limit should still hash and verify
	// bcrypt has a 72-byte input limit for the effective password; test at that boundary
	veryLong := strings.Repeat("a", 72)
	hashedLong, err := HashPassword(veryLong)
	require.NoError(t, err)
	err = CheckPassword(hashedLong, veryLong)
	assert.NoError(t, err, "72-byte password should roundtrip via hash and verify")

	// Extremely large password (well over 72 bytes) should produce an error from bcrypt
	tooLong := strings.Repeat("a", 5000)
	_, err = HashPassword(tooLong)
	assert.Error(t, err, "passwords longer than bcrypt's limit should return an error")
}

func TestCheckPassword_EdgeCases(t *testing.T) {
	// Empty hash must fail verification
	err := CheckPassword("", "anything")
	assert.Error(t, err, "empty hash should not verify any password")

	// Case sensitivity: different case should fail
	hashed, err := HashPassword("PasswordCase")
	require.NoError(t, err)
	err = CheckPassword(hashed, "passwordcase")
	assert.Error(t, err, "password checking should be case-sensitive")

	// Hash an empty password and verify empty succeeds
	emptyHash, err := HashPassword("")
	require.NoError(t, err)
	err = CheckPassword(emptyHash, "")
	assert.NoError(t, err, "empty password when hashed should verify an empty password")
}
