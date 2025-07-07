package errorx

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// ErrorType is the type of an error
type ErrorType uint

const (
	// NoType error
	NoType ErrorType = iota
	// BadRequest error
	BadRequest
	// NotFound error
	NotFound
	// Forbidden error
	Forbidden
	// Unauthorized error
	Unauthorized
	// InternalServerError error
	InternalServerError
)

// Error represents a custom error with type and message
type Error struct {
	errorType     ErrorType
	originalError error
	context       map[string]string
}

// New creates a new Error
func New(errorType ErrorType, msg string) *Error {
	return &Error{
		errorType:     errorType,
		originalError: errors.New(msg),
		context:       make(map[string]string),
	}
}

// Newf creates a new Error with formatted message
func Newf(errorType ErrorType, msg string, args ...interface{}) *Error {
	return &Error{
		errorType:     errorType,
		originalError: errors.New(fmt.Sprintf(msg, args...)),
		context:       make(map[string]string),
	}
}

// Wrap creates a new wrapped error
func Wrap(err error, errorType ErrorType, msg string) *Error {
	return &Error{
		errorType:     errorType,
		originalError: errors.Wrap(err, msg),
		context:       make(map[string]string),
	}
}

// Wrapf creates a new wrapped error with formatted message
func Wrapf(err error, errorType ErrorType, msg string, args ...interface{}) *Error {
	return &Error{
		errorType:     errorType,
		originalError: errors.Wrapf(err, msg, args...),
		context:       make(map[string]string),
	}
}

// Error returns the message of an error
func (e *Error) Error() string {
	return e.originalError.Error()
}

// WithContext adds context to an error
func (e *Error) WithContext(key, value string) *Error {
	e.context[key] = value
	return e
}

// GetContext returns the context of an error
func (e *Error) GetContext() map[string]string {
	return e.context
}

// GetType returns the error type
func (e *Error) GetType() ErrorType {
	return e.errorType
}

// HTTPStatusCode returns the associated HTTP status code
func (e *Error) HTTPStatusCode() int {
	switch e.errorType {
	case BadRequest:
		return http.StatusBadRequest
	case NotFound:
		return http.StatusNotFound
	case Forbidden:
		return http.StatusForbidden
	case Unauthorized:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

// Cause returns the original error
func (e *Error) Cause() error {
	return errors.Cause(e.originalError)
}

// StackTrace returns the error stack trace
func (e *Error) StackTrace() string {
	return fmt.Sprintf("%+v", e.originalError)
}

// IsType checks if an error is of specific ErrorType
func IsType(err error, errorType ErrorType) bool {
	if customErr, ok := err.(*Error); ok {
		return customErr.GetType() == errorType
	}
	return false
}
