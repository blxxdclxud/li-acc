package model

import "fmt"

type Mail struct {
	Subject         string            // subject of the mail
	Body            string            // body of the message
	To              []string          // list of emails of the receivers
	From            string            // sender email
	AttachmentPaths map[string]string // receiver email -> attachment file path
}

func (m Mail) GetAttachmentPath(email string) (string, error) {
	p, ok := m.AttachmentPaths[email]
	if !ok {
		return "", fmt.Errorf("no such email in AttachmentPaths")
	}
	return p, nil
}

type SMTP struct {
	Host     string
	Port     int
	Email    string
	Password string
	UseTLS   bool
}

const (
	MailDefaultSubject = "Квитанция об оплате ЛИ7"
	MailDefaultBody    = ""
)
