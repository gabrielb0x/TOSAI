package main

import (
	"fmt"
	"log"

	"github.com/auto-tos/auto-tos/internal/config"
	"github.com/auto-tos/auto-tos/internal/server"
)

func main() {
	cfg := config.Load()

	r := server.New(cfg)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
