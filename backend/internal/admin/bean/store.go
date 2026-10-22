package bean

import (
	"database/sql"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) InsertBean(name string, imageFileID *string) (*Bean, error) {
	var newId uint16

	err := s.db.QueryRow(
		`
			INSERT INTO beans (name, image_file_id)
			VALUES (
				$1, $2
			)
			RETURNING id 
		`,
		name,
		imageFileID,
	).Scan(&newId)

	if err != nil {
		return &Bean{}, err
	}

	return &Bean{
		Id:          newId,
		Name:        name,
		ImageFileID: imageFileID,
	}, nil
}
