package service

import (
	"fmt"
	"li-acc/internal/errs"
)

// EmailSendingError error raised when sender.SendMail() returned an error.
// Implement interface errs.CodedError
type EmailSendingError struct {
	MapReceiverCause map[string]string // map receiver's email -> error message (cause)
	AttachmentPaths  map[string]string
}

func (e *EmailSendingError) Error() string {
	var msg []string
	for rec, cause := range e.MapReceiverCause {
		msg = append(msg, fmt.Sprint(rec, ": ", cause))
	}
	return fmt.Sprintf("errors occured sending emails: %v", msg)
}

func (e *EmailSendingError) Kind() errs.Kind {
	return errs.User
}

func (e *EmailSendingError) Unwrap() error {
	return nil
}

func (e *EmailSendingError) FailedCount() int {
	return len(e.MapReceiverCause)
}
