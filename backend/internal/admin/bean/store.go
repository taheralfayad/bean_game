package bean

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) InsertBean(ctx context.Context, name string, imageKey *string) (*Bean, error) {
	var newId uint16

	err := s.db.QueryRowContext(ctx, `
		INSERT INTO beans (name, image_key)
		VALUES ($1, $2)
		RETURNING id
	`, name, imageKey).Scan(&newId)
	if err != nil {
		return nil, fmt.Errorf("insert bean: %w", err)
	}

	return &Bean{
		Id:       newId,
		Name:     name,
		ImageKey: imageKey,
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
