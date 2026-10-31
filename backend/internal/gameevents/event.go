package gameevents

type Event struct {
	Topic   string
	Type    string
	ID      string
	Payload any
	Retry   int
}
