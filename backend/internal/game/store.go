package game

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrGameNotFound = errors.New("game not found")
	ErrGameStarted  = errors.New("game already started")
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) CreateGame(ctx context.Context) (*Game, *Player, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var game Game
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO games DEFAULT VALUES
		RETURNING id, shuffle_number, current_seat, current_phase, variant_id
	`).Scan(
		&game.Id,
		&game.ShuffleNumber,
		&game.CurrentSeat,
		&game.CurrentPhase,
		&game.VariantId,
	); err != nil {
		return nil, nil, fmt.Errorf("insert game: %w", err)
	}

	playerId, err := insertPlayer(ctx, tx)
	if err != nil {
		return nil, nil, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO games_players (game_id, player_id, seat)
		VALUES ($1, $2, $3)
	`, game.Id, playerId, 1); err != nil {
		return nil, nil, fmt.Errorf("insert game player: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("commit game: %w", err)
	}

	return &game, &Player{Id: playerId}, nil
}

func (s *Store) JoinGame(ctx context.Context, gameId uint64) (*Player, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var variantId *uint16
	if err := tx.QueryRowContext(ctx, `
		SELECT variant_id
		FROM games
		WHERE id = $1
		FOR UPDATE
	`, gameId).Scan(&variantId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGameNotFound
		}
		return nil, fmt.Errorf("lock game: %w", err)
	}
	if variantId != nil {
		return nil, ErrGameStarted
	}

	var seat uint16
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(seat), 0) + 1
		FROM games_players
		WHERE game_id = $1
	`, gameId).Scan(&seat); err != nil {
		return nil, fmt.Errorf("next seat: %w", err)
	}

	playerId, err := insertPlayer(ctx, tx)
	if err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO games_players (game_id, player_id, seat)
		VALUES ($1, $2, $3)
	`, gameId, playerId, seat); err != nil {
		return nil, fmt.Errorf("insert game player: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit game: %w", err)
	}

	return &Player{Id: playerId}, nil
}

func (s *Store) CountPlayers(ctx context.Context, gameId uint64) (uint16, error) {
	var count uint16
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM games_players
		WHERE game_id = $1
	`, gameId).Scan(&count); err != nil {
		return 0, fmt.Errorf("count game players: %w", err)
	}

	return count, nil
}

func (s *Store) StartGame(ctx context.Context, gameId uint64, variantId uint16) (*Game, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var existingVariantId *uint16
	if err := tx.QueryRowContext(ctx, `
		SELECT variant_id
		FROM games
		WHERE id = $1
		FOR UPDATE
	`, gameId).Scan(&existingVariantId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGameNotFound
		}
		return nil, fmt.Errorf("lock game: %w", err)
	}
	if existingVariantId != nil {
		return nil, ErrGameStarted
	}

	var game Game
	if err := tx.QueryRowContext(ctx, `
		UPDATE games
		SET variant_id = $2
		WHERE id = $1
		RETURNING id, shuffle_number, current_seat, current_phase, variant_id
	`, gameId, variantId).Scan(
		&game.Id,
		&game.ShuffleNumber,
		&game.CurrentSeat,
		&game.CurrentPhase,
		&game.VariantId,
	); err != nil {
		return nil, fmt.Errorf("start game: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit game: %w", err)
	}

	return &game, nil
}
