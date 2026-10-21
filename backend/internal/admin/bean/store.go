package bean

import (
	"database/sql"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) AddBeanCard(id int16) (*Bean, error) {
	return &Bean{}, nil
}
