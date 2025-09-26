package sender

import (
	"fmt"
	"sync"

	"gopkg.in/gomail.v2"
)

// Sender represents entity that sends messages using SMTP.
// Stores parameters that are necessary for sender:
// SMTP server host and port, email address of the sender and its SMTP app password.
type Sender struct {
	SmtpHost       string
	SmtpPort       int
	SenderEmail    string
	SenderPassword string
}

// NewSender Initializes new Sender object with passed values.
func NewSender(smtpHost string, smtpPort int, senderEmail, senderPassword string) *Sender {
	return &Sender{
		SmtpHost:       smtpHost,
		SmtpPort:       smtpPort,
		SenderEmail:    senderEmail,
		SenderPassword: senderPassword,
	}
}

// SendEmail method sends the message [Msg] using SMTP. Expecting execution in parallel goroutine,
// so requires sync.WaitGroup [wg] and chanel of for EmailStatus, where the error will be stored, if occurs.
func (s *Sender) SendEmail(msg *gomail.Message, status chan EmailStatus, wg *sync.WaitGroup) {
	defer wg.Done()
	emailStatus := EmailStatus{Msg: msg, StatusType: SuccessType, Status: ""}

	// Add sender email to recipient, so sender will have the copy of the message.
	// It is needed since SMTP do not save the message for sender.
	existingRecipients := msg.GetHeader("To")
	msg.SetHeader("To", append(existingRecipients, s.SenderEmail)...)

	// initialize SMTP Dialer with SSL connection
	dialer := gomail.NewDialer(s.SmtpHost, s.SmtpPort, s.SenderEmail, s.SenderPassword)
	dialer.SSL = true

	err := dialer.DialAndSend(msg)
	if err != nil {
		emailStatus.StatusType = ErrorType
		emailStatus.Status = fmt.Sprintf("failed to send message to %s: %v", msg.GetHeader("To"), err)
	}
	status <- emailStatus
}
