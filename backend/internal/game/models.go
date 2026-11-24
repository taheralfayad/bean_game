package game

import "github.com/taheralfayad/bean_game/backend/internal/deck"

type State string

const (
	StateLobby    State = "lobby"
	StateStarting State = "starting"
	StateStarted  State = "started"
)

type Location string

const (
	LocationDraw     Location = "draw"
	LocationDiscard  Location = "discard"
	LocationHand     Location = "hand"
	LocationField    Location = "field"
	LocationRedeemed Location = "redeemed"
	LocationTrade    Location = "trade"
)

type Game struct {
	Id            uint64  `json:"id"`
	ShuffleNumber uint16  `json:"shuffle_number"`
	CurrentSeat   uint16  `json:"current_seat"`
	CurrentPhase  uint16  `json:"current_phase"`
	VariantId     *uint16 `json:"variant_id"`
	State         State   `json:"state"`
}

type Player struct {
	Id    uint64          `json:"id"`
	Seat  uint16          `json:"seat"`
	Cards []deck.DeckBean `json:"cards"`
}
