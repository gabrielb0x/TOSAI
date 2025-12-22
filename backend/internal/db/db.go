package db

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	configdata "github.com/gabrielb0x/TOSAI/backend/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

	log.Println("database: connexion OK")

	if err := runDDL(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	log.Println("database: schéma ok")

	return pool, nil
}

func runDDL(ctx context.Context, pool *pgxpool.Pool) error {
	ddlBytes, err := configdata.SchemaFS.ReadFile("database_init.sql")
	if err != nil {
		return fmt.Errorf("read embedded schema: %w", err)
	}

	ddl := strings.TrimSpace(string(ddlBytes))
	if ddl == "" {
		log.Println("database: aucun DDL à appliquer")
		return nil
	}

	if _, err := pool.Exec(ctx, ddl); err != nil {
		return fmt.Errorf("apply schema: %w. Vérifiez que l'utilisateur PostgreSQL dispose des droits CREATE/ALTER sur la base cible.", err)
	}
	return nil
}
