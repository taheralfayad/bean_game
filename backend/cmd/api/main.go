package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/taheralfayad/bean_game/backend/internal/config"
	"github.com/taheralfayad/bean_game/backend/internal/db"
	filess3 "github.com/taheralfayad/bean_game/backend/internal/files/s3"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()

	database, err := db.Open(ctx, cfg.DB)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer database.Close()

	fileStorage := filess3.New(cfg.Files)
	if err := fileStorage.EnsureBucket(ctx); err != nil {
		log.Fatalf("ensure files bucket: %v", err)
	}

	r := gin.Default()

	if err := r.Run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
