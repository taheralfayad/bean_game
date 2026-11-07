package game

import (
	"context"
	"errors"
	"fmt"

	"github.com/taheralfayad/bean_game/backend/internal/gameevents"
	"github.com/taheralfayad/bean_game/backend/internal/gamevariant"
)

const startGameAttempts = 3

type gameStore interface {
	CreateGame(ctx context.Context) (*Game, *Player, error)
	JoinGame(ctx context.Context, gameId uint64) (*Player, error)
	CountPlayers(ctx context.Context, gameId uint64) (uint16, error)
	ListPlayers(ctx context.Context, gameId uint64) ([]Player, error)
	StartGame(ctx context.Context, gameId uint64, variantId uint16, expectedPlayerCount uint16) (*Game, error)
}

type variantFinder interface {
	FindByPlayerCount(ctx context.Context, playerCount uint16) (*gamevariant.GameVariant, error)
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

func (s *Service) ListPlayers(ctx context.Context, gameId uint64) ([]Player, error) {
	return s.store.ListPlayers(ctx, gameId)
}

func (s *Service) StartGame(ctx context.Context, gameId uint64) (*Game, error) {
	// documenting a decision i had to make. in order to find the variant
	// we have to call the method that the variants repository exposes for us
	// otherwise, we would be breaking module boundaries. however, this could
	// potentially introduce a time-of-check to time-of-use (TOCTOU race condition).
	// to solve this without checking for game variant id in the game repository method,
	// we basically retry starting the game if the player count has changed.

	var game *Game

	for attempt := 0; ; attempt++ {

		playerCount, err := s.store.CountPlayers(ctx, gameId)
		if err != nil {
			return nil, err
		}
		if playerCount == 0 {
			return nil, ErrGameNotFound
		}

		variant, err := s.variants.FindByPlayerCount(ctx, playerCount)
		if err != nil {
			return nil, err
		}

		game, err = s.store.StartGame(ctx, gameId, variant.Id, playerCount)
		if errors.Is(err, ErrPlayersChanged) && attempt+1 < startGameAttempts {
			continue
		}
		if err != nil {
			return nil, err
		}

		break
	}

	s.publisher.Publish(gameevents.Event{
		Topic:   fmt.Sprintf("game:%d", gameId),
		Type:    "game.started",
		Payload: game,
	})

	return game, nil
}
