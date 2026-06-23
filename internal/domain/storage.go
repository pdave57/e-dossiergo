// internal/domain/storage.go
package domain

import (
	"context"
	"io"
)

// ImageStorage defines the contract for file storage (e.g., Cloudinary).
type ImageStorage interface {
	Upload(ctx context.Context, file io.Reader, filename string, folder string) (url string, publicID string, err error)
	Delete(ctx context.Context, publicID string) error
}