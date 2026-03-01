package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

type R2Provider struct {
	client    *s3.S3
	bucket    string
	publicURL string
}

func NewR2Provider(accountID, accessKey, secretKey, bucket, publicURL string) (*R2Provider, error) {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("auto"),
		Endpoint:    aws.String(endpoint),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create R2 session: %w", err)
	}

	return &R2Provider{
		client:    s3.New(sess),
		bucket:    bucket,
		publicURL: publicURL,
	}, nil
}

func (r *R2Provider) Upload(file multipart.File, header *multipart.FileHeader, path string) (string, string, error) {
	ext := filepath.Ext(header.Filename)
	filename := uuid.New().String() + ext
	key := fmt.Sprintf("%s/%s", path, filename)

	_, err := r.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(header.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to R2: %w", err)
	}

	publicURL := fmt.Sprintf("%s/%s", r.publicURL, key)
	return publicURL, key, nil
}

func (r *R2Provider) Delete(storagePath string) error {
	_, err := r.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(storagePath),
	})
	return err
}

func (r *R2Provider) GetReader(storagePath string) (io.ReadCloser, error) {
	result, err := r.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(storagePath),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}
