package errs

import "fmt"

// CodedError is interface for all custom errors that have category (Kind)
type CodedError interface {
	error
	Kind() Kind
	Unwrap() error
}

// BaseError is a realization of CodedError interface.
// Wraps internal error `err` and adds a context `msg`
type BaseError struct {
	kind Kind
	msg  string
	err  error
}

// Error realizes error interface
func (e *BaseError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %v", e.msg, e.err)
	}
	return e.msg
}

func (e *BaseError) Unwrap() error {
	return e.err
}

// Kind returns error kind.
func (e *BaseError) Kind() Kind {
	return e.kind
}

// Wrap creates wrapped error with kind.
func Wrap(kind Kind, msg string, err error) *BaseError {
	return &BaseError{
		kind: kind,
		msg:  msg,
		err:  err,
	}
}

// New creates new error without wrapping existing one (no `err` parameter passed).
func New(kind Kind, msg string) *BaseError {
	return &BaseError{
		kind: kind,
		msg:  msg,
	}
}
