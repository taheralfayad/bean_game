package deck

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) InsertDeck(ctx context.Context, slug string, beans []DeckBean) (*Deck, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var id uint16
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO decks (slug)
		VALUES ($1)
		RETURNING id
	`, slug).Scan(&id); err != nil {
		return nil, fmt.Errorf("insert deck: %w", err)
	}

	for _, bean := range beans {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO beans_decks (deck_id, bean_id, card_count)
			VALUES ($1, $2, $3)
		`, id, bean.BeanId, bean.CardCount); err != nil {
			return nil, fmt.Errorf("insert deck bean: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit deck: %w", err)
	}

	return &Deck{Id: id, Slug: slug, Beans: beans}, nil
}

func (s *Store) DeleteDeck(ctx context.Context, id uint16) error {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM decks
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete deck: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete deck rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}
