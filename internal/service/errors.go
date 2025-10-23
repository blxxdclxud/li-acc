package service

import (
	"fmt"
	"li-acc/internal/errs"
	"strings"
)

// CompositeError contains multiple errors together
type CompositeError struct {
	Errors []error
}

func (c *CompositeError) Error() string {
	var msgs []string
	for _, e := range c.Errors {
		msgs = append(msgs, e.Error())
	}
	return fmt.Sprintf("multiple errors: [%s]", strings.Join(msgs, "; "))
}

func (c *CompositeError) Kind() errs.Kind {
	return errs.User
}

func (c *CompositeError) Unwrap() []error {
	if len(c.Errors) == 0 {
		return nil
	}
	return c.Errors
}

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

// EmailMappingError error raised when there are no emails mapped for some payers
type EmailMappingError struct {
	MapPayerReceipt map[string]string // map payer name -> pdf receipt path
}

func (e *EmailMappingError) Error() string {
	var msg []string
	for payer, path := range e.MapPayerReceipt {
		msg = append(msg, fmt.Sprint(payer, ": ", path))
	}
	return fmt.Sprintf("no emails found for some payers. %v", msg)
}

func (e *EmailMappingError) Kind() errs.Kind {
	return errs.User
}

func (e *EmailMappingError) Unwrap() error {
	return nil
}

func (e *EmailMappingError) FailedCount() int {
	return len(e.MapPayerReceipt)
}
