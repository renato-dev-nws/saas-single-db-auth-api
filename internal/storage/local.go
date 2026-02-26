package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type LocalProvider struct {
	basePath string
	baseURL  string
}

func NewLocalProvider(basePath, baseURL string) *LocalProvider {
	return &LocalProvider{basePath: basePath, baseURL: baseURL}
}

func (l *LocalProvider) Upload(file multipart.File, header *multipart.FileHeader, path string) (string, string, error) {
	ext := filepath.Ext(header.Filename)
	filename := uuid.New().String() + ext
	storagePath := filepath.Join(path, filename)
	fullPath := filepath.Join(l.basePath, storagePath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", "", fmt.Errorf("failed to create directory: %w", err)
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", "", fmt.Errorf("failed to write file: %w", err)
	}

	publicURL := fmt.Sprintf("%s/%s", l.baseURL, storagePath)
	return publicURL, storagePath, nil
}

func (l *LocalProvider) Delete(storagePath string) error {
	fullPath := filepath.Join(l.basePath, storagePath)
	return os.Remove(fullPath)
}

func (l *LocalProvider) GetReader(storagePath string) (io.ReadCloser, error) {
	fullPath := filepath.Join(l.basePath, storagePath)
	return os.Open(fullPath)
}
