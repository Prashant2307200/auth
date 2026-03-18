package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCloudinaryUploadService_GenerateUploadSignature(t *testing.T) {
	svc := NewCloudinaryUploadService("test-cloud", "test-key", "test-secret")

	sig, err := svc.GenerateUploadSignature(context.Background(), 42)
	require.NoError(t, err)
	require.NotNil(t, sig)
	require.Contains(t, sig.UploadURL, "test-cloud")
	require.Equal(t, "test-key", sig.APIKey)
	require.NotEmpty(t, sig.Signature)
	require.NotEmpty(t, sig.Timestamp)
	require.Equal(t, "profile_pics", sig.Folder)
	require.Equal(t, "test-cloud", sig.CloudName)
	require.Contains(t, sig.PublicID, "profile_42_")
}
