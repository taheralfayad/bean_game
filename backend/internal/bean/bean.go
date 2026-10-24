package bean

type Bean struct {
	Id       uint16   `json:"id"`
	Name     string   `json:"name"`
	ImageKey *string  `json:"image_key"`
	Payouts  []Payout `json:"payouts"`
}

type Payout struct {
	NumCards uint16 `json:"num_cards"`
	Coins    uint16 `json:"coins"`
}
