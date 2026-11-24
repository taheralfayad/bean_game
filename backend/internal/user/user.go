package user

type Role uint8

const (
	UserRole Role = iota
	AdminRole
)

type User struct {
	Id       uint16 `json:"id"`
	Username string `json:"username"`
	Token    string `json:"-"`
	Role     Role   `json:"role"`
}
