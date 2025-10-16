package sender

import (
	"os"

	"gopkg.in/gomail.v2"
)

// StatusType is the type of the StatusMsg message for EmailStatus
type StatusType int

const Info = StatusType(0)    // Regular info message, the same as success but with additional information
const Error = StatusType(1)   // Error message
const Success = StatusType(2) // Success

// EmailStatus is needed to store information about sent email in chanel, or just in regular execution.
type EmailStatus struct {
	Msg       *gomail.Message // The message attempted to be sent
	Status    StatusType      // The StatusMsg type: success, error or just info (last one usually for logs)
	StatusMsg string          // The enhanced message for the error that occurs while sending.
	Cause     error           // != nil if underlying smtp error
}

// FormMessage forms the email message using gomail.Message instance. Fills in following parameters:
// sender email, recipients' emails, subject, body text and the attachment if there is so.
// Returns *gomail.Message and the boolean:
//
//	false - if attachment path is empty or not found, true - otherwise
func FormMessage(subject, body, attachmentFilePath, senderEmail string, recipientEmails ...string) (*gomail.Message, bool) {
	// Create new message
	message := gomail.NewMessage()

	message.SetHeader("From", senderEmail)
	message.SetHeader("To", recipientEmails...)
	message.SetHeader("Subject", subject)

	message.SetBody("text/plain", body)

	if attachmentFilePath != "" {
		if _, err := os.Stat(attachmentFilePath); err == nil {
			message.Attach(attachmentFilePath)
			return message, true
		}
	}

	return message, false
}
