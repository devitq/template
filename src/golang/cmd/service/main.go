package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"antifraud/internal/app"
	"antifraud/internal/config"
	_ "go.uber.org/automaxprocs"
)

// @title		Go Template
// @version	1.0
// @BasePath	/api/v1
// @host		localhost:8080
func main() {
	logger := setupLogger()

	cfg, err := config.New()
	if err != nil {
		logger.Error("failed to read config", slog.Any("error", err))
		os.Exit(1)
	}

	a, err := app.New(context.Background(), cfg, logger)
	if err != nil {
		logger.Error("failed to create app", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errChan := make(chan error)
	go a.Start(errChan)

	select {
	case err := <-errChan:
		logger.Error("failed to start server", slog.Any("error", err))
		os.Exit(1)
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.Stop(ctx); err != nil {
			logger.Error("failed to stop server", slog.Any("error", err))
			os.Exit(1)
		}
	}
}

func setupLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	})

	logger := slog.New(handler)
	return logger
}
