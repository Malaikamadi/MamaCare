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

// EventTriggerPayload represents the standard Hasura event trigger payload
type EventTriggerPayload struct {
	// ID is the unique event ID
	ID string `json:"id"`
	
	// Event contains the event data
	Event Event `json:"event"`
	
	// CreatedAt is the timestamp when the event was triggered
	CreatedAt string `json:"created_at"`
	
	// Table contains information about the table that triggered the event
	Table Table `json:"table"`
	
	// Trigger contains information about the trigger that fired
	Trigger Trigger `json:"trigger"`
	
	// DeliveryInfo contains delivery attempt information
	DeliveryInfo DeliveryInfo `json:"delivery_info"`
}

// Event contains the data for the event
type Event struct {
	// SessionVariables contains Hasura session variables
	SessionVariables map[string]string `json:"session_variables"`
	
	// Op is the operation type (INSERT, UPDATE, DELETE)
	Op string `json:"op"`
	
	// Data contains the record data
	Data EventData `json:"data"`
}

// EventData contains the old and new data for the record
type EventData struct {
	// Old contains the old state of the record (for UPDATE/DELETE)
	Old json.RawMessage `json:"old"`
	
	// New contains the new state of the record (for INSERT/UPDATE)
	New json.RawMessage `json:"new"`
}

// Table contains information about the table
type Table struct {
	// Schema is the database schema name
	Schema string `json:"schema"`
	
	// Name is the table name
	Name string `json:"name"`
}

// Trigger contains information about the trigger
type Trigger struct {
	// Name is the trigger name
	Name string `json:"name"`
	
	// ID is the trigger ID
	ID string `json:"id"`
}

// DeliveryInfo contains information about the delivery
type DeliveryInfo struct {
	// CurrentRetry is the current retry count
	CurrentRetry int `json:"current_retry"`
	
	// MaxRetries is the maximum number of retries
	MaxRetries int `json:"max_retries"`
}

// EventHandler is the interface for handling Hasura event triggers
type EventHandler interface {
	// Handle processes the event trigger request
	Handle(w http.ResponseWriter, r *http.Request)
}

// BaseEventHandler provides common functionality for event handlers
type BaseEventHandler struct {
	log logger.Logger
}

// NewBaseEventHandler creates a new base event handler
func NewBaseEventHandler(log logger.Logger) *BaseEventHandler {
	return &BaseEventHandler{
		log: log,
	}
}

// ParseEventPayload parses the standard Hasura event trigger payload
func (h *BaseEventHandler) ParseEventPayload(r *http.Request) (*EventTriggerPayload, error) {
	// Get request ID for logging
	reqID := middleware.GetRequestID(r.Context())
	
	// Parse request body
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	
	if err != nil {
		h.log.Error("Failed to read event trigger body", err, 
			logger.Field{Key: "request_id", Value: reqID})
		// Use the errorx.New() function with the correct BadRequest ErrorType
		return nil, errorx.New(errorx.BadRequest, "failed to read request body")
	}
	
	// Parse the payload
	var payload EventTriggerPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		h.log.Error("Failed to parse event trigger payload", err, 
			logger.Field{Key: "request_id", Value: reqID})
		// Use the errorx.New() function with the correct BadRequest ErrorType
		return nil, errorx.New(errorx.BadRequest, "invalid event payload format")
	}
	
	h.log.Debug("Parsed event trigger payload", 
		logger.Field{Key: "request_id", Value: reqID},
		logger.Field{Key: "event_id", Value: payload.ID},
		logger.Field{Key: "table", Value: payload.Table.Name},
		logger.Field{Key: "operation", Value: payload.Event.Op})
	
	return &payload, nil
}

// ParseNewData parses the new data from an event trigger
func (h *BaseEventHandler) ParseNewData(payload *EventTriggerPayload, data interface{}) error {
	if payload.Event.Data.New == nil {
		// Use the errorx.New() function with the correct BadRequest ErrorType
		return errorx.New(errorx.BadRequest, "event has no new data")
	}
	
	if err := json.Unmarshal(payload.Event.Data.New, data); err != nil {
		return errorx.New(errorx.BadRequest, "failed to parse new event data")
	}
	
	return nil
}

// ParseOldData parses the old data from an event trigger
func (h *BaseEventHandler) ParseOldData(payload *EventTriggerPayload, data interface{}) error {
	if payload.Event.Data.Old == nil {
		// Use the errorx.New() function with the correct BadRequest ErrorType
		return errorx.New(errorx.BadRequest, "event has no old data")
	}
	
	if err := json.Unmarshal(payload.Event.Data.Old, data); err != nil {
		return errorx.New(errorx.BadRequest, "failed to parse old event data")
	}
	
	return nil
}

// SendEventSuccess sends a success response for event triggers
func (h *BaseEventHandler) SendEventSuccess(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Hasura expects an empty 200 response for successful event processing
	w.Write([]byte("{}"))
}

// SendEventError sends an error response for event triggers
func (h *BaseEventHandler) SendEventError(w http.ResponseWriter, r *http.Request, err error) {
	reqID := middleware.GetRequestID(r.Context())
	h.log.Error("Event processing error", err, logger.Field{Key: "request_id", Value: reqID})
	
	if err := response.SendError(w, r, err); err != nil {
		h.log.Error("Failed to send event error response", err, 
			logger.Field{Key: "request_id", Value: reqID})
	}
}
