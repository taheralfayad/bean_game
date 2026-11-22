package user

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) GetUser(ctx context.Context, username string, password []byte) (*User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var user User
	var hashed_password []byte

	err = tx.QueryRowContext(ctx, `
		SELECT id, username, password, role
		FROM users
		WHERE username = $1
		LIMIT 1
	`, username).Scan(
		&user.Id,
		&user.Username,
		hashed_password,
		&user.Role,
	)

	if err != nil || !comparePasswords(string(hashed_password), password) {
		return nil, fmt.Errorf("Invalid username or password")
	}

	token, err := createToken(username)
	if err != nil {
		return nil, fmt.Errorf("Fetch user: %w", err)
	}

	return &User{
		Id:       user.Id,
		Username: user.Username,
		Token:    token,
		Role:     user.Role,
	}, nil

}

func (s *Store) InsertUser(ctx context.Context, username string, password []byte, role Role) (*User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var newId uint16
	hashed_password := hashAndSalt([]byte(password))

	err = tx.QueryRowContext(ctx, `
		INSERT INTO users (username, password, role)
		VALUES ($1, $2, $3)
		RETURNING id
	`, username, hashed_password, role).Scan(&newId)
	if err != nil {
		return nil, fmt.Errorf("Insert User: %w", err)
	}

	token, err := createToken(username)
	if err != nil {
		return nil, fmt.Errorf("Insert User: %w", err)
	}

	return &User{
		Id:       newId,
		Username: username,
		Token:    token,
		Role:     UserRole,
	}, nil
}
