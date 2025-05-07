package response

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mamacare/services/pkg/errorx"
)

// Response is a standardized API response structure
type Response struct {
	Success    bool        `json:"success"`               // Whether the request was successful
	Data       interface{} `json:"data,omitempty"`        // Response payload for successful requests
	Error      *Error      `json:"error,omitempty"`       // Error details for failed requests
	Meta       *Meta       `json:"meta,omitempty"`        // Metadata about the response
	RequestID  string      `json:"request_id,omitempty"`  // Correlation ID for request tracing
	ServerTime time.Time   `json:"server_time"`           // Server timestamp for response
}

// Meta contains additional response metadata
type Meta struct {
	Page       int `json:"page,omitempty"`        // Current page number for paginated responses
	PerPage    int `json:"per_page,omitempty"`    // Items per page for paginated responses
	TotalItems int `json:"total_items,omitempty"` // Total number of items available
	TotalPages int `json:"total_pages,omitempty"` // Total number of pages available
}

// Error contains standardized error information
type Error struct {
	Code    string                 `json:"code"`              // Error code for client handling
	Message string                 `json:"message"`           // User-friendly error message
	Details map[string]interface{} `json:"details,omitempty"` // Additional error details
}

// contextKey is a custom type for context keys
type contextKey string

// Context keys
const (
	requestIDKey contextKey = "request_id"
)

// getRequestID extracts request ID from the context
func getRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// NewSuccessResponse creates a standardized successful response
func NewSuccessResponse(data interface{}, meta *Meta, r *http.Request) *Response {
	requestID := getRequestID(r.Context())

	return &Response{
		Success:    true,
		Data:       data,
		Meta:       meta,
		RequestID:  requestID,
		ServerTime: time.Now().UTC(),
	}
}

// NewErrorResponse creates a standardized error response
func NewErrorResponse(err error, r *http.Request) *Response {
	requestID := getRequestID(r.Context())

	var errorData *Error
	
	// Convert to an errorx type if possible
	if appErr, ok := err.(*errorx.Error); ok {
		// Use the Error struct fields
		errorData = &Error{
			Code:    string(appErr.Code),
			Message: appErr.Message,
			Details: appErr.Details,
		}
	} else {
		// Default error handling
		errorData = &Error{
			Code:    string(errorx.InternalServerError),
			Message: err.Error(),
		}
	}

	return &Response{
		Success:    false,
		Error:      errorData,
		RequestID:  requestID,
		ServerTime: time.Now().UTC(),
	}
}

// WithMeta adds metadata to a response
func (r *Response) WithMeta(meta *Meta) *Response {
	r.Meta = meta
	return r
}

// JSON sends the response as JSON to the client
func JSON(w http.ResponseWriter, statusCode int, response *Response) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(response)
}

// Send sends a successful response to the client
func Send(w http.ResponseWriter, r *http.Request, data interface{}) error {
	response := NewSuccessResponse(data, nil, r)
	return JSON(w, http.StatusOK, response)
}

// SendWithStatus sends a successful response with a custom status code
func SendWithStatus(w http.ResponseWriter, r *http.Request, data interface{}, statusCode int) error {
	response := NewSuccessResponse(data, nil, r)
	return JSON(w, statusCode, response)
}

// SendError sends an error response to the client
func SendError(w http.ResponseWriter, r *http.Request, err error) error {
	response := NewErrorResponse(err, r)
	
	// Determine HTTP status code from error type
	statusCode := http.StatusInternalServerError
	if _, ok := err.(*errorx.Error); ok {
		statusCode = errorx.HTTPStatusCode(err)
	}
	
	return JSON(w, statusCode, response)
}

// SendValidationError sends a validation error response
func SendValidationError(w http.ResponseWriter, r *http.Request, err error) error {
	response := NewErrorResponse(err, r)
	return JSON(w, http.StatusBadRequest, response)
}
