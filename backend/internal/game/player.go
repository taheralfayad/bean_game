package game

import (
	"context"
	"database/sql"
	"fmt"
)

type Player struct {
	Id   int64  `json:"id"`
	Seat uint16 `json:"seat"`
}

func insertPlayer(ctx context.Context, tx *sql.Tx) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO players DEFAULT VALUES
		RETURNING id
	`).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert player: %w", err)
	}

	return id, nil
}
