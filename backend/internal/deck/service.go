package deck

import (
	"context"
	"errors"
	"regexp"
)

var (
	ErrNotFound         = errors.New("deck not found")
	ErrInvalidSlug      = errors.New("slug must match ^[a-z0-9_]+$ and be at most 40 characters")
	ErrNoBeans          = errors.New("deck must contain at least one bean")
	ErrInvalidCardCount = errors.New("card count must be greater than zero")
	ErrDuplicateBean    = errors.New("deck contains a duplicate bean")
)

var slugPattern = regexp.MustCompile(`^[a-z0-9_]+$`)

const maxSlugLength = 40

type deckStore interface {
	InsertDeck(ctx context.Context, slug string, beans []DeckBean) (*Deck, error)
	DeleteDeck(ctx context.Context, id uint16) error
}

type Service struct {
	store deckStore
}

func NewService(store deckStore) *Service {
	return &Service{store: store}
}

func (s *Service) Create(ctx context.Context, slug string, beans []DeckBean) (*Deck, error) {
	if len(slug) > maxSlugLength || !slugPattern.MatchString(slug) {
		return nil, ErrInvalidSlug
	}
	if len(beans) == 0 {
		return nil, ErrNoBeans
	}

	seen := make(map[uint16]struct{}, len(beans))
	for _, bean := range beans {
		if bean.CardCount == 0 {
			return nil, ErrInvalidCardCount
		}
		if _, ok := seen[bean.BeanId]; ok {
			return nil, ErrDuplicateBean
		}
		seen[bean.BeanId] = struct{}{}
	}

	return s.store.InsertDeck(ctx, slug, beans)
}

func (s *Service) Delete(ctx context.Context, id uint16) error {
	return s.store.DeleteDeck(ctx, id)
}
