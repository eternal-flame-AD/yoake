package config

type Communication struct {
	Gotify   CommGotify
	Email    CommEmail
	Telegram CommTelegram
}

type CommTelegram struct {
	Token string
	Owner string
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
