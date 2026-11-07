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
	joinedPlayer   *Player
	joinErr        error
	joinCalled     bool
	joinGameId     uint64
	playerCount    uint16
	countErr       error
	countCalled    bool
	counts         []uint16
	startCalled    bool
	startCalls     int
	startedGameId  uint64
	startedVariant uint16
	startedCount   uint16
	startErrs      []error
	startErr       error
}

func (f *fakeStore) CreateGame(ctx context.Context) (*Game, *Player, error) {
	f.createCalled = true
	if f.createErr != nil {
		return nil, nil, f.createErr
	}
	return f.createdGame, f.createdPlayer, nil
}

func (f *fakeStore) JoinGame(ctx context.Context, gameId uint64) (*Player, error) {
	f.joinCalled = true
	f.joinGameId = gameId
	if f.joinErr != nil {
		return nil, f.joinErr
	}
	return f.joinedPlayer, nil
}

func (f *fakeStore) CountPlayers(ctx context.Context, gameId uint64) (uint16, error) {
	f.countCalled = true
	if len(f.counts) > 0 {
		count := f.counts[0]
		f.counts = f.counts[1:]
		return count, f.countErr
	}
	return f.playerCount, f.countErr
}

func (f *fakeStore) StartGame(ctx context.Context, gameId uint64, variantId uint16, expectedPlayerCount uint16) (*Game, error) {
	f.startCalled = true
	f.startCalls++
	f.startedGameId = gameId
	f.startedVariant = variantId
	f.startedCount = expectedPlayerCount
	if len(f.startErrs) > 0 {
		err := f.startErrs[0]
		f.startErrs = f.startErrs[1:]
		if err != nil {
			return nil, err
		}
	}
	if f.startErr != nil {
		return nil, f.startErr
	}
	return &Game{Id: gameId, VariantId: &variantId}, nil
}

type fakeVariants struct {
	variant     *gamevariant.GameVariant
	err         error
	playerCount uint16
	called      bool
}

func (f *fakeVariants) FindByPlayerCount(ctx context.Context, playerCount uint16) (*gamevariant.GameVariant, error) {
	f.called = true
	f.playerCount = playerCount
	return f.variant, f.err
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

func TestJoinGame(t *testing.T) {
	wantPlayer := &Player{Id: 8, Seat: 2}
	store := &fakeStore{joinedPlayer: wantPlayer}
	publisher := &fakePublisher{}
	service := NewService(store, &fakeVariants{}, publisher)

	player, err := service.JoinGame(context.Background(), 12)
	if err != nil {
		t.Fatalf("JoinGame() error = %v", err)
	}
	if !store.joinCalled {
		t.Fatal("JoinGame was not called")
	}
	if store.joinGameId != 12 {
		t.Errorf("joined game id = %d, want 12", store.joinGameId)
	}
	if player != wantPlayer {
		t.Errorf("player = %v, want %v", player, wantPlayer)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("published events = %d, want 1", len(publisher.events))
	}
	if publisher.events[0].Topic != "game:12" {
		t.Errorf("event topic = %q, want %q", publisher.events[0].Topic, "game:12")
	}
	if publisher.events[0].Type != "game.player_joined" {
		t.Errorf("event type = %q, want %q", publisher.events[0].Type, "game.player_joined")
	}
	if publisher.events[0].Payload != wantPlayer {
		t.Errorf("event payload = %v, want %v", publisher.events[0].Payload, wantPlayer)
	}
}

func TestJoinGameNotFound(t *testing.T) {
	store := &fakeStore{joinErr: ErrGameNotFound}
	publisher := &fakePublisher{}
	service := NewService(store, &fakeVariants{}, publisher)

	_, err := service.JoinGame(context.Background(), 12)
	if !errors.Is(err, ErrGameNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrGameNotFound)
	}
	if len(publisher.events) != 0 {
		t.Fatalf("published events = %d, want 0", len(publisher.events))
	}
}

func TestJoinGameAlreadyStarted(t *testing.T) {
	store := &fakeStore{joinErr: ErrGameStarted}
	publisher := &fakePublisher{}
	service := NewService(store, &fakeVariants{}, publisher)

	_, err := service.JoinGame(context.Background(), 12)
	if !errors.Is(err, ErrGameStarted) {
		t.Fatalf("error = %v, want %v", err, ErrGameStarted)
	}
	if len(publisher.events) != 0 {
		t.Fatalf("published events = %d, want 0", len(publisher.events))
	}
}

func TestStartGame(t *testing.T) {
	store := &fakeStore{playerCount: 4}
	variants := &fakeVariants{variant: &gamevariant.GameVariant{Id: 9}}
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
	variants := &fakeVariants{variant: &gamevariant.GameVariant{Id: 9}}
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
	variants := &fakeVariants{variant: &gamevariant.GameVariant{Id: 9}}
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

func TestStartGameRetriesOnPlayerCountChange(t *testing.T) {
	store := &fakeStore{counts: []uint16{4, 5}, startErrs: []error{ErrPlayersChanged}}
	variants := &fakeVariants{variant: &gamevariant.GameVariant{Id: 9}}
	publisher := &fakePublisher{}
	service := NewService(store, variants, publisher)

	game, err := service.StartGame(context.Background(), 12)
	if err != nil {
		t.Fatalf("StartGame() error = %v", err)
	}
	if store.startCalls != 2 {
		t.Fatalf("StartGame calls = %d, want 2", store.startCalls)
	}
	if store.startedCount != 5 {
		t.Errorf("started count = %d, want 5", store.startedCount)
	}
	if game.VariantId == nil || *game.VariantId != 9 {
		t.Errorf("VariantId = %v, want 9", game.VariantId)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("published events = %d, want 1", len(publisher.events))
	}
}

func TestStartGameGivesUpAfterPlayerCountChanges(t *testing.T) {
	store := &fakeStore{playerCount: 4, startErrs: []error{ErrPlayersChanged, ErrPlayersChanged, ErrPlayersChanged}}
	publisher := &fakePublisher{}
	service := NewService(store, &fakeVariants{variant: &gamevariant.GameVariant{Id: 9}}, publisher)

	_, err := service.StartGame(context.Background(), 12)
	if !errors.Is(err, ErrPlayersChanged) {
		t.Fatalf("error = %v, want %v", err, ErrPlayersChanged)
	}
	if store.startCalls != startGameAttempts {
		t.Errorf("StartGame calls = %d, want %d", store.startCalls, startGameAttempts)
	}
	if len(publisher.events) != 0 {
		t.Fatalf("published events = %d, want 0", len(publisher.events))
	}
}

func TestStartGameStoreError(t *testing.T) {
	want := ErrGameStarted
	store := &fakeStore{playerCount: 4, startErr: want}
	variants := &fakeVariants{variant: &gamevariant.GameVariant{Id: 9}}
	service := NewService(store, variants, &fakePublisher{})

	_, err := service.StartGame(context.Background(), 12)
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}
