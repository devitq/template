package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	_ "antifraud/api/openapi"
	"antifraud/internal/config"
	"antifraud/internal/db"
	"antifraud/internal/db/sqlc/storage"
	userrepo "antifraud/internal/repository/user"
	userservice "antifraud/internal/application/service/user"
	userhandler "antifraud/internal/transport/http/v1/user"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/gofiber/fiber/v2"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	fiberSwagger "github.com/gofiber/swagger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/segmentio/kafka-go"
	"github.com/valkey-io/valkey-go"
)

type App struct {
	cfg          *config.Config
	l            *slog.Logger
	router       *fiber.App
	dbPool       *pgxpool.Pool
	s3Client     *s3.Client
	kafkaWriter  *kafka.Writer
	kafkaReader  *kafka.Reader
	valkeyClient valkey.Client
}

func New(ctx context.Context, cfg *config.Config, l *slog.Logger) (*App, error) {
	a := &App{
		cfg: cfg,
		l:   l,
	}

	if err := a.initDB(ctx); err != nil {
		return nil, err
	}

	if err := a.migrateDB(); err != nil {
		return nil, err
	}

	queries := storage.New(a.dbPool)
	userRepo := userrepo.New(queries)
	userService := userservice.New(userRepo)
	userHandler := userhandler.New(userService)

	a.initFiber()

	apiGroup := a.router.Group("/api/v1")
	userGroup := apiGroup.Group("/user")

	userHandler.Setup(userGroup)

	return a, nil
}

func (a *App) Start(errChan chan<- error) {
	a.l.Info("Starting...")
	if err := a.router.Listen(a.cfg.HttpSrv.Addr); err != nil {
		errChan <- err
	}
}

func (a *App) Stop(ctx context.Context) error {
	a.l.Info("[!] Shutting down...")

	var stopErr error

	a.l.Info("Closing database pool...")
	a.dbPool.Close()

	a.l.Info("Stopping http server...")
	if err := a.router.ShutdownWithContext(ctx); err != nil {
		stopErr = errors.Join(stopErr, fmt.Errorf("failed to shutdown http server: %w", err))
	}

	if stopErr != nil {
		return stopErr
	}

	a.l.Info("Stopped gracefully")
	return nil
}

func initDB(ctx context.Context, dbURL string, minIdleConns int32, maxConns int32) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	pool.Config().HealthCheckPeriod = time.Minute
	pool.Config().MinIdleConns = minIdleConns
	pool.Config().MaxConns = maxConns

	return pool, nil
}

func (a *App) initDB(ctx context.Context) error {
	dbPool, err := initDB(ctx, a.cfg.Postgres.URL, a.cfg.Postgres.MinIdleConns, a.cfg.Postgres.MaxConns)
	if err != nil {
		return fmt.Errorf("failed to init db connection: %w", err)
	}
	a.dbPool = dbPool
	return nil
}

func (a *App) migrateDB() error {
	return db.Migrate(sql.OpenDB(stdlib.GetConnector(*a.dbPool.Config().ConnConfig)))
}

func (a *App) initFiber() {
	a.router = fiber.New(fiber.Config{DisableStartupMessage: true})

	a.router.Use(fiberrecover.New())
	a.router.Use(a.loggerMiddleware())

	a.router.Get("/api/v1/docs/*", fiberSwagger.HandlerDefault)
	a.router.Get("/api/v1/docs", func(c *fiber.Ctx) error {
		return c.Redirect("/api/v1/docs/index.html", fiber.StatusMovedPermanently)
	})
}

func (a *App) loggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		latency := time.Since(start)
		statusCode := c.Response().StatusCode()
		clientIP := c.IP()
		path := c.Path()
		rawQuery := string(c.Context().URI().QueryString())
		if rawQuery != "" {
			path += "?" + rawQuery
		}

		level := slog.LevelInfo
		if err != nil {
			level = slog.LevelError
		} else if statusCode >= fiber.StatusBadRequest {
			level = slog.LevelWarn
		}

		attrs := []slog.Attr{
			slog.String("uri", path),
			slog.Int("status", statusCode),
			slog.String("ip", clientIP),
			slog.String("latency", latency.String()),
		}
		if err != nil {
			attrs = append(attrs, slog.String("err", err.Error()))
		}

		a.l.LogAttrs(context.Background(), level, "REQUEST", attrs...)

		return err
	}
}

func (a *App) initS3(ctx context.Context) error {
	cfg := aws.Config{
		Region:       a.cfg.S3.Region,
		BaseEndpoint: aws.String(a.cfg.S3.Endpoint),
		Credentials: aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(
				a.cfg.S3.AccessKey,
				a.cfg.S3.SecretKey,
				""),
		),
	}

	a.s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	_, err := a.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(a.cfg.S3.BucketName),
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) && apiError.ErrorCode() == "BucketAlreadyOwnedByYou" {
			return nil
		}
		return fmt.Errorf("failed to create s3 bucket: %w", err)
	}

	return nil
}

func (a *App) initKafkaWriter() {
	a.kafkaWriter = &kafka.Writer{
		Addr:                   kafka.TCP(a.cfg.Kafka.Addr),
		AllowAutoTopicCreation: true,
	}
}

func (a *App) initKafkaReader() {
	a.kafkaReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{a.cfg.Kafka.Addr},
		GroupID:     a.cfg.Kafka.ReaderGroupID,
		GroupTopics: []string{a.cfg.Kafka.Topic},
	})
}

func (a *App) initValkey() error {
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{a.cfg.Valkey.Addr},
		Password:    a.cfg.Valkey.Password,
	})
	if err != nil {
		return fmt.Errorf("failed to init Valkey connection: %w", err)
	}
	a.valkeyClient = client
	return nil
}
