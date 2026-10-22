package bean

type Bean struct {
	Id       uint16  `json:"id"`
	Name     string  `json:"name"`
	ImageKey *string `json:"image_key"`
}
