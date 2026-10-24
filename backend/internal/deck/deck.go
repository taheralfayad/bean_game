package deck

type Deck struct {
	Id    uint16     `json:"id"`
	Slug  string     `json:"slug"`
	Beans []DeckBean `json:"beans"`
}

type DeckBean struct {
	BeanId    uint16 `json:"bean_id"`
	CardCount uint16 `json:"card_count"`
}
