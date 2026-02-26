package storage

import (
	"io"
	"mime/multipart"
)

// Provider defines the interface for file storage operations
type Provider interface {
	Upload(file multipart.File, header *multipart.FileHeader, path string) (publicURL string, storagePath string, err error)
	Delete(storagePath string) error
	GetReader(storagePath string) (io.ReadCloser, error)
}
