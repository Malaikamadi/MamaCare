package hasura

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mamacare/services/internal/port/middleware"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// ActionRequest represents the standard envelope for Hasura action requests
type ActionRequest struct {
	// ActionName is the name of the action being executed
	ActionName string `json:"action_name"`
	
	// Input contains the action request payload
	Input json.RawMessage `json:"input"`
	
	// SessionVariables contains Hasura session variables
	SessionVariables map[string]string `json:"session_variables"`
	
	// RequestQuery is the GraphQL query that resulted in this action
	RequestQuery string `json:"request_query"`
}

// BaseActionHandler provides common functionality for Hasura action handlers
type BaseActionHandler struct {
	log logger.Logger
}

// NewBaseActionHandler creates a new base action handler
func NewBaseActionHandler(log logger.Logger) *BaseActionHandler {
	return &BaseActionHandler{
		log: log,
	}
}

// ParseRequest parses a Hasura action request
func (h *BaseActionHandler) ParseRequest(r *http.Request, dst interface{}) (*ActionRequest, error) {
	// Get request ID for correlation
	requestID := middleware.GetRequestID(r.Context())
	
	// Read request body
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	
	if err != nil {
		h.log.Error("Failed to read request body", err,
			logger.Field{Key: "request_id", Value: requestID})
		return nil, errorx.Wrap(err, errorx.BadRequest, "failed to read request body")
	}
	
	// Parse the envelope
	var actionReq ActionRequest
	if err := json.Unmarshal(body, &actionReq); err != nil {
		h.log.Error("Failed to parse action request", err,
			logger.Field{Key: "request_id", Value: requestID})
		return nil, errorx.Wrap(err, errorx.BadRequest, "invalid action request format")
	}
	
	// Parse the input payload
	if dst != nil {
		if err := json.Unmarshal(actionReq.Input, dst); err != nil {
			h.log.Error("Failed to parse action input", err,
				logger.Field{Key: "request_id", Value: requestID},
				logger.Field{Key: "action", Value: actionReq.ActionName})
			return nil, errorx.Wrap(err, errorx.BadRequest, "invalid action input format")
		}
	}
	
	h.log.Debug("Parsed Hasura action request",
		logger.Field{Key: "request_id", Value: requestID},
		logger.Field{Key: "action", Value: actionReq.ActionName})
	
	return &actionReq, nil
}

// SendResponse sends a response for a Hasura action
func (h *BaseActionHandler) SendResponse(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		requestID := middleware.GetRequestID(r.Context())
		h.log.Error("Failed to encode response", err,
			logger.Field{Key: "request_id", Value: requestID})
		
		// Send error response
		h.SendError(w, r, errorx.Wrap(err, errorx.InternalServerError, "failed to encode response"))
	}
}

// SendError sends an error response for a Hasura action
func (h *BaseActionHandler) SendError(w http.ResponseWriter, r *http.Request, err error) {
	requestID := middleware.GetRequestID(r.Context())
	h.log.Error("Hasura action error", err,
		logger.Field{Key: "request_id", Value: requestID})
	
	// Define error response structure expected by Hasura
	response := struct {
		Message string                 `json:"message"`
		Code    string                 `json:"code"`
		Details map[string]interface{} `json:"details,omitempty"`
	}{
		Message: "An error occurred",
		Code:    "INTERNAL_ERROR",
	}
	
	// Extract more details from errorx.Error if available
	if appErr, ok := err.(*errorx.Error); ok {
		// Convert context map to interface map
		contextMap := appErr.GetContext()
		detailsMap := make(map[string]interface{})
		for k, v := range contextMap {
			detailsMap[k] = v
		}
		
		response.Message = appErr.Error()
		response.Code = fmt.Sprintf("%d", appErr.GetType())
		response.Details = detailsMap
	} else {
		response.Message = err.Error()
	}
	
	// Set status code based on error type
	statusCode := http.StatusInternalServerError
	if appErr, ok := err.(*errorx.Error); ok {
		statusCode = errorx.HTTPStatusCode(appErr)
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("Failed to encode error response", err,
			logger.Field{Key: "request_id", Value: requestID})
		
		// Fallback to simpler response
		w.Write([]byte(`{"message":"Internal server error","code":"INTERNAL_ERROR"}`))
	}
}
