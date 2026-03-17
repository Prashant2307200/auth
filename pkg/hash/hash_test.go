package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
