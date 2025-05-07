package hasura

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/mamacare/services/internal/port/middleware"
	"github.com/mamacare/services/internal/port/response"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// ActionHandler interface and implementation are defined here
// The ActionRequest type is defined in base_action_handler.go

// ActionHandler is the interface for handling Hasura actions
type ActionHandler interface {
	// Handle processes the action request and returns a response
	Handle(w http.ResponseWriter, r *http.Request)
}

// ParseActionRequest is implemented here, but BaseActionHandler is defined in base_action_handler.go

// ParseActionRequest parses the standard Hasura action request envelope
func (h *BaseActionHandler) ParseActionRequest(r *http.Request, input interface{}) (*ActionRequest, error) {
	// Get request ID for logging (if any)
	reqID := middleware.GetRequestID(r.Context())
	
	// Parse request body
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	
	if err != nil {
		h.log.Error("Failed to read action request body", err, 
			logger.Field{Key: "request_id", Value: reqID})
		return nil, errorx.New(errorx.BadRequest, "failed to read request body")
	}
	
	// Parse the envelope
	var actionReq ActionRequest
	if err := json.Unmarshal(body, &actionReq); err != nil {
		h.log.Error("Failed to parse action request", err, 
			logger.Field{Key: "request_id", Value: reqID})
		return nil, errorx.New(errorx.BadRequest, "invalid action request format")
	}
	
	// Parse the input into the provided struct
	if input != nil {
		if err := json.Unmarshal(actionReq.Input, input); err != nil {
			h.log.Error("Failed to parse action input", err, 
				logger.Field{Key: "request_id", Value: reqID},
				logger.Field{Key: "action", Value: actionReq.ActionName})
			return nil, errorx.New(errorx.BadRequest, "invalid input format")
		}
	}
	
	h.log.Debug("Parsed action request", 
		logger.Field{Key: "request_id", Value: reqID},
		logger.Field{Key: "action", Value: actionReq.ActionName})
	
	return &actionReq, nil
}

// SendActionResponse sends a standardized response for Hasura actions
func (h *BaseActionHandler) SendActionResponse(w http.ResponseWriter, r *http.Request, data interface{}) {
	if err := response.Send(w, r, data); err != nil {
		reqID := middleware.GetRequestID(r.Context())
		h.log.Error("Failed to send action response", err, 
			logger.Field{Key: "request_id", Value: reqID})
	}
}

// SendActionError sends a standardized error response for Hasura actions
func (h *BaseActionHandler) SendActionError(w http.ResponseWriter, r *http.Request, err error) {
	reqID := middleware.GetRequestID(r.Context())
	h.log.Error("Action error", err, logger.Field{Key: "request_id", Value: reqID})
	
	if err := response.SendError(w, r, err); err != nil {
		h.log.Error("Failed to send action error response", err, 
			logger.Field{Key: "request_id", Value: reqID})
	}
}
