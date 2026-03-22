package encrypt

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestKey(t *testing.T) []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)
	return key
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := generateTestKey(t)
	plaintext := "my-secret-totp-key-JBSWY3DPEHPK3PXP"

	encrypted, err := Encrypt(plaintext, key)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := Decrypt(encrypted, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncrypt_DifferentOutputs(t *testing.T) {
	key := generateTestKey(t)
	plaintext := "test-secret"

	enc1, err := Encrypt(plaintext, key)
	require.NoError(t, err)

	enc2, err := Encrypt(plaintext, key)
	require.NoError(t, err)

	assert.NotEqual(t, enc1, enc2, "each encryption should produce unique output due to random nonce")
}

func TestDecrypt_WrongKey(t *testing.T) {
	key1 := generateTestKey(t)
	key2 := generateTestKey(t)

	encrypted, err := Encrypt("secret", key1)
	require.NoError(t, err)

	_, err = Decrypt(encrypted, key2)
	assert.Error(t, err)
}

func TestEncrypt_InvalidKeyLength(t *testing.T) {
	_, err := Encrypt("test", []byte("short"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "32 bytes")
}

func TestDecrypt_InvalidKeyLength(t *testing.T) {
	_, err := Decrypt("dGVzdA==", []byte("short"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "32 bytes")
}

func TestDecrypt_InvalidBase64(t *testing.T) {
	key := generateTestKey(t)
	_, err := Decrypt("not-valid-base64!!!", key)
	assert.Error(t, err)
}

func TestDecrypt_TooShortCiphertext(t *testing.T) {
	key := generateTestKey(t)
	_, err := Decrypt("dGVzdA==", key) // "test" in base64, too short for nonce
	assert.Error(t, err)
}
