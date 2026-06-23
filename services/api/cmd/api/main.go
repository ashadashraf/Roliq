package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/roliq/roliq/internal/auth"
	"github.com/roliq/roliq/internal/config"
	"github.com/roliq/roliq/internal/database"
	"github.com/roliq/roliq/internal/httpapi"
	"github.com/roliq/roliq/internal/observability"
	"github.com/roliq/roliq/internal/storage"
	"github.com/roliq/roliq/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
	cfg, err := config.LoadAPI()
	if err != nil {
		logger.Error("configuration_invalid", "error", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	shutdownTracing, err := observability.InitTracing(ctx, "roliq-api", cfg.OTLPEndpoint)
	if err != nil {
		logger.Error("tracing_initialization_failed", "error", err)
		os.Exit(1)
	}
	defer shutdownTracing(context.Background())
	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database_unavailable", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	redisClient, err := httpapi.RedisFromURL(cfg.RedisURL)
	if err != nil {
		logger.Error("redis_configuration_invalid", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	verifier, err := auth.NewVerifier(ctx, cfg.OIDCIssuerURL, cfg.OIDCJWKSURL, cfg.OIDCAudience)
	if err != nil {
		logger.Error("oidc_configuration_invalid", "error", err)
		os.Exit(1)
	}
	objects, err := storage.NewS3(ctx, cfg)
	if err != nil {
		logger.Error("storage_configuration_invalid", "error", err)
		os.Exit(1)
	}
	server := httpapi.New(store.New(db), objects, verifier, redisClient, cfg.S3Bucket, cfg.WebOrigin, logger)
	go func() {
		logger.Info("api_started", "address", cfg.HTTPAddress, "environment", cfg.Environment)
		if err := server.Echo().Start(cfg.HTTPAddress); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api_stopped_unexpectedly", "error", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Echo().Shutdown(shutdownCtx); err != nil {
		logger.Error("api_shutdown_failed", "error", err)
	}
}
