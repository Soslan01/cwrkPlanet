package database

import (
	"context"
	"fmt"

	"github.com/cwrk-planet/auth-service/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, cfg config.Postgres) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	if cfg.MaxConns > 0 {
		poolConfig.MaxConns = cfg.MaxConns
	}
	if cfg.MinConns > 0 {
		poolConfig.MinConns = cfg.MinConns
	}
	if cfg.MaxConnLifeTime > 0 {
		poolConfig.MaxConnLifetime = cfg.MaxConnLifeTime
	}
	if cfg.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	}
	if cfg.ApplicationName != "" {
		poolConfig.ConnConfig.Config.RuntimeParams["application_name"] = cfg.ApplicationName
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return &DB{Pool: pool}, nil
}
