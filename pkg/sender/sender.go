package sender

import (
	"fmt"
	"sync"

	"gopkg.in/gomail.v2"
)

// Sender represents entity that sends messages using SmtpPort.
// Stores parameters that are necessary for sender:
// SmtpPort server host and port, email address of the sender and its SmtpPort app password.
type Sender struct {
	SmtpHost       string
	SmtpPort       int
	SenderEmail    string
	SenderPassword string
	UseSSl         bool
}

// NewSender Initializes new Sender object with passed values.
func NewSender(smtpHost string, smtpPort int, senderEmail, senderPassword string, useSSL bool) *Sender {
	return &Sender{
		SmtpHost:       smtpHost,
		SmtpPort:       smtpPort,
		SenderEmail:    senderEmail,
		SenderPassword: senderPassword,
		UseSSl:         useSSL,
	}
}

// SendEmail method sends the message [Msg] using SMTP. Expecting execution in parallel goroutine,
// so requires sync.WaitGroup [wg] and chanel of for EmailStatus, where the error will be stored, if occurs.
// Set storeForSender as true, if you want to store the mail in sender's mailbox, false otherwise.
func (s *Sender) SendEmail(msg *gomail.Message, status chan EmailStatus, wg *sync.WaitGroup, storeForSender bool) {
	defer wg.Done()
	emailStatus := EmailStatus{Msg: msg, Status: Success, StatusMsg: ""}

	if storeForSender {
		// Add sender email to recipient, so sender will have the copy of the message.
		// It is needed since SMTP do not save the message for sender.
		existingRecipients := msg.GetHeader("To")
		msg.SetHeader("To", append(existingRecipients, s.SenderEmail)...)
	}

	// initialize SmtpPort Dialer with SSL connection
	dialer := gomail.NewDialer(s.SmtpHost, s.SmtpPort, s.SenderEmail, s.SenderPassword)
	dialer.SSL = s.UseSSl

	err := dialer.DialAndSend(msg)
	if err != nil {
		emailStatus.Status = Error
		emailStatus.StatusMsg = fmt.Sprintf("failed to send message to %s", msg.GetHeader("To"))
		emailStatus.Cause = err
	}
	status <- emailStatus
}
