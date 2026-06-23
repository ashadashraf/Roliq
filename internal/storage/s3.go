package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/roliq/roliq/internal/config"
)

type ObjectInfo struct {
	SizeBytes   int64
	ContentType string
	SHA256      string
}

type Store interface {
	PresignUpload(ctx context.Context, key, contentType, sha256 string, expires time.Duration) (string, error)
	Head(ctx context.Context, key string) (ObjectInfo, error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
}

type S3Store struct {
	client  *s3.Client
	presign *s3.PresignClient
	bucket  string
}

func NewS3(ctx context.Context, cfg config.Config) (*S3Store, error) {
	options := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(cfg.AWSRegion)}
	if cfg.S3AccessKeyID != "" {
		options = append(options, awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3AccessKeyID, cfg.S3SecretAccessKey, "")))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.S3ForcePathStyle
		if cfg.S3Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.S3Endpoint)
		}
	})
	presignClient := client
	if cfg.S3PublicEndpoint != "" {
		presignClient = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.UsePathStyle = cfg.S3ForcePathStyle
			o.BaseEndpoint = aws.String(cfg.S3PublicEndpoint)
		})
	}
	return &S3Store{client: client, presign: s3.NewPresignClient(presignClient), bucket: cfg.S3Bucket}, nil
}

func (s *S3Store) PresignUpload(ctx context.Context, key, contentType, sha256 string, expires time.Duration) (string, error) {
	result, err := s.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket), Key: aws.String(key), ContentType: aws.String(contentType),
		Metadata: map[string]string{"sha256": sha256},
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("presign upload: %w", err)
	}
	return result.URL, nil
}

func (s *S3Store) Head(ctx context.Context, key string) (ObjectInfo, error) {
	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		return ObjectInfo{}, fmt.Errorf("head object: %w", err)
	}
	return ObjectInfo{SizeBytes: aws.ToInt64(result.ContentLength), ContentType: aws.ToString(result.ContentType), SHA256: result.Metadata["sha256"]}, nil
}

func (s *S3Store) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		return nil, fmt.Errorf("download object: %w", err)
	}
	return result.Body, nil
}
