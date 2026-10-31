package game

import (
	"context"
	"fmt"

	"github.com/taheralfayad/bean_game/backend/internal/gameevents"
)

type gameStore interface {
	CreateGame(ctx context.Context) (*Game, *Player, error)
	JoinGame(ctx context.Context, gameId uint64) (*Player, error)
	CountPlayers(ctx context.Context, gameId uint64) (uint16, error)
	StartGame(ctx context.Context, gameId uint64, variantId uint16) (*Game, error)
}

type variantFinder interface {
	FindByPlayerCount(ctx context.Context, playerCount uint16) (uint16, error)
}

type eventPublisher interface {
	Publish(event gameevents.Event)
}

type Service struct {
	store     gameStore
	variants  variantFinder
	publisher eventPublisher
}

func NewService(store gameStore, variants variantFinder, publisher eventPublisher) *Service {
	return &Service{store: store, variants: variants, publisher: publisher}
}

func (s *Service) CreateGame(ctx context.Context) (*Game, *Player, error) {
	return s.store.CreateGame(ctx)
}

func (s *Service) JoinGame(ctx context.Context, gameId uint64) (*Player, error) {
	player, err := s.store.JoinGame(ctx, gameId)
	if err != nil {
		return nil, err
	}

	s.publisher.Publish(gameevents.Event{
		Topic:   fmt.Sprintf("game:%d", gameId),
		Type:    "game.player_joined",
		Payload: player,
	})

	return player, nil
}

func (s *Service) StartGame(ctx context.Context, gameId uint64) (*Game, error) {
	playerCount, err := s.store.CountPlayers(ctx, gameId)
	if err != nil {
		return nil, err
	}
	if playerCount == 0 {
		return nil, ErrGameNotFound
	}

	variantId, err := s.variants.FindByPlayerCount(ctx, playerCount)
	if err != nil {
		return nil, err
	}

	game, err := s.store.StartGame(ctx, gameId, variantId)
	if err != nil {
		return nil, err
	}

	s.publisher.Publish(gameevents.Event{
		Topic:   fmt.Sprintf("game:%d", gameId),
		Type:    "game.started",
		Payload: game,
	})

	return game, nil
}
