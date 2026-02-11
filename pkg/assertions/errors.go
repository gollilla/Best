package assertions

import "fmt"

// AssertionError represents a failed assertion
type AssertionError struct {
	Message  string
	Expected interface{}
	Actual   interface{}
}

// Error implements the error interface
func (e *AssertionError) Error() string {
	return e.Message
}

// NewAssertionError creates a new AssertionError
func NewAssertionError(message string, expected, actual interface{}) *AssertionError {
	return &AssertionError{
		Message:  message,
		Expected: expected,
		Actual:   actual,
	}
}

// Errorf creates a new AssertionError with formatted message
func Errorf(format string, args ...interface{}) *AssertionError {
	return &AssertionError{
		Message: fmt.Sprintf(format, args...),
	}
}
