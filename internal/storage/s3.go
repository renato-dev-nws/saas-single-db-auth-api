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

type S3Provider struct {
	client *s3.S3
	bucket string
}

func NewS3Provider(accessKey, secretKey, region, bucket string) (*S3Provider, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 session: %w", err)
	}

	return &S3Provider{
		client: s3.New(sess),
		bucket: bucket,
	}, nil
}

func (s *S3Provider) Upload(file multipart.File, header *multipart.FileHeader, path string) (string, string, error) {
	ext := filepath.Ext(header.Filename)
	filename := uuid.New().String() + ext
	key := fmt.Sprintf("%s/%s", path, filename)

	_, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(header.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	publicURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, key)
	return publicURL, key, nil
}

func (s *S3Provider) Delete(storagePath string) error {
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storagePath),
	})
	return err
}

func (s *S3Provider) GetReader(storagePath string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storagePath),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}
