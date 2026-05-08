package main

import (
	"log"

	"paper-assistant-backend/internal/api/router"
	"paper-assistant-backend/internal/pkg/config"
)

func main() {
	cfg := config.Load()
	r := router.New(cfg)
	log.Printf("paper assistant backend listening on %s", cfg.HTTPAddr)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
