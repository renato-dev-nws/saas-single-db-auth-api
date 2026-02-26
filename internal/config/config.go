package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL string
	RedisURL    string

	JWTSecret      string
	JWTExpiryHours int

	StorageProvider  string
	StorageLocalPath string
	StorageBaseURL   string

	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	AWSBucket          string

	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2Bucket          string
	R2PublicURL       string

	TenantAPIPort string
	AdminAPIPort  string
	AppAPIPort    string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://saasuser:saaspassword@localhost:5432/saasdb?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),

		JWTSecret:      getEnv("JWT_SECRET", "supersecretkey"),
		JWTExpiryHours: getEnvInt("JWT_EXPIRY_HOURS", 24),

		StorageProvider:  getEnv("STORAGE_PROVIDER", "local"),
		StorageLocalPath: getEnv("STORAGE_LOCAL_PATH", "./uploads"),
		StorageBaseURL:   getEnv("STORAGE_BASE_URL", "http://localhost:8080/uploads"),

		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		AWSRegion:          getEnv("AWS_REGION", "us-east-1"),
		AWSBucket:          getEnv("AWS_BUCKET", ""),

		R2AccountID:       getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2Bucket:          getEnv("R2_BUCKET", ""),
		R2PublicURL:       getEnv("R2_PUBLIC_URL", ""),

		TenantAPIPort: getEnv("TENANT_API_PORT", "8080"),
		AdminAPIPort:  getEnv("ADMIN_API_PORT", "8081"),
		AppAPIPort:    getEnv("APP_API_PORT", "8082"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}
