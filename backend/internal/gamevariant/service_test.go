package gamevariant

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeStore struct {
	variants     []GameVariant
	listCalled   bool
	listErr      error
	inserted     GameVariant
	insertCalled bool
	insertErr    error
	deleteID     uint16
	deleteCalled bool
	deleteErr    error
}

func (f *fakeStore) ListVariants(ctx context.Context) ([]GameVariant, error) {
	f.listCalled = true
	return f.variants, f.listErr
}

func (f *fakeStore) FindByPlayerCount(ctx context.Context, playerCount uint16) (uint16, error) {
	return 0, nil
}

func (f *fakeStore) InsertVariant(ctx context.Context, variant GameVariant) (*GameVariant, error) {
	f.insertCalled = true
	f.inserted = variant
	if f.insertErr != nil {
		return nil, f.insertErr
	}
	variant.Id = 7
	return &variant, nil
}

func (f *fakeStore) DeleteVariant(ctx context.Context, id uint16) error {
	f.deleteCalled = true
	f.deleteID = id
	return f.deleteErr
}

func TestCreateVariant(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store)

	created, err := service.Create(context.Background(), GameVariant{
		Slug:               "base_3_5",
		Name:               "  Base 3-5 players  ",
		MinPlayers:         3,
		MaxPlayers:         5,
		DeckId:             2,
		EndOfTurnDrawCount: 3,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.Id != 7 {
		t.Errorf("Id = %d, want 7", created.Id)
	}
	if !store.insertCalled {
		t.Fatal("InsertVariant was not called")
	}
	if store.inserted.Name != "Base 3-5 players" {
		t.Errorf("inserted name = %q, want %q", store.inserted.Name, "Base 3-5 players")
	}
	if store.inserted.Slug != "base_3_5" {
		t.Errorf("inserted slug = %q, want %q", store.inserted.Slug, "base_3_5")
	}
	if store.inserted.MinPlayers != 3 || store.inserted.MaxPlayers != 5 {
		t.Errorf("inserted players = %d-%d, want 3-5", store.inserted.MinPlayers, store.inserted.MaxPlayers)
	}
	if store.inserted.DeckId != 2 {
		t.Errorf("inserted deck id = %d, want 2", store.inserted.DeckId)
	}
	if store.inserted.EndOfTurnDrawCount != 3 {
		t.Errorf("inserted draw count = %d, want 3", store.inserted.EndOfTurnDrawCount)
	}
}

func TestCreateVariantValidation(t *testing.T) {
	valid := GameVariant{
		Slug:               "base_3_5",
		Name:               "Base",
		MinPlayers:         3,
		MaxPlayers:         5,
		DeckId:             1,
		EndOfTurnDrawCount: 3,
	}

	tests := []struct {
		name    string
		mutate  func(*GameVariant)
		wantErr error
	}{
		{
			name:    "empty slug",
			mutate:  func(v *GameVariant) { v.Slug = "" },
			wantErr: ErrInvalidSlug,
		},
		{
			name:    "uppercase slug",
			mutate:  func(v *GameVariant) { v.Slug = "Base" },
			wantErr: ErrInvalidSlug,
		},
		{
			name:    "slug with dash",
			mutate:  func(v *GameVariant) { v.Slug = "base-3-5" },
			wantErr: ErrInvalidSlug,
		},
		{
			name:    "slug too long",
			mutate:  func(v *GameVariant) { v.Slug = strings.Repeat("a", 41) },
			wantErr: ErrInvalidSlug,
		},
		{
			name:    "empty name",
			mutate:  func(v *GameVariant) { v.Name = "" },
			wantErr: ErrInvalidName,
		},
		{
			name:    "whitespace name",
			mutate:  func(v *GameVariant) { v.Name = "   " },
			wantErr: ErrInvalidName,
		},
		{
			name:    "name too long",
			mutate:  func(v *GameVariant) { v.Name = strings.Repeat("a", 81) },
			wantErr: ErrInvalidName,
		},
		{
			name:    "zero min players",
			mutate:  func(v *GameVariant) { v.MinPlayers = 0 },
			wantErr: ErrInvalidPlayers,
		},
		{
			name:    "max below min",
			mutate:  func(v *GameVariant) { v.MaxPlayers = 2 },
			wantErr: ErrInvalidPlayers,
		},
		{
			name:    "zero draw count",
			mutate:  func(v *GameVariant) { v.EndOfTurnDrawCount = 0 },
			wantErr: ErrInvalidDrawCount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{}
			service := NewService(store)

			variant := valid
			tt.mutate(&variant)

			_, err := service.Create(context.Background(), variant)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Create() error = %v, want %v", err, tt.wantErr)
			}
			if store.insertCalled {
				t.Error("InsertVariant was called for invalid input")
			}
		})
	}
}

func TestCreateVariantStoreError(t *testing.T) {
	storeErr := errors.New("boom")
	service := NewService(&fakeStore{insertErr: storeErr})

	_, err := service.Create(context.Background(), GameVariant{
		Slug:               "base_3_5",
		Name:               "Base",
		MinPlayers:         3,
		MaxPlayers:         5,
		DeckId:             1,
		EndOfTurnDrawCount: 3,
	})
	if !errors.Is(err, storeErr) {
		t.Fatalf("Create() error = %v, want %v", err, storeErr)
	}
}

func TestListVariants(t *testing.T) {
	want := []GameVariant{{Id: 1, Slug: "base_3_5"}, {Id: 2, Slug: "two_player"}}
	store := &fakeStore{variants: want}
	service := NewService(store)

	got, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if !store.listCalled {
		t.Fatal("ListVariants was not called")
	}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("List() = %v, want %v", got, want)
	}
}

func TestListVariantsError(t *testing.T) {
	storeErr := errors.New("boom")
	service := NewService(&fakeStore{listErr: storeErr})

	_, err := service.List(context.Background())
	if !errors.Is(err, storeErr) {
		t.Fatalf("List() error = %v, want %v", err, storeErr)
	}
}

func TestDeleteVariant(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store)

	if err := service.Delete(context.Background(), 42); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if !store.deleteCalled {
		t.Fatal("DeleteVariant was not called")
	}
	if store.deleteID != 42 {
		t.Errorf("deleted id = %d, want 42", store.deleteID)
	}
}

func TestDeleteVariantError(t *testing.T) {
	service := NewService(&fakeStore{deleteErr: ErrNotFound})

	err := service.Delete(context.Background(), 42)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Delete() error = %v, want %v", err, ErrNotFound)
	}
}
