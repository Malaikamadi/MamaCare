package errorx

import (
	"fmt"
	"net/http"
)

// ErrorCode represents standardized error codes
type ErrorCode string

// Predefined error codes
const (
	// General errors
	Unknown             ErrorCode = "UNKNOWN_ERROR"
	InternalServerError ErrorCode = "INTERNAL_SERVER_ERROR"
	
	// Authentication & authorization errors
	Unauthorized        ErrorCode = "UNAUTHORIZED"
	Forbidden           ErrorCode = "FORBIDDEN"
	TokenExpired        ErrorCode = "TOKEN_EXPIRED"
	InvalidToken        ErrorCode = "INVALID_TOKEN"
	
	// Resource errors
	NotFound            ErrorCode = "NOT_FOUND"
	AlreadyExists       ErrorCode = "ALREADY_EXISTS"
	ResourceExhausted   ErrorCode = "RESOURCE_EXHAUSTED"
	
	// Input validation errors
	InvalidRequest      ErrorCode = "INVALID_REQUEST"
	ValidationFailed    ErrorCode = "VALIDATION_FAILED"
	MissingParameter    ErrorCode = "MISSING_PARAMETER"
	InvalidParameter    ErrorCode = "INVALID_PARAMETER"
	
	// Database errors
	DatabaseError       ErrorCode = "DATABASE_ERROR"
	TransactionFailed   ErrorCode = "TRANSACTION_FAILED"
	
	// External service errors
	ExternalServiceError ErrorCode = "EXTERNAL_SERVICE_ERROR"
	NetworkError         ErrorCode = "NETWORK_ERROR"
	
	// Business logic errors
	BusinessRuleViolation ErrorCode = "BUSINESS_RULE_VIOLATION"
)

// Error represents an application error
type Error struct {
	// Code is the error code
	Code ErrorCode
	
	// Message is the user-friendly error message
	Message string
	
	// Details contains additional error information
	Details map[string]interface{}
	
	// Cause is the underlying error
	Cause error
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// New creates a new Error
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code ErrorCode, message string) *Error {
	if err == nil {
		return New(code, message)
	}
	
	// Check if we're already wrapping an Error
	if appErr, ok := err.(*Error); ok {
		// Preserve original cause and details
		return &Error{
			Code:    code,
			Message: message,
			Details: appErr.Details,
			Cause:   appErr.Cause,
		}
	}
	
	return &Error{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
		Cause:   err,
	}
}

// AddDetail adds detail information to the error
func (e *Error) AddDetail(field, tag, value string) *Error {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	
	e.Details[field] = map[string]string{
		"tag":   tag,
		"value": value,
	}
	
	return e
}

// HTTPStatusCode maps error codes to HTTP status codes
func HTTPStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	
	appErr, ok := err.(*Error)
	if !ok {
		return http.StatusInternalServerError
	}
	
	switch appErr.Code {
	case Unknown, InternalServerError:
		return http.StatusInternalServerError
	case Unauthorized, TokenExpired, InvalidToken:
		return http.StatusUnauthorized
	case Forbidden:
		return http.StatusForbidden
	case NotFound:
		return http.StatusNotFound
	case AlreadyExists:
		return http.StatusConflict
	case InvalidRequest, ValidationFailed, MissingParameter, InvalidParameter:
		return http.StatusBadRequest
	case ResourceExhausted:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// ErrorResponse represents a standardized error response format
type ErrorResponse struct {
	Success   bool                   `json:"success"`
	Error     *ErrorResponseDetails  `json:"error"`
	RequestID string                 `json:"request_id,omitempty"`
}

// ErrorResponseDetails contains error details
type ErrorResponseDetails struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err error, requestID string) *ErrorResponse {
	if err == nil {
		return &ErrorResponse{
			Success: true,
		}
	}
	
	appErr, ok := err.(*Error)
	if !ok {
		return &ErrorResponse{
			Success: false,
			Error: &ErrorResponseDetails{
				Code:    string(Unknown),
				Message: err.Error(),
			},
			RequestID: requestID,
		}
	}
	
	return &ErrorResponse{
		Success: false,
		Error: &ErrorResponseDetails{
			Code:    string(appErr.Code),
			Message: appErr.Message,
			Details: appErr.Details,
		},
		RequestID: requestID,
	}
}

// String returns the JSON representation of the error response
func (r *ErrorResponse) String() string {
	if r.Error == nil {
		return `{"success":true}`
	}
	
	details := ""
	if len(r.Error.Details) > 0 {
		details = `,details:...`
	}
	
	reqID := ""
	if r.RequestID != "" {
		reqID = fmt.Sprintf(`,request_id:"%s"`, r.RequestID)
	}
	
	return fmt.Sprintf(
		`{"success":false,"error":{"code":"%s","message":"%s"%s}%s}`,
		r.Error.Code,
		r.Error.Message,
		details,
		reqID,
	)
}
