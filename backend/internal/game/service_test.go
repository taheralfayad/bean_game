package game

import (
	"context"
	"errors"
	"testing"

	"github.com/taheralfayad/bean_game/backend/internal/gameevents"
	"github.com/taheralfayad/bean_game/backend/internal/gamevariant"
)

type fakeStore struct {
	createdGame    *Game
	createdPlayer  *Player
	createErr      error
	createCalled   bool
	playerCount    uint16
	countErr       error
	countCalled    bool
	startCalled    bool
	startedGameId  uint64
	startedVariant uint16
	startErr       error
}

func (f *fakeStore) CreateGame(ctx context.Context) (*Game, *Player, error) {
	f.createCalled = true
	if f.createErr != nil {
		return nil, nil, f.createErr
	}
	return f.createdGame, f.createdPlayer, nil
}

func (f *fakeStore) CountPlayers(ctx context.Context, gameId uint64) (uint16, error) {
	f.countCalled = true
	return f.playerCount, f.countErr
}

func (f *fakeStore) StartGame(ctx context.Context, gameId uint64, variantId uint16) (*Game, error) {
	f.startCalled = true
	f.startedGameId = gameId
	f.startedVariant = variantId
	if f.startErr != nil {
		return nil, f.startErr
	}
	return &Game{Id: gameId, VariantId: &variantId}, nil
}

type fakeVariants struct {
	id          uint16
	err         error
	playerCount uint16
	called      bool
}

func (f *fakeVariants) FindByPlayerCount(ctx context.Context, playerCount uint16) (uint16, error) {
	f.called = true
	f.playerCount = playerCount
	return f.id, f.err
}

type fakePublisher struct {
	events []gameevents.Event
}

func (f *fakePublisher) Publish(event gameevents.Event) {
	f.events = append(f.events, event)
}

func TestCreateGame(t *testing.T) {
	wantGame := &Game{Id: 7}
	wantPlayer := &Player{Id: 8}
	store := &fakeStore{createdGame: wantGame, createdPlayer: wantPlayer}
	service := NewService(store, &fakeVariants{}, &fakePublisher{})

	game, player, err := service.CreateGame(context.Background())
	if err != nil {
		t.Fatalf("CreateGame() error = %v", err)
	}
	if !store.createCalled {
		t.Fatal("CreateGame was not called")
	}
	if game != wantGame {
		t.Errorf("game = %v, want %v", game, wantGame)
	}
	if player != wantPlayer {
		t.Errorf("player = %v, want %v", player, wantPlayer)
	}
}

func TestCreateGameError(t *testing.T) {
	want := errors.New("db down")
	store := &fakeStore{createErr: want}
	service := NewService(store, &fakeVariants{}, &fakePublisher{})

	_, _, err := service.CreateGame(context.Background())
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}

func TestStartGame(t *testing.T) {
	store := &fakeStore{playerCount: 4}
	variants := &fakeVariants{id: 9}
	publisher := &fakePublisher{}
	service := NewService(store, variants, publisher)

	game, err := service.StartGame(context.Background(), 12)
	if err != nil {
		t.Fatalf("StartGame() error = %v", err)
	}

	if game.Id != 12 {
		t.Errorf("Id = %d, want 12", game.Id)
	}
	if game.VariantId == nil || *game.VariantId != 9 {
		t.Errorf("VariantId = %v, want 9", game.VariantId)
	}
	if variants.playerCount != 4 {
		t.Errorf("variant lookup count = %d, want 4", variants.playerCount)
	}
	if !store.startCalled {
		t.Fatal("StartGame was not called")
	}
	if store.startedVariant != 9 {
		t.Errorf("started variant = %d, want 9", store.startedVariant)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("published events = %d, want 1", len(publisher.events))
	}
	if publisher.events[0].Topic != "game:12" {
		t.Errorf("event topic = %q, want %q", publisher.events[0].Topic, "game:12")
	}
	if publisher.events[0].Type != "game.started" {
		t.Errorf("event type = %q, want %q", publisher.events[0].Type, "game.started")
	}
}

func TestStartGameNoPlayers(t *testing.T) {
	store := &fakeStore{playerCount: 0}
	variants := &fakeVariants{id: 9}
	service := NewService(store, variants, &fakePublisher{})

	_, err := service.StartGame(context.Background(), 12)
	if !errors.Is(err, ErrGameNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrGameNotFound)
	}
	if variants.called {
		t.Fatal("variant lookup was called for an empty game")
	}
	if store.startCalled {
		t.Fatal("StartGame was called for an empty game")
	}
}

func TestStartGameCountError(t *testing.T) {
	want := errors.New("db down")
	store := &fakeStore{countErr: want}
	variants := &fakeVariants{id: 9}
	service := NewService(store, variants, &fakePublisher{})

	_, err := service.StartGame(context.Background(), 12)
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
	if store.startCalled {
		t.Fatal("StartGame was called after a count error")
	}
}

func TestStartGameVariantNotFound(t *testing.T) {
	store := &fakeStore{playerCount: 4}
	variants := &fakeVariants{err: gamevariant.ErrNoMatch}
	service := NewService(store, variants, &fakePublisher{})

	_, err := service.StartGame(context.Background(), 12)
	if !errors.Is(err, gamevariant.ErrNoMatch) {
		t.Fatalf("error = %v, want %v", err, gamevariant.ErrNoMatch)
	}
	if store.startCalled {
		t.Fatal("StartGame was called after a variant error")
	}
}

func TestStartGameStoreError(t *testing.T) {
	want := ErrGameStarted
	store := &fakeStore{playerCount: 4, startErr: want}
	variants := &fakeVariants{id: 9}
	service := NewService(store, variants, &fakePublisher{})

	_, err := service.StartGame(context.Background(), 12)
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}
