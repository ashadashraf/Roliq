package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/roliq/roliq/internal/background"
	"github.com/roliq/roliq/internal/config"
	"github.com/roliq/roliq/internal/database"
	"github.com/roliq/roliq/internal/storage"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	cfg, err := config.LoadWorker()
	if err != nil {
		logger.Error("configuration_invalid", "error", err)
		os.Exit(1)
	}
	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database_unavailable", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	objects, err := storage.NewS3(ctx, cfg)
	if err != nil {
		logger.Error("storage_unavailable", "error", err)
		os.Exit(1)
	}
	var queue *sqs.Client
	if cfg.SQSQueueURL != "" {
		options := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(cfg.AWSRegion)}
		if cfg.S3AccessKeyID != "" {
			options = append(options, awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3AccessKeyID, cfg.S3SecretAccessKey, "")))
		}
		awsCfg, loadErr := awsconfig.LoadDefaultConfig(ctx, options...)
		if loadErr != nil {
			logger.Error("queue_configuration_invalid", "error", loadErr)
			os.Exit(1)
		}
		queue = sqs.NewFromConfig(awsCfg, func(o *sqs.Options) {
			if cfg.SQSEndpoint != "" {
				o.BaseEndpoint = aws.String(cfg.SQSEndpoint)
			}
		})
	}
	logger.Info("background_worker_started")
	background.New(db, objects, queue, cfg.SQSQueueURL, cfg.ClamAVAddress, logger).Run(ctx, cfg.PollInterval)
}
