package deck

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeStore struct {
	insertSlug   string
	insertBeans  []DeckBean
	insertCalled bool
	deleteID     uint16
	deleteCalled bool
	deleteErr    error
	insertErr    error
}

func (f *fakeStore) InsertDeck(ctx context.Context, slug string, beans []DeckBean) (*Deck, error) {
	f.insertCalled = true
	f.insertSlug = slug
	f.insertBeans = beans
	if f.insertErr != nil {
		return nil, f.insertErr
	}
	return &Deck{Id: 3, Slug: slug, Beans: beans}, nil
}

func (f *fakeStore) DeleteDeck(ctx context.Context, id uint16) error {
	f.deleteCalled = true
	f.deleteID = id
	return f.deleteErr
}

func TestCreateDeck(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store)

	beans := []DeckBean{
		{BeanId: 1, CardCount: 8},
		{BeanId: 2, CardCount: 6},
	}

	created, err := service.Create(context.Background(), "base_3_5", beans)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.Id != 3 {
		t.Errorf("Id = %d, want 3", created.Id)
	}
	if created.Slug != "base_3_5" {
		t.Errorf("Slug = %q, want %q", created.Slug, "base_3_5")
	}
	if len(created.Beans) != 2 {
		t.Errorf("Beans length = %d, want 2", len(created.Beans))
	}
	if !store.insertCalled {
		t.Fatal("InsertDeck was not called")
	}
	if store.insertSlug != "base_3_5" {
		t.Errorf("inserted slug = %q, want %q", store.insertSlug, "base_3_5")
	}
	if len(store.insertBeans) != 2 || store.insertBeans[0] != beans[0] || store.insertBeans[1] != beans[1] {
		t.Errorf("inserted beans = %v, want %v", store.insertBeans, beans)
	}
}

func TestCreateDeckValidation(t *testing.T) {
	validBeans := []DeckBean{{BeanId: 1, CardCount: 8}}

	tests := []struct {
		name  string
		slug  string
		beans []DeckBean
		want  error
	}{
		{
			name:  "empty slug",
			slug:  "",
			beans: validBeans,
			want:  ErrInvalidSlug,
		},
		{
			name:  "uppercase slug",
			slug:  "Base",
			beans: validBeans,
			want:  ErrInvalidSlug,
		},
		{
			name:  "slug with dash",
			slug:  "base-3-5",
			beans: validBeans,
			want:  ErrInvalidSlug,
		},
		{
			name:  "slug too long",
			slug:  strings.Repeat("a", 41),
			beans: validBeans,
			want:  ErrInvalidSlug,
		},
		{
			name:  "no beans",
			slug:  "base_3_5",
			beans: nil,
			want:  ErrNoBeans,
		},
		{
			name:  "zero card count",
			slug:  "base_3_5",
			beans: []DeckBean{{BeanId: 1, CardCount: 0}},
			want:  ErrInvalidCardCount,
		},
		{
			name:  "duplicate bean",
			slug:  "base_3_5",
			beans: []DeckBean{{BeanId: 1, CardCount: 4}, {BeanId: 1, CardCount: 5}},
			want:  ErrDuplicateBean,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{}
			service := NewService(store)

			_, err := service.Create(context.Background(), tt.slug, tt.beans)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Create() error = %v, want %v", err, tt.want)
			}
			if store.insertCalled {
				t.Error("InsertDeck was called for invalid input")
			}
		})
	}
}

func TestCreateDeckStoreError(t *testing.T) {
	storeErr := errors.New("boom")
	service := NewService(&fakeStore{insertErr: storeErr})

	_, err := service.Create(context.Background(), "base_3_5", []DeckBean{{BeanId: 1, CardCount: 8}})
	if !errors.Is(err, storeErr) {
		t.Fatalf("Create() error = %v, want %v", err, storeErr)
	}
}

func TestDeleteDeck(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store)

	if err := service.Delete(context.Background(), 42); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if !store.deleteCalled {
		t.Fatal("DeleteDeck was not called")
	}
	if store.deleteID != 42 {
		t.Errorf("deleted id = %d, want 42", store.deleteID)
	}
}

func TestDeleteDeckError(t *testing.T) {
	service := NewService(&fakeStore{deleteErr: ErrNotFound})

	err := service.Delete(context.Background(), 42)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Delete() error = %v, want %v", err, ErrNotFound)
	}
}
