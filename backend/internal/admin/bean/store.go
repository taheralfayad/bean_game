package bean

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) InsertBean(ctx context.Context, name string, imageKey *string, payouts []Payout) (*Bean, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var newId uint16

	err = tx.QueryRowContext(ctx, `
		INSERT INTO beans (name, image_key)
		VALUES ($1, $2)
		RETURNING id
	`, name, imageKey).Scan(&newId)
	if err != nil {
		return nil, fmt.Errorf("insert bean: %w", err)
	}

	for _, payout := range payouts {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO beans_payout (bean_id, num_cards, coins)
			VALUES ($1, $2, $3)
		`, newId, payout.NumCards, payout.Coins); err != nil {
			return nil, fmt.Errorf("insert bean payout: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit bean: %w", err)
	}

	return &Bean{
		Id:       newId,
		Name:     name,
		ImageKey: imageKey,
		Payouts:  payouts,
	}, nil
}

func (s *Store) DeleteBean(ctx context.Context, id uint16) (string, error) {
	var imageKey sql.NullString

	err := s.db.QueryRowContext(ctx, `
		DELETE FROM beans
		WHERE id = $1
		RETURNING image_key
	`, id).Scan(&imageKey)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("delete bean: %w", err)
	}

	return imageKey.String, nil
}
