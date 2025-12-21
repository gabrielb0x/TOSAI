package main

import (
	"fmt"
	"log"

	"github.com/auto-tos/auto-tos/internal/config"
	"github.com/auto-tos/auto-tos/internal/db"
	"github.com/auto-tos/auto-tos/internal/server"
)

func main() {
	cfg := config.Load()

	pool, err := db.ConnectAndMigrate(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database setup failed: %v", err)
	}
	defer pool.Close()

	r := server.New(cfg, pool)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
