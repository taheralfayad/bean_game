package bean

type Bean struct {
	Id          uint16  `json:"id"`
	Name        string  `json:"name"`
	ImageFileID *string `json:"image_file_id"`
}
