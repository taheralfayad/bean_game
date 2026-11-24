package user

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) GetUser(ctx context.Context, username string) (*User, string, error) {
	var user User
	var hashedPassword string

	err := s.db.QueryRowContext(ctx, `
		SELECT id, username, password, role
		FROM users
		WHERE username = $1
		LIMIT 1
	`, username).Scan(
		&user.Id,
		&user.Username,
		&hashedPassword,
		&user.Role,
	)
	if err != nil {
		return nil, "", fmt.Errorf("fetch user: %w", err)
	}

	return &user, hashedPassword, nil
}

func (s *Store) InsertUser(ctx context.Context, username, hashedPassword string, role Role) (*User, error) {
	var user User
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO users (username, password, role)
		VALUES ($1, $2, $3)
		RETURNING id, username, role
	`, username, hashedPassword, role).Scan(&user.Id, &user.Username, &user.Role)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return &user, nil
}
