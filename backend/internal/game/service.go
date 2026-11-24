package game

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/taheralfayad/bean_game/backend/internal/deck"
	"github.com/taheralfayad/bean_game/backend/internal/gameevents"
	"github.com/taheralfayad/bean_game/backend/internal/gamevariant"
)

var ErrNotEnoughCards = errors.New("not enough cards to deal starting hands")

type gameStore interface {
	CreateGame(ctx context.Context) (*Game, *Player, error)
	JoinGame(ctx context.Context, gameId uint64) (*Player, error)
	CountPlayers(ctx context.Context, gameId uint64) (uint16, error)
	ListPlayers(ctx context.Context, gameId uint64) ([]Player, error)
	StartGame(ctx context.Context, gameId uint64, variantId uint16, playerCards map[uint64][]uint16, remainingDeck []uint16) (*Game, error)
	BeginGameStart(ctx context.Context, gameId uint64) error
	AbortGameStart(ctx context.Context, gameId uint64) error
}

type variantFinder interface {
	FindByPlayerCount(ctx context.Context, playerCount uint16) (*gamevariant.GameVariant, error)
	FindStartingHandBySeat(ctx context.Context, variantId, seat uint16) (uint16, error)
}

type deckFinder interface {
	FindByID(ctx context.Context, id uint16) (*deck.Deck, error)
}

type eventPublisher interface {
	Publish(event gameevents.Event)
}

type Service struct {
	store     gameStore
	variants  variantFinder
	decks     deckFinder
	publisher eventPublisher
}

func NewService(store gameStore, variants variantFinder, decks deckFinder, publisher eventPublisher) *Service {
	return &Service{store: store, variants: variants, decks: decks, publisher: publisher}
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

func (s *Service) StartGame(ctx context.Context, gameId uint64) (game *Game, err error) {
	if err := s.store.BeginGameStart(ctx, gameId); err != nil {
		return nil, err
	}

	defer func() {
		if err == nil {
			return
		}
		abortCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if abortErr := s.store.AbortGameStart(abortCtx, gameId); abortErr != nil {
			err = errors.Join(err, fmt.Errorf("abort game start: %w", abortErr))
		}
	}()

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

	gameDeck, err := s.decks.FindByID(ctx, variant.DeckId)

	if err != nil {
		return nil, err
	}

	gameCards := []uint16{}

	for _, cardType := range gameDeck.Beans {
		for range cardType.CardCount {
			gameCards = append(gameCards, cardType.BeanId)
		}
	}

	shuffledCards := shuffleCards(gameCards)

	players, err := s.store.ListPlayers(ctx, gameId)

	if err != nil {
		return nil, err
	}

	playersCards := make(map[uint64][]uint16)

	for _, player := range players {
		numCards, err := s.variants.FindStartingHandBySeat(ctx, variant.Id, player.Seat)

		if err != nil {
			return &Game{}, err
		}

		if numCards > uint16(len(shuffledCards)) {
			return nil, fmt.Errorf("%w: seat %d needs %d, deck has %d left", ErrNotEnoughCards, player.Seat, numCards, len(shuffledCards))
		}

		playersCards[player.Id] = shuffledCards[:numCards:numCards]

		shuffledCards = shuffledCards[numCards:]
	}

	game, err = s.store.StartGame(ctx, gameId, variant.Id, playersCards, shuffledCards)
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

func shuffleCards(cards []uint16) []uint16 {
	shuffled := make([]uint16, len(cards))
	copy(shuffled, cards)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.IntN(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled
}
