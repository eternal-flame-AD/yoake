package auth

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleUser    Role = "user"
	RoleTelgram Role = "tg"
)
