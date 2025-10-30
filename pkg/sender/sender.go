package sender

import (
	"fmt"
	"sync"

	"gopkg.in/gomail.v2"
)

type MailSender interface {
	SendEmail(msg *gomail.Message, status chan EmailStatus, storeForSender bool)
	GetSenderEmail() string
}

// Sender represents entity that sends messages using SmtpPort.
// Stores parameters that are necessary for sender:
// SmtpPort server host and port, email address of the sender and its SmtpPort app password.
type Sender struct {
	SmtpHost       string
	SmtpPort       int
	SenderEmail    string
	SenderPassword string
	UseSSL         bool
	dialer         *gomail.Dialer
	dialerOnce     sync.Once
}

// NewSender Initializes new Sender object with passed values.
func NewSender(smtpHost string, smtpPort int, senderEmail, senderPassword string, useSSL bool) *Sender {
	return &Sender{
		SmtpHost:       smtpHost,
		SmtpPort:       smtpPort,
		SenderEmail:    senderEmail,
		SenderPassword: senderPassword,
		UseSSL:         useSSL,
	}
}

// getDialer returns a reusable dialer (initialized once)
func (s *Sender) getDialer() *gomail.Dialer {
	s.dialerOnce.Do(func() {
		s.dialer = gomail.NewDialer(s.SmtpHost, s.SmtpPort, s.SenderEmail, s.SenderPassword)
		s.dialer.SSL = s.UseSSL
	})
	return s.dialer
}

// SendEmail method sends the message [Msg] using SMTP. Expecting execution in parallel goroutine,
// so requires sync.WaitGroup [wg] and chanel of for EmailStatus, where the error will be stored, if occurs.
// Set storeForSender as true, if you want to store the mail in sender's mailbox, false otherwise.
func (s *Sender) SendEmail(msg *gomail.Message, status chan EmailStatus, storeForSender bool) {
	emailStatus := EmailStatus{Msg: msg, Status: Success, StatusMsg: ""}

	if storeForSender {
		// Add sender email to recipient, so sender will have the copy of the message.
		// It is needed since SMTP do not save the message for sender.
		existingRecipients := msg.GetHeader("To")
		msg.SetHeader("To", append(existingRecipients, s.SenderEmail)...)
	}

	dialer := s.getDialer()

	// Use DialAndSend which reuses connections internally
	err := dialer.DialAndSend(msg)
	if err != nil {
		emailStatus.Status = Error
		emailStatus.StatusMsg = fmt.Sprintf("failed to send message to %s", msg.GetHeader("To"))
		emailStatus.Cause = err
	}
	status <- emailStatus
}

func (s *Sender) GetSenderEmail() string {
	return s.SenderEmail
}
