package gamevariant

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) ListVariants(ctx context.Context) ([]GameVariant, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, slug, name, min_players, max_players, deck_id, end_of_turn_draw_count
		FROM game_variants
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("list game variants: %w", err)
	}
	defer rows.Close()

	var variants []GameVariant
	for rows.Next() {
		var variant GameVariant
		if err := rows.Scan(
			&variant.Id,
			&variant.Slug,
			&variant.Name,
			&variant.MinPlayers,
			&variant.MaxPlayers,
			&variant.DeckId,
			&variant.EndOfTurnDrawCount,
		); err != nil {
			return nil, fmt.Errorf("scan game variant: %w", err)
		}
		variants = append(variants, variant)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate game variants: %w", err)
	}

	return variants, nil
}

func (s *Store) InsertVariant(ctx context.Context, variant GameVariant) (*GameVariant, error) {
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO game_variants (slug, name, min_players, max_players, deck_id, end_of_turn_draw_count)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`,
		variant.Slug,
		variant.Name,
		variant.MinPlayers,
		variant.MaxPlayers,
		variant.DeckId,
		variant.EndOfTurnDrawCount,
	).Scan(&variant.Id)
	if err != nil {
		return nil, fmt.Errorf("insert game variant: %w", err)
	}

	return &variant, nil
}

func (s *Store) DeleteVariant(ctx context.Context, id uint16) error {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM game_variants
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete game variant: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete game variant rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}
