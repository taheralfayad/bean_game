package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/taheralfayad/bean_game/backend/internal/admin/bean"
	"github.com/taheralfayad/bean_game/backend/internal/admin/deck"
	blobss3 "github.com/taheralfayad/bean_game/backend/internal/blobs/s3"
	"github.com/taheralfayad/bean_game/backend/internal/config"
	"github.com/taheralfayad/bean_game/backend/internal/db"
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

	blobStorage := blobss3.New(cfg.Files)
	if err := blobStorage.EnsureBucket(ctx); err != nil {
		log.Fatalf("ensure blobs bucket: %v", err)
	}

	beanHandler := bean.NewHandler(
		bean.NewService(bean.NewStore(database), blobStorage, cfg.Files.MaxUploadBytes),
	)

	deckHandler := deck.NewHandler(
		deck.NewService(deck.NewStore(database)),
	)

	r := gin.Default()
	beanHandler.Register(r)
	deckHandler.Register(r)

	if err := r.Run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
