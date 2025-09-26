package sender

import (
	"fmt"
	"os"

	"gopkg.in/gomail.v2"
)

// StatusType is the type of the Status message for Emailstatus
type StatusType int

const InfoType = StatusType(0)    // Regular info message, the same as success but with additional information
const ErrorType = StatusType(1)   // Error message
const SuccessType = StatusType(2) // Success

// EmailStatus is needed to store information about sent email in chanel, or just in regular execution.
type EmailStatus struct {
	Msg        *gomail.Message // The message attempted to be sent
	StatusType StatusType      // The Status type: success, error or just info (last one usually for logs)
	Status     string          // The error message that occurs while sending; nil in the case of success.
}

// FormMessage forms the email message using gomail.Message instance. Fills in following parameters:
// sender email, recipients' emails, subject, body text and the attachment if there is so.
func FormMessage(subject, body, attachmentFilePath, senderEmail string, recipientEmails ...string) *EmailStatus {
	// Create new message
	message := gomail.NewMessage()

	message.SetHeader("From", senderEmail)
	message.SetHeader("To", recipientEmails...)
	message.SetHeader("Subject", subject)

	message.SetBody("text/plain", body)

	var status string
	statusType := SuccessType

	if attachmentFilePath != "" {
		if _, err := os.Stat(attachmentFilePath); err == nil {
			message.Attach(attachmentFilePath)
		} else {
			status = fmt.Sprintf("attachment not found, skipping: %v", err)
			statusType = InfoType
		}
	}

	return &EmailStatus{
		Msg:        message,
		StatusType: statusType,
		Status:     status,
	}
}
