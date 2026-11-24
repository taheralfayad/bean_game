package gamevariant

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row
}

type Store struct {
	db *sql.DB
	q  DBTX
}

func NewStore(db *sql.DB) *Store { return &Store{db: db, q: db} }

func (s *Store) WithTx(ctx context.Context, fn func(tx *Store) error) (err error) {
	sqlTx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = sqlTx.Rollback()
			panic(p)
		}
		if err != nil {
			if rbErr := sqlTx.Rollback(); rbErr != nil {
				err = errors.Join(err, rbErr)
			}
		}
	}()

	if err = fn(&Store{q: sqlTx}); err != nil {
		return err
	}
	return sqlTx.Commit()
}

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

func (s *Store) FindByPlayerCount(ctx context.Context, playerCount uint16) (*GameVariant, error) {
	var variant GameVariant
	err := s.db.QueryRowContext(ctx, `
		SELECT id, slug, name, min_players, max_players, deck_id, end_of_turn_draw_count
		FROM game_variants
		WHERE $1 BETWEEN min_players AND max_players
	`, playerCount).Scan(
		&variant.Id,
		&variant.Slug,
		&variant.Name,
		&variant.MinPlayers,
		&variant.MaxPlayers,
		&variant.DeckId,
		&variant.EndOfTurnDrawCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoMatch
	}
	if err != nil {
		return nil, fmt.Errorf("find game variant: %w", err)
	}

	return &variant, nil
}

func (s *Store) insertVariant(ctx context.Context, variant GameVariant) (*GameVariant, error) {
	err := s.q.QueryRowContext(ctx, `
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
			return nil, ErrRangeOverlap
		}
		return nil, fmt.Errorf("insert game variant: %w", err)
	}

	return &variant, nil
}

func (s *Store) InsertVariant(ctx context.Context, variant GameVariant, hands []StartingHand) (*GameVariant, error) {
	var created *GameVariant
	err := s.WithTx(ctx, func(tx *Store) error {
		var err error
		created, err = tx.insertVariant(ctx, variant)
		if err != nil {
			return err
		}

		for _, hand := range hands {
			if err := tx.insertStartingHand(ctx, created.Id, hand.Seat, hand.CardCount); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return created, nil
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

func (s *Store) FindStartingHandBySeat(ctx context.Context, variantId, seat uint16) (uint16, error) {
	var cardCount uint16
	err := s.db.QueryRowContext(ctx, `
		SELECT
			card_count
		FROM game_variant_starting_hands
		WHERE variant_id = $1 AND seat = $2
	`, variantId, seat).Scan(&cardCount)

	if err != nil {
		return cardCount, err
	}

	return cardCount, nil
}

func (s *Store) insertStartingHand(ctx context.Context, variantId, seat, cardCount uint16) error {
	_, err := s.q.ExecContext(ctx, `
		INSERT INTO game_variant_starting_hands (variant_id, seat, card_count)
		VALUES ($1, $2, $3)
	`, variantId, seat, cardCount)
	if err != nil {
		return fmt.Errorf("insert game variant starting hand: %w", err)
	}

	return nil
}
