package gamevariant

import (
	"context"
	"errors"
	"regexp"
	"strings"
)

var (
	ErrNotFound             = errors.New("game variant not found")
	ErrInvalidSlug          = errors.New("slug must match ^[a-z0-9_]+$ and be at most 40 characters")
	ErrInvalidName          = errors.New("name must be between 1 and 80 characters")
	ErrInvalidPlayers       = errors.New("min_players must be at least 1 and max_players must be at least min_players")
	ErrInvalidDrawCount     = errors.New("end_of_turn_draw_count must be greater than zero")
	ErrNoMatch              = errors.New("no game variant matches player count")
	ErrRangeOverlap         = errors.New("game variant player range overlaps an existing variant")
	ErrInvalidStartingHands = errors.New("starting_hands must contain one entry per seat from 1 to max_players, with card_count greater than zero")
)

var slugPattern = regexp.MustCompile(`^[a-z0-9_]+$`)

const (
	maxSlugLength = 40
	maxNameLength = 80
)

type gamevariantStore interface {
	ListVariants(ctx context.Context) ([]GameVariant, error)
	FindByPlayerCount(ctx context.Context, playerCount uint16) (*GameVariant, error)
	InsertVariant(ctx context.Context, variant GameVariant, hands []StartingHand) (*GameVariant, error)
	DeleteVariant(ctx context.Context, id uint16) error
	FindStartingHandBySeat(ctx context.Context, variantId, seat uint16) (uint16, error)
}

type Service struct {
	store gamevariantStore
}

func NewService(store gamevariantStore) *Service {
	return &Service{store: store}
}

func (s *Service) FindStartingHandBySeat(ctx context.Context, variantId, seat uint16) (uint16, error) {
	return s.store.FindStartingHandBySeat(ctx, variantId, seat)
}

func (s *Service) List(ctx context.Context) ([]GameVariant, error) {
	return s.store.ListVariants(ctx)
}

func (s *Service) FindByPlayerCount(ctx context.Context, playerCount uint16) (*GameVariant, error) {
	return s.store.FindByPlayerCount(ctx, playerCount)
}

func (s *Service) Create(ctx context.Context, variant GameVariant, hands []StartingHand) (*GameVariant, error) {
	if len(variant.Slug) > maxSlugLength || !slugPattern.MatchString(variant.Slug) {
		return nil, ErrInvalidSlug
	}

	variant.Name = strings.TrimSpace(variant.Name)
	if variant.Name == "" || len(variant.Name) > maxNameLength {
		return nil, ErrInvalidName
	}

	if variant.MinPlayers < 1 || variant.MaxPlayers < variant.MinPlayers {
		return nil, ErrInvalidPlayers
	}

	if variant.EndOfTurnDrawCount < 1 {
		return nil, ErrInvalidDrawCount
	}

	if err := validateStartingHands(hands, variant.MaxPlayers); err != nil {
		return nil, err
	}

	return s.store.InsertVariant(ctx, variant, hands)
}

func validateStartingHands(hands []StartingHand, maxPlayers uint16) error {
	if len(hands) != int(maxPlayers) {
		return ErrInvalidStartingHands
	}

	seen := make([]bool, maxPlayers+1)
	for _, hand := range hands {
		if hand.Seat < 1 || hand.Seat > maxPlayers || seen[hand.Seat] || hand.CardCount < 1 {
			return ErrInvalidStartingHands
		}
		seen[hand.Seat] = true
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, id uint16) error {
	return s.store.DeleteVariant(ctx, id)
}
