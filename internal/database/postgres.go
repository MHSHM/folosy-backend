package database

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var MigrationFiles embed.FS

func Connect(databaseURL string) (*pgxpool.Pool, error) {
	if databaseURL == "" {
		return nil, errors.New("database URL is required")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	config.MaxConns = 25                      // Max number of open connections at the same time
	config.MinConns = 5                       // Min number of connections the pool start with them being ready
	config.MaxConnLifetime = time.Hour        // Max time a given connection can be reused
	config.MaxConnIdleTime = 30 * time.Minute // Max time a connection can be idle before closing it
	config.HealthCheckPeriod = time.Minute    // How often the pool health is checked

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

func RunDBMigrations(databaseURL string) error {
	sourceDriver, err := iofs.New(MigrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("create source driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, databaseURL)
	if err != nil {
		fmt.Errorf("failed to create migration instance: %v", err)
	}
	defer m.Close()

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		fmt.Errorf("migration failed: %v", err)
	}

	log.Println("Database migrated successfully!")

	return nil
}
