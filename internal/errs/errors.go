// errors.go file contains common pre-defined errors or helper functions

package errs

import "fmt"

// WrapIOError is a helper function for IO operations, that wraps given error,
// describing it with specified action and filename
func WrapIOError(action, filename string, err error) error {
	return Wrap(System, fmt.Sprintf("failed to %s '%s'", action, filename), err)
}
