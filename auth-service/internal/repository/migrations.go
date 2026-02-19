package repository

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"sort"

	"github.com/cwrk-planet/auth-service/internal/database"
	"github.com/cwrk-planet/auth-service/internal/logger"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

// Embed all *.up.sql migration files from the local migrations directory.
// Paths are relative to this file's location.
//
//go:embed migrations/*.up.sql
var migrationFiles embed.FS

// RunMigrations executes all embedded *.up.sql migration files in
// lexicographical order. The SQL is written to be idempotent (IF NOT EXISTS).
func RunMigrations(ctx context.Context, db *database.DB, log *logger.Logger) error {
	if db == nil || db.Pool == nil {
		return fmt.Errorf("database pool is not initialized")
	}

	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read embedded migrations: %w", err)
	}

	// Deterministic order: 001_..., 002_..., etc.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		path := "migrations/" + name

		log.Info("Running database migration", zap.String("file", name))

		sqlBytes, err := migrationFiles.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", name, err)
		}

		if _, err := db.Pool.Exec(ctx, string(sqlBytes)); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == "42710" {
					log.Info("migration skipped (object already exists)", zap.String("object", pgErr.TableName))
					return nil
				}
			}
			return fmt.Errorf("failed to execute migration %s: %w", name, err)
		}
	}

	log.Info("Database migrations completed successfully")
	return nil
}
