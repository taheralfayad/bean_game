package gamevariant

type GameVariant struct {
	Id                 uint16 `json:"id"`
	Slug               string `json:"slug"`
	Name               string `json:"name"`
	MinPlayers         uint16 `json:"min_players"`
	MaxPlayers         uint16 `json:"max_players"`
	DeckId             uint16 `json:"deck_id"`
	EndOfTurnDrawCount uint16 `json:"end_of_turn_draw_count"`
}
