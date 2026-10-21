package types

type Game struct {
	Id            uint64 `json:"id"`
	ShuffleNumber uint16 `json:"shuffle_number"`
	CurrentSeat   uint16 `json:"current_seat"`
	CurrentPhase  uint16 `json:"current_phase"`
	VariantId     uint16 `json:"variant_id"`
}
