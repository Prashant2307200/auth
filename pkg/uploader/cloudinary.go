package uploader

import "github.com/cloudinary/cloudinary-go/v2"

type Cloudinary struct {
	Cld *cloudinary.Cloudinary
}

func Connect(name, key, secret string) (*Cloudinary, error) {

	cld, err := cloudinary.NewFromParams(name, key, secret)
	if err != nil {
		return nil, err
	}
	return &Cloudinary{Cld: cld}, nil
}