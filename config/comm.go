package config

type Communication struct {
	Gotify CommGotify
	Email  CommEmail
}

type CommEmail struct {
	SMTP struct {
		Host string
		Port int

		From string
		To   string

		UserName string
		Password string

		DefaultSubject string
	}
}

type CommGotify struct {
	BaseURL  string
	AppToken string
}
