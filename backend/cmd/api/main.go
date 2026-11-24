package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/taheralfayad/bean_game/backend/internal/bean"
	blobss3 "github.com/taheralfayad/bean_game/backend/internal/blobs/s3"
	"github.com/taheralfayad/bean_game/backend/internal/config"
	"github.com/taheralfayad/bean_game/backend/internal/db"
	"github.com/taheralfayad/bean_game/backend/internal/deck"
	"github.com/taheralfayad/bean_game/backend/internal/game"
	"github.com/taheralfayad/bean_game/backend/internal/gameevents"
	"github.com/taheralfayad/bean_game/backend/internal/gamevariant"
	"github.com/taheralfayad/bean_game/backend/internal/user"
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

	gameEventsBroker := gameevents.NewBroker()

	beanHandler := bean.NewHandler(
		bean.NewService(bean.NewStore(database), blobStorage, cfg.Files.MaxUploadBytes),
	)

	deckService := deck.NewService(deck.NewStore(database))

	deckHandler := deck.NewHandler(deckService)

	gameVariantService := gamevariant.NewService(gamevariant.NewStore(database))

	gamevariantHandler := gamevariant.NewHandler(
		gameVariantService,
	)

	gameHandler := game.NewHandler(
		game.NewService(game.NewStore(database), gameVariantService, deckService, gameEventsBroker),
	)

	gameeventsHandler := gameevents.NewHandler(
		gameEventsBroker,
	)

	userHandler := user.NewHandler(user.NewService(user.NewStore(database)))

	r := gin.Default()
	r.Use(user.Cors())
	adminRoutes := r.Group("", user.AdminOnly())
	beanHandler.Register(adminRoutes)
	deckHandler.Register(adminRoutes)
	gamevariantHandler.Register(adminRoutes)
	gameHandler.Register(r)
	gameeventsHandler.Register(r)
	userHandler.Register(r)

	if err := r.Run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
