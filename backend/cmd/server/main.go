package main

import (
	"fmt"
	"log"

	"github.com/gabrielb0x/TOSAI/backend/internal/config"
	"github.com/gabrielb0x/TOSAI/backend/internal/db"
	"github.com/gabrielb0x/TOSAI/backend/internal/server"
)

func main() {
	cfg := config.Load()

	pool, err := db.ConnectAndMigrate(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("initialisation base TOSAI échouée: %v", err)
	}
	defer pool.Close()

	r := server.New(cfg, pool)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("TOSAI backend prêt sur %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
