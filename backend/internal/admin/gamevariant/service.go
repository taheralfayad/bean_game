package gamevariant

import (
	"context"
	"errors"
	"regexp"
	"strings"
)

var (
	ErrNotFound         = errors.New("game variant not found")
	ErrInvalidSlug      = errors.New("slug must match ^[a-z0-9_]+$ and be at most 40 characters")
	ErrInvalidName      = errors.New("name must be between 1 and 80 characters")
	ErrInvalidPlayers   = errors.New("min_players must be at least 1 and max_players must be at least min_players")
	ErrInvalidDrawCount = errors.New("end_of_turn_draw_count must be greater than zero")
)

var slugPattern = regexp.MustCompile(`^[a-z0-9_]+$`)

const (
	maxSlugLength = 40
	maxNameLength = 80
)

type gamevariantStore interface {
	ListVariants(ctx context.Context) ([]GameVariant, error)
	InsertVariant(ctx context.Context, variant GameVariant) (*GameVariant, error)
	DeleteVariant(ctx context.Context, id uint16) error
}

type Service struct {
	store gamevariantStore
}

func NewService(store gamevariantStore) *Service {
	return &Service{store: store}
}

func (s *Service) List(ctx context.Context) ([]GameVariant, error) {
	return s.store.ListVariants(ctx)
}

func (s *Service) Create(ctx context.Context, variant GameVariant) (*GameVariant, error) {
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

	return s.store.InsertVariant(ctx, variant)
}

func (s *Service) Delete(ctx context.Context, id uint16) error {
	return s.store.DeleteVariant(ctx, id)
}
