package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConfig struct {
	URL          string
	MaxConns     int32
	MinIdleConns int32
}

func (c PostgresConfig) Validate() error {
	if c.MaxConns <= 0 {
		return fmt.Errorf("MaxConns must be greater than 0")
	}
	if c.MinIdleConns < 0 {
		return fmt.Errorf("MinIdleConns cannot be negative")
	}
	if c.MinIdleConns > c.MaxConns {
		return fmt.Errorf("MinIdleConns cannot be greater than MaxConns")
	}
	return nil
}

func NewPostgresPool(ctx context.Context, cfg PostgresConfig) (*pgxpool.Pool, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid postgres config: %w", err)
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres url: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinIdleConns
	poolCfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return pool, nil
}
