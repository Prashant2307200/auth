package service

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
	"github.com/cloudinary/cloudinary-go/v2/api"
)

type CloudinaryUploadService struct {
	CloudName string
	APIKey    string
	APISecret string
}

func NewCloudinaryUploadService(cloudName, apiKey, apiSecret string) *CloudinaryUploadService {
	return &CloudinaryUploadService{
		CloudName: cloudName,
		APIKey:    apiKey,
		APISecret: apiSecret,
	}
}

func (u *CloudinaryUploadService) GenerateUploadSignature(ctx context.Context, userID int64) (*interfaces.UploadSignature, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	publicID := fmt.Sprintf("profile_%d_%s", userID, timestamp)

	params := url.Values{}
	params.Set("folder", "profile_pics")
	params.Set("public_id", publicID)
	params.Set("timestamp", timestamp)

	signature, err := api.SignParameters(params, u.APISecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign upload params: %w", err)
	}

	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", u.CloudName)

	return &interfaces.UploadSignature{
		UploadURL: uploadURL,
		APIKey:    u.APIKey,
		Signature: signature,
		Timestamp: timestamp,
		Folder:    "profile_pics",
		PublicID:  publicID,
		CloudName: u.CloudName,
	}, nil
}
