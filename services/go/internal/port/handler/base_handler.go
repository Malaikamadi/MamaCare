package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mamacare/services/internal/port/middleware"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
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

// BaseHandler provides common handler functionality
type BaseHandler struct {
	log logger.Logger
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(log logger.Logger) *BaseHandler {
	return &BaseHandler{
		log: log,
	}
}

// SendResponse sends a standard API response
func (h *BaseHandler) SendResponse(w http.ResponseWriter, r *http.Request, status int, data interface{}) error {
	response := &Response{
		Success:    true,
		Data:       data,
		RequestID:  middleware.GetRequestID(r.Context()),
		ServerTime: time.Now().UTC(),
	}
	
	return h.writeJSON(w, status, response)
}

// SendPaginatedResponse sends a standard paginated API response
func (h *BaseHandler) SendPaginatedResponse(
	w http.ResponseWriter, 
	r *http.Request, 
	status int, 
	data interface{}, 
	page, perPage, totalItems int,
) error {
	// Calculate total pages
	totalPages := 0
	if perPage > 0 {
		totalPages = (totalItems + perPage - 1) / perPage
	}
	
	meta := &Meta{
		Page:       page,
		PerPage:    perPage,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
	
	response := &Response{
		Success:    true,
		Data:       data,
		Meta:       meta,
		RequestID:  middleware.GetRequestID(r.Context()),
		ServerTime: time.Now().UTC(),
	}
	
	return h.writeJSON(w, status, response)
}

// SendError sends a standard error response
func (h *BaseHandler) SendError(w http.ResponseWriter, r *http.Request, err error) error {
	var response *Response
	
	// Determine status code and create response
	statusCode := http.StatusInternalServerError
	
	// Convert to an errorx type if possible
	if appErr, ok := err.(*errorx.Error); ok {
			// Convert context map to interface map
		contextMap := appErr.GetContext()
		detailsMap := make(map[string]interface{})
		for k, v := range contextMap {
			detailsMap[k] = v
		}
		
		errorData := &Error{
			Code:    fmt.Sprintf("%d", appErr.GetType()),
			Message: appErr.Error(),
			Details: detailsMap,
		}
		
		response = &Response{
			Success:    false,
			Error:      errorData,
			RequestID:  middleware.GetRequestID(r.Context()),
			ServerTime: time.Now().UTC(),
		}
		
		statusCode = errorx.HTTPStatusCode(appErr)
	} else {
		// Default error handling
		response = &Response{
			Success: false,
			Error: &Error{
				Code:    string(errorx.InternalServerError),
				Message: err.Error(),
			},
			RequestID:  middleware.GetRequestID(r.Context()),
			ServerTime: time.Now().UTC(),
		}
	}
	
	// Log the error
	h.log.Error(
		"API error response", 
		err,
		logger.Field{Key: "request_id", Value: response.RequestID},
		logger.Field{Key: "status_code", Value: statusCode},
	)
	
	return h.writeJSON(w, statusCode, response)
}

// writeJSON writes a JSON response
func (h *BaseHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) error {
	// Set content type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	// Encode and write response
	return json.NewEncoder(w).Encode(data)
}
