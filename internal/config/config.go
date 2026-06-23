package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Environment       string
	HTTPAddress       string
	DatabaseURL       string
	RedisURL          string
	OIDCIssuerURL     string
	OIDCJWKSURL       string
	OIDCAudience      string
	OIDCClientID      string
	WebOrigin         string
	AWSRegion         string
	S3Bucket          string
	S3Endpoint        string
	S3PublicEndpoint  string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3ForcePathStyle  bool
	OTLPEndpoint      string
	LogLevel          string
	ClamAVAddress     string
	SQSQueueURL       string
	SQSEndpoint       string
	PollInterval      time.Duration
}

func LoadAPI() (Config, error) {
	c := base()
	c.HTTPAddress = env("HTTP_ADDRESS", ":8080")
	c.DatabaseURL = os.Getenv("DATABASE_URL")
	c.RedisURL = env("REDIS_URL", "redis://localhost:6379/0")
	c.OIDCIssuerURL = os.Getenv("OIDC_ISSUER_URL")
	c.OIDCJWKSURL = os.Getenv("OIDC_JWKS_URL")
	c.OIDCAudience = os.Getenv("OIDC_AUDIENCE")
	c.OIDCClientID = os.Getenv("OIDC_CLIENT_ID")
	c.WebOrigin = env("WEB_ORIGIN", "http://localhost:3000")
	c.OTLPEndpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if c.DatabaseURL == "" || c.OIDCIssuerURL == "" || c.OIDCAudience == "" {
		return Config{}, fmt.Errorf("DATABASE_URL, OIDC_ISSUER_URL, and OIDC_AUDIENCE are required")
	}
	return c, nil
}

func LoadWorker() (Config, error) {
	c := base()
	c.DatabaseURL = os.Getenv("WORKER_DATABASE_URL")
	c.ClamAVAddress = env("CLAMAV_ADDRESS", "localhost:3310")
	c.SQSQueueURL = os.Getenv("SQS_QUEUE_URL")
	c.SQSEndpoint = os.Getenv("SQS_ENDPOINT")
	var err error
	c.PollInterval, err = time.ParseDuration(env("POLL_INTERVAL", "2s"))
	if err != nil || c.PollInterval < time.Second {
		return Config{}, fmt.Errorf("POLL_INTERVAL must be at least one second")
	}
	if c.DatabaseURL == "" {
		return Config{}, fmt.Errorf("WORKER_DATABASE_URL is required")
	}
	return c, nil
}

func base() Config {
	forcePath, _ := strconv.ParseBool(env("S3_FORCE_PATH_STYLE", "false"))
	return Config{
		Environment:       env("APP_ENV", "development"),
		AWSRegion:         env("AWS_REGION", "us-east-1"),
		S3Bucket:          os.Getenv("S3_BUCKET"),
		S3Endpoint:        os.Getenv("S3_ENDPOINT"),
		S3PublicEndpoint:  os.Getenv("S3_PUBLIC_ENDPOINT"),
		S3AccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
		S3SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
		S3ForcePathStyle:  forcePath,
		LogLevel:          env("LOG_LEVEL", "info"),
		OTLPEndpoint:      os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
