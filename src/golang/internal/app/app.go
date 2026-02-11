package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	_ "antifraud/api/openapi"
	"antifraud/internal/infra/config"
	"antifraud/internal/infra/database"
	sqlcstorage "antifraud/internal/infra/database/sqlc/storage"
	"antifraud/internal/infra/kafka"
	"antifraud/internal/infra/observability"
	infrastorage "antifraud/internal/infra/storage"
	"antifraud/internal/infra/valkey"
	"antifraud/internal/infra/validator"
	userrepo "antifraud/internal/adapter/repository/user"
	userservice "antifraud/internal/usecase/user"
	userhandler "antifraud/internal/adapter/handler/http/v1/user"
	"antifraud/internal/adapter/handler/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	fiberSwagger "github.com/gofiber/swagger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	kafkago "github.com/segmentio/kafka-go"
	valkeygo "github.com/valkey-io/valkey-go"
	"go.opentelemetry.io/otel/trace"
)

type App struct {
	cfg          *config.Config
	l            *slog.Logger
	router       *fiber.App
	dbPool       *pgxpool.Pool
	s3Client     *s3.Client
	kafkaWriter  *kafkago.Writer
	kafkaReader  *kafkago.Reader
	valkeyClient valkeygo.Client
	otelShutdown func(context.Context) error
}

func New(ctx context.Context, cfg *config.Config, l *slog.Logger) (*App, error) {
	a := &App{
		cfg: cfg,
		l:   l,
	}

	if cfg.Otel.Enabled {
		shutdown, err := observability.SetupOTEL(ctx, observability.OTELConfig{
			ServiceName:           cfg.Otel.ServiceName,
			CollectorURL:          cfg.Otel.CollectorURL,
			UseTLS:                cfg.Otel.UseTLS,
			CertFile:              cfg.Otel.CertFile,
			MetricsExportInterval: time.Duration(cfg.Otel.MetricsExportInterval) * time.Second,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to setup otel: %w", err)
		}
		a.otelShutdown = shutdown
	}

	if err := a.initDB(ctx); err != nil {
		return nil, err
	}

	if err := a.migrateDB(); err != nil {
		return nil, err
	}

	v := validator.New()

	queries := sqlcstorage.New(a.dbPool)
	userRepo := userrepo.New(queries)
	userService := userservice.New(userRepo)
	userHandler := userhandler.New(userService, v, a.l)

	a.initFiber()

	apiGroup := a.router.Group("/api/v1")

	a.router.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

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

	if a.otelShutdown != nil {
		a.l.Info("Shutting down OTEL...")
		if err := a.otelShutdown(ctx); err != nil {
			stopErr = errors.Join(stopErr, fmt.Errorf("failed to shutdown otel: %w", err))
		}
	}

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

func (a *App) initDB(ctx context.Context) error {
	dbPool, err := database.NewPostgresPool(ctx, database.PostgresConfig{
		URL:          a.cfg.Postgres.URL,
		MaxConns:     a.cfg.Postgres.MaxConns,
		MinIdleConns: a.cfg.Postgres.MinIdleConns,
	})
	if err != nil {
		return fmt.Errorf("failed to init db connection: %w", err)
	}
	a.dbPool = dbPool
	return nil
}

func (a *App) migrateDB() error {
	tempDB := sql.OpenDB(stdlib.GetConnector(*a.dbPool.Config().ConnConfig))
	defer tempDB.Close()
	return database.Migrate(tempDB)
}

func (a *App) initFiber() {
	a.router = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler:         middleware.ErrorHandler,
	})

	a.router.Use(fiberrecover.New())
	if a.cfg.Otel.Enabled {
		a.router.Use(otelfiber.Middleware())
	}
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

		if a.cfg.Otel.Enabled {
			span := trace.SpanFromContext(c.UserContext())
			if span.SpanContext().IsValid() {
				attrs = append(attrs, slog.String("traceId", span.SpanContext().TraceID().String()))
				c.Locals("traceId", span.SpanContext().TraceID().String())
			}
		}
		if err != nil {
			attrs = append(attrs, slog.String("err", err.Error()))
		}

		a.l.LogAttrs(context.Background(), level, "REQUEST", attrs...)

		return err
	}
}

func (a *App) initS3(ctx context.Context) error {
	client, err := infrastorage.NewS3Client(ctx, infrastorage.S3Config{
		Region:     a.cfg.S3.Region,
		Endpoint:   a.cfg.S3.Endpoint,
		AccessKey:  a.cfg.S3.AccessKey,
		SecretKey:  a.cfg.S3.SecretKey,
		BucketName: a.cfg.S3.BucketName,
	})
	if err != nil {
		return err
	}
	a.s3Client = client
	return nil
}

func (a *App) initKafkaWriter() {
	a.kafkaWriter = kafka.NewKafkaWriter(kafka.KafkaConfig{
		Addr: a.cfg.Kafka.Addr,
	})
}

func (a *App) initKafkaReader() {
	a.kafkaReader = kafka.NewKafkaReader(kafka.KafkaConfig{
		Addr:          a.cfg.Kafka.Addr,
		ReaderGroupID: a.cfg.Kafka.ReaderGroupID,
		Topic:         a.cfg.Kafka.Topic,
	})
}

func (a *App) initValkey() error {
	client, err := valkey.NewValkeyClient(valkey.ValkeyConfig{
		Addr:     a.cfg.Valkey.Addr,
		Username: a.cfg.Valkey.Username,
		Password: a.cfg.Valkey.Password,
	})
	if err != nil {
		return err
	}
	a.valkeyClient = client
	return nil
}
