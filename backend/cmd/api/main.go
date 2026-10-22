package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/taheralfayad/bean_game/backend/internal/config"
	"github.com/taheralfayad/bean_game/backend/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	database, err := db.Open(context.Background(), cfg.DB)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer database.Close()

	r := gin.Default()

	if err := r.Run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
