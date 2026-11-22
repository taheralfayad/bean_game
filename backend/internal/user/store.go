package user

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) InsertUser(ctx context.Context, username string, password []byte) (*User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var newId uint16
	hashed_password := hashAndSalt([]byte(password))

	err = tx.QueryRowContext(ctx, `
		INSERT INTO users (username, password)
		VALUES ($1, $2)
		RETURNING id
	`, username, hashed_password).Scan(&newId)
	if err != nil {
		return nil, fmt.Errorf("Insert User: %w", err)
	}

	jwt, err := createToken(username)
	if err != nil {
		return nil, fmt.Errorf("Insert User: %w", err)
	}

	return &User{
		Id:       newId,
		Username: username,
		Token:    jwt,
		Role:     UserRole,
	}, nil
}
