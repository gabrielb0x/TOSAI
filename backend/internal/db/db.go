package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const ddl = `
CREATE TABLE IF NOT EXISTS tos_sources (
    id BIGSERIAL PRIMARY KEY,
    url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tos_summaries (
    id BIGSERIAL PRIMARY KEY,
    source_id BIGINT REFERENCES tos_sources(id) ON DELETE CASCADE,
    summary TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

// ConnectAndMigrate creates a PostgreSQL connection pool and ensures that the
// baseline tables exist. It returns a live pool that callers should Close when
// shutting down the application.
func ConnectAndMigrate(databaseURL string) (*pgxpool.Pool, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := runDDL(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func runDDL(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, ddl); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}
