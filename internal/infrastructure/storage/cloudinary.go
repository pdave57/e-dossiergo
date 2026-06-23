package storage
// internal/infrastructure/storage/cloudinary.go

import (
	"context"
	"io"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryStorage struct {
	client *cloudinary.Cloudinary
}

func NewCloudinaryStorage(cloudName, apiKey, apiSecret string) (*CloudinaryStorage, error) {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, err
	}
	return &CloudinaryStorage{client: cld}, nil
}

func (c *CloudinaryStorage) Upload(ctx context.Context, file io.Reader, filename string, folder string) (string, string, error) {
	uploadParams := uploader.UploadParams{
		PublicID: filename,
		Folder:   folder,
	}
	
	result, err := c.client.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", "", err
	}
	
	return result.SecureURL, result.PublicID, nil
}

func (c *CloudinaryStorage) Delete(ctx context.Context, publicID string) error {
	_, err := c.client.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
	return err
}