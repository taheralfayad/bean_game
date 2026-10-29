package game

import "context"

type gameStore interface {
	CountPlayers(ctx context.Context, gameId uint64) (uint16, error)
	StartGame(ctx context.Context, gameId uint64, variantId uint16) (*Game, error)
}

type variantFinder interface {
	FindByPlayerCount(ctx context.Context, playerCount uint16) (uint16, error)
}

type Service struct {
	store    gameStore
	variants variantFinder
}

func NewService(store gameStore, variants variantFinder) *Service {
	return &Service{store: store, variants: variants}
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

	return s.store.StartGame(ctx, gameId, variantId)
}
