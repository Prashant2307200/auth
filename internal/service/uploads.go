package service

import (
	"context"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CoudinaryUploadService struct {
	Cld *cloudinary.Cloudinary
}

func NewCoudinaryUploadService(cld *cloudinary.Cloudinary) *CoudinaryUploadService {
	return &CoudinaryUploadService{
		Cld: cld,
	}
}

func (u *CoudinaryUploadService) UploadImage(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {

	res, err := u.Cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: "profile_pics", // Optional: Cloudinary folder
		PublicID: fileHeader.Filename,
	})
	if err != nil {
		return "", err
	}

	return res.SecureURL, nil
}
