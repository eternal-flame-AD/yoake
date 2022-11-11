package email

import (
	"fmt"
	"strings"

	"github.com/eternal-flame-AD/yoake/config"
	"github.com/eternal-flame-AD/yoake/internal/comm/model"
	"github.com/vanng822/go-premailer/premailer"
	"gopkg.in/gomail.v2"
)

type Message struct {
	MIME    string
	Subject string
	Message string

	To string
}

type Handler struct {
	dialer *gomail.Dialer
}

func NewHandler() (*Handler, error) {
	conf := config.Config().Comm.Email
	if conf.SMTP.Host == "" || conf.SMTP.Port == 0 {
		return nil, fmt.Errorf("invalid email configuration")
	}
	dialer := gomail.NewDialer(conf.SMTP.Host, conf.SMTP.Port, conf.SMTP.UserName, conf.SMTP.Password)
	return &Handler{
		dialer: dialer,
	}, nil
}

func (h *Handler) SendGenericMessage(gmsg model.GenericMessage) error {
	msg := Message{
		MIME:    gmsg.MIME,
		Subject: gmsg.Subject,
		Message: gmsg.Body,
	}
	return h.SendEmail(msg)
}

func (h *Handler) SupportedMIME() []string {
	return []string{"text/plain", "text/html"}
}

func (h *Handler) SendEmail(msg Message) error {
	conf := config.Config().Comm.Email
	if !strings.HasPrefix(msg.MIME, "text/html") &&
		!strings.HasPrefix(msg.MIME, "text/plain") {
		return fmt.Errorf("does not know how to send MIME type %s", msg.MIME)
	}
	if msg.MIME == "text/html" {
		prem, err := premailer.NewPremailerFromString(msg.Message, premailer.NewOptions())
		if err != nil {
			return err
		}
		msg.Message, err = prem.Transform()
		if err != nil {
			return err
		}
	}

	email := gomail.NewMessage()
	email.SetHeader("From", conf.SMTP.From)
	if msg.To != "" {
		email.SetHeader("To", msg.To)
	} else {
		email.SetHeader("To", conf.SMTP.To)
	}
	if msg.Subject != "" {
		email.SetHeader("Subject", msg.Subject)
	} else {
		email.SetHeader("Subject", conf.SMTP.DefaultSubject)
	}
	email.SetBody(msg.MIME, msg.Message, gomail.SetPartEncoding("base64"))
	return h.dialer.DialAndSend(email)
}
