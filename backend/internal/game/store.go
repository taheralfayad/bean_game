package game

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrGameNotFound    = errors.New("game not found")
	ErrGameStarted     = errors.New("game already started")
	ErrGameStarting    = errors.New("game already starting")
	ErrGameNotStarting = errors.New("game not starting")
	ErrPlayersChanged  = errors.New("player count changed")
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

func (s *Store) CreateGame(ctx context.Context) (*Game, *Player, error) {
	var (
		game   *Game
		player *Player
	)

	err := s.WithTx(ctx, func(tx *Store) error {
		var err error

		if game, err = tx.insertGame(ctx); err != nil {
			return err
		}

		playerId, err := tx.insertPlayer(ctx)
		if err != nil {
			return err
		}

		const hostSeat = 1
		if err := tx.insertGamePlayer(ctx, game.Id, playerId, hostSeat); err != nil {
			return err
		}

		player = &Player{Id: playerId, Seat: hostSeat}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return game, player, nil
}

func (s *Store) JoinGame(ctx context.Context, gameId uint64) (*Player, error) {
	var player *Player

	err := s.WithTx(ctx, func(tx *Store) error {
		state, err := tx.lockGameState(ctx, gameId)
		if err != nil {
			return err
		}
		if state != StateLobby {
			return ErrGameStarted
		}

		seat, err := tx.nextSeat(ctx, gameId)
		if err != nil {
			return err
		}

		playerId, err := tx.insertPlayer(ctx)
		if err != nil {
			return err
		}

		if err := tx.insertGamePlayer(ctx, gameId, playerId, seat); err != nil {
			return err
		}

		player = &Player{Id: playerId, Seat: seat}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return player, nil
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

func (s *Store) ListPlayers(ctx context.Context, gameId uint64) ([]Player, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT player_id, seat
		FROM games_players
		WHERE game_id = $1
		ORDER BY seat
	`, gameId)
	if err != nil {
		return nil, fmt.Errorf("list game players: %w", err)
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var player Player
		if err := rows.Scan(&player.Id, &player.Seat); err != nil {
			return nil, fmt.Errorf("scan game player: %w", err)
		}
		players = append(players, player)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate game players: %w", err)
	}

	return players, nil
}

func (s *Store) StartGame(ctx context.Context, gameId uint64, variantId uint16, playerCards map[uint64][]uint16, remainingDeck []uint16) (*Game, error) {
	var game Game
	err := s.WithTx(ctx, func(tx *Store) error {
		if err := tx.insertCardsAtGameStart(ctx, gameId, playerCards, remainingDeck); err != nil {
			return err
		}

		if err := tx.q.QueryRowContext(ctx, `
			UPDATE games
			SET variant_id = $2, state = 'started'
			WHERE id = $1
			RETURNING id, shuffle_number, current_seat, current_phase, variant_id, state
		`, gameId, variantId).Scan(
			&game.Id,
			&game.ShuffleNumber,
			&game.CurrentSeat,
			&game.CurrentPhase,
			&game.VariantId,
			&game.State,
		); err != nil {
			return fmt.Errorf("start game: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &game, nil
}

func (s *Store) BeginGameStart(ctx context.Context, gameId uint64) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE games
		SET state = 'starting'
		WHERE id = $1 AND state = 'lobby'
	`, gameId)
	if err != nil {
		return fmt.Errorf("begin game start: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("begin game start rows affected: %w", err)
	}
	if affected > 0 {
		return nil
	}

	state, err := s.GameState(ctx, gameId)
	if err != nil {
		return err
	}
	if state == StateStarted {
		return ErrGameStarted
	}

	return ErrGameStarting
}

func (s *Store) AbortGameStart(ctx context.Context, gameId uint64) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE games
		SET state = 'lobby'
		WHERE id = $1 AND state = 'starting'
	`, gameId)
	if err != nil {
		return fmt.Errorf("abort game start: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("abort game start rows affected: %w", err)
	}
	if affected > 0 {
		return nil
	}

	state, err := s.GameState(ctx, gameId)
	if err != nil {
		return err
	}
	if state == StateStarted {
		return ErrGameStarted
	}

	return nil
}

func (s *Store) GameState(ctx context.Context, gameId uint64) (State, error) {
	var state State
	if err := s.db.QueryRowContext(ctx, `
		SELECT state
		FROM games
		WHERE id = $1
	`, gameId).Scan(&state); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrGameNotFound
		}
		return "", fmt.Errorf("find game state: %w", err)
	}

	return state, nil
}

func (s *Store) insertCardsAtGameStart(ctx context.Context, gameId uint64, playerCards map[uint64][]uint16, remainingDeck []uint16) error {
	for playerId, cards := range playerCards {
		for position, beanId := range cards {
			if err := s.insertPlayerCard(ctx, gameId, playerId, beanId, position); err != nil {
				return fmt.Errorf("insert cards: %w", err)
			}
		}
	}

	for position, beanId := range remainingDeck {
		if err := s.insertRemainingDeck(ctx, gameId, beanId, position); err != nil {
			return fmt.Errorf("insert cards: %w", err)
		}
	}

	return nil
}

func (s *Store) insertPlayerCard(ctx context.Context, gameId, playerId uint64, beanId uint16, position int) error {
	if _, err := s.q.ExecContext(ctx, `
		INSERT INTO cards (
			game_id,
			bean_id,
			location,
			position,
			player_id
		)
		VALUES ($1, $2, $3, $4, $5)
	`, gameId, beanId, LocationHand, position, playerId); err != nil {
		return fmt.Errorf("insert player card: %w", err)
	}

	return nil
}

func (s *Store) insertRemainingDeck(ctx context.Context, gameId uint64, beanId uint16, position int) error {
	if _, err := s.q.ExecContext(ctx, `
		INSERT INTO cards (
			game_id,
			bean_id,
			location,
			position
		)
		VALUES ($1, $2, $3, $4)
	`, gameId, beanId, LocationDraw, position); err != nil {
		return fmt.Errorf("insert remaining deck: %w", err)
	}

	return nil

}

func (s *Store) insertGame(ctx context.Context) (*Game, error) {
	var game Game
	if err := s.q.QueryRowContext(ctx, `
		INSERT INTO games DEFAULT VALUES
		RETURNING id, shuffle_number, current_seat, current_phase, variant_id, state
	`).Scan(
		&game.Id,
		&game.ShuffleNumber,
		&game.CurrentSeat,
		&game.CurrentPhase,
		&game.VariantId,
		&game.State,
	); err != nil {
		return nil, fmt.Errorf("insert game: %w", err)
	}
	return &game, nil
}

func (s *Store) insertGamePlayer(ctx context.Context, gameId, playerId uint64, seat uint16) error {
	if _, err := s.q.ExecContext(ctx, `
		INSERT INTO games_players (game_id, player_id, seat)
		VALUES ($1, $2, $3)
	`, gameId, playerId, seat); err != nil {
		return fmt.Errorf("insert game player: %w", err)
	}
	return nil
}

func (s *Store) insertPlayer(ctx context.Context) (uint64, error) {
	var id uint64
	err := s.q.QueryRowContext(ctx, `
		INSERT INTO players DEFAULT VALUES
		RETURNING id
	`).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert player: %w", err)
	}

	return id, nil
}

func (s *Store) lockGameState(ctx context.Context, gameId uint64) (State, error) {
	var state State
	if err := s.q.QueryRowContext(ctx, `
		SELECT state
		FROM games
		WHERE id = $1
		FOR UPDATE
	`, gameId).Scan(&state); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrGameNotFound
		}
		return "", fmt.Errorf("lock game: %w", err)
	}
	return state, nil
}

func (s *Store) nextSeat(ctx context.Context, gameId uint64) (uint16, error) {
	var seat uint16
	if err := s.q.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(seat), 0) + 1
		FROM games_players
		WHERE game_id = $1
	`, gameId).Scan(&seat); err != nil {
		return 0, fmt.Errorf("next seat: %w", err)
	}
	return seat, nil
}
