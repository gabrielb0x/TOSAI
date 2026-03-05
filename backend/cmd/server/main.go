package main

import (
	"fmt"
	"log"

	"github.com/gabrielb0x/TOSAI/backend/internal/config"
	"github.com/gabrielb0x/TOSAI/backend/internal/db"
	"github.com/gabrielb0x/TOSAI/backend/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	var pool *pgxpool.Pool
	if cfg.DatabaseURL != "" {
		connectedPool, err := db.ConnectAndMigrate(cfg.DatabaseURL)
		if err != nil {
			log.Printf("avertissement: initialisation PostgreSQL échouée (%v) ; démarrage sans DB", err)
		} else {
			pool = connectedPool
			defer connectedPool.Close()
		}
	}

	r := server.New(cfg, pool)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("TOSAI backend prêt sur %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
