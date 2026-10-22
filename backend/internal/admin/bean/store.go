package bean

import (
	"database/sql"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) InsertBean(name string, imageLink string) (*Bean, error) {
	var newId uint16

	err := s.db.QueryRow(
		`
			INSERT INTO beans (name, image_link)
			VALUES (
				$1, $2
			)
			RETURNING id 
		`,
		name,
		imageLink,
	).Scan(&newId)

	if err != nil {
		return &Bean{}, err
	}

	return &Bean{
		Id:        newId,
		Name:      name,
		ImageLink: imageLink,
	}, nil

}
