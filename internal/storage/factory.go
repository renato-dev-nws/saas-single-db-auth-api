package storage

import (
	"fmt"

	"github.com/saas-single-db-api/internal/config"
)

func NewProvider(cfg *config.Config) (Provider, error) {
	switch cfg.StorageProvider {
	case "local":
		return NewLocalProvider(cfg.StorageLocalPath, cfg.StorageBaseURL), nil
	case "s3":
		return NewS3Provider(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, cfg.AWSRegion, cfg.AWSBucket)
	case "r2":
		return NewR2Provider(cfg.R2AccountID, cfg.R2AccessKeyID, cfg.R2SecretAccessKey, cfg.R2Bucket, cfg.R2PublicURL)
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", cfg.StorageProvider)
	}
}
