package action

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/emergency/escalation"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
	"github.com/mamacare/services/internal/port/hasura"
	"github.com/mamacare/services/internal/port/response"
)

// EscalationHandler handles all escalation related actions
type EscalationHandler struct {
	*hasura.BaseActionHandler
	escalationService *escalation.Service
	logger            logger.Logger
}

// CreateEscalationTierRequest defines the request payload for creating an escalation tier
type CreateEscalationTierRequest struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	Level               int    `json:"level"`
	ResponseTimeMinutes int    `json:"response_time_minutes"`
}

// LinkTiersRequest defines the request payload for linking escalation tiers
type LinkTiersRequest struct {
	TierID     string `json:"tier_id"`
	NextTierID string `json:"next_tier_id"`
}

// AddContactToTierRequest defines the request payload for adding a contact to a tier
type AddContactToTierRequest struct {
	TierID    string `json:"tier_id"`
	ContactID string `json:"contact_id"`
}

// CreateContactRequest defines the request payload for creating a contact
type CreateContactRequest struct {
	Name         string  `json:"name"`
	Role         string  `json:"role"`
	Phone        string  `json:"phone"`
	Email        string  `json:"email"`
	IsEmergency  bool    `json:"is_emergency"`
	IsEscalation bool    `json:"is_escalation"`
	FacilityID   *string `json:"facility_id,omitempty"`
}

// CreateEscalationPathRequest defines the request payload for creating an escalation path
type CreateEscalationPathRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	TierIDs     []string `json:"tier_ids"`
	FacilityID  *string  `json:"facility_id,omitempty"`
	DistrictID  *string  `json:"district_id,omitempty"`
}

// GetEscalationPathsRequest defines the request payload for getting escalation paths
type GetEscalationPathsRequest struct {
	FacilityID string `json:"facility_id"`
}

// EscalateSOSEventRequest defines the request payload for escalating an SOS event
type EscalateSOSEventRequest struct {
	SOSID  string `json:"sos_id"`
	PathID string `json:"path_id"`
}

// EscalateToNextTierRequest defines the request payload for escalating to the next tier
type EscalateToNextTierRequest struct {
	SOSID       string `json:"sos_id"`
	CurrentTierID string `json:"current_tier_id"`
}

// SendEscalationReminderRequest defines the request payload for sending an escalation reminder
type SendEscalationReminderRequest struct {
	SOSID    string `json:"sos_id"`
	TierID   string `json:"tier_id"`
	Attempts int    `json:"attempts"`
}

// TierResponse defines the response for an escalation tier
type TierResponse struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	Level          int               `json:"level"`
	ResponseTime   int               `json:"response_time_minutes"`
	Contacts       []ContactResponse `json:"contacts"`
	NextTierID     *string           `json:"next_tier_id,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// ContactResponse defines the response for a contact
type ContactResponse struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Role         string     `json:"role"`
	Phone        string     `json:"phone"`
	Email        string     `json:"email,omitempty"`
	FacilityID   *string    `json:"facility_id,omitempty"`
	IsEmergency  bool       `json:"is_emergency"`
	IsEscalation bool       `json:"is_escalation"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// PathResponse defines the response for an escalation path
type PathResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	TierIDs     []string  `json:"tier_ids"`
	IsActive    bool      `json:"is_active"`
	FacilityID  *string   `json:"facility_id,omitempty"`
	DistrictID  *string   `json:"district_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PathListResponse defines the response for a list of escalation paths
type PathListResponse struct {
	Paths []PathResponse `json:"paths"`
}

// NewEscalationHandler creates a new escalation handler
func NewEscalationHandler(escalationService *escalation.Service, logger logger.Logger) *EscalationHandler {
	return &EscalationHandler{
		BaseActionHandler: hasura.NewBaseActionHandler(logger),
		escalationService: escalationService,
		logger:            logger,
	}
}

// CreateEscalationTier handles creating a new escalation tier
func (h *EscalationHandler) CreateEscalationTier(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &CreateEscalationTierRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse create escalation tier request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	tierReq := req.(*CreateEscalationTierRequest)

	// Convert int level to enum
	var level model.EscalationLevel
	switch tierReq.Level {
	case 1:
		level = model.EscalationLevelLow
	case 2:
		level = model.EscalationLevelMedium
	case 3:
		level = model.EscalationLevelHigh
	case 4:
		level = model.EscalationLevelCritical
	default:
		h.logger.Error(ctx, "Invalid escalation level", logger.FieldsMap{
			"level":      tierReq.Level,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid escalation level, must be 1-4", requestID)
		return
	}

	// Call service
	tier, err := h.escalationService.CreateEscalationTier(
		ctx,
		tierReq.Name,
		tierReq.Description,
		level,
		tierReq.ResponseTimeMinutes,
	)
	if err != nil {
		h.logger.Error(ctx, "Failed to create escalation tier", logger.FieldsMap{
			"error":      err.Error(),
			"name":       tierReq.Name,
			"level":      tierReq.Level,
			"request_id": requestID,
		})

		var httpStatus int
		if errorx.IsOfType(err, errorx.Validation) {
			httpStatus = http.StatusBadRequest
		} else {
			httpStatus = http.StatusInternalServerError
		}

		response.WriteErrorResponse(w, httpStatus, err.Error(), requestID)
		return
	}

	// Prepare response
	resp := tierToResponse(tier)

	h.logger.Info(ctx, "Created escalation tier", logger.FieldsMap{
		"tier_id":    tier.ID.String(),
		"name":       tier.Name,
		"level":      int(tier.Level),
		"request_id": requestID,
	})

	response.WriteJSONResponse(w, http.StatusCreated, resp, requestID)
}

// Helper method to convert tier model to response
func tierToResponse(tier *model.EscalationTier) TierResponse {
	contacts := make([]ContactResponse, 0, len(tier.Contacts))
	for _, contact := range tier.Contacts {
		contacts = append(contacts, contactToResponse(contact))
	}

	var nextTierID *string
	if tier.NextTierID != nil {
		nextID := tier.NextTierID.String()
		nextTierID = &nextID
	}

	return TierResponse{
		ID:             tier.ID.String(),
		Name:           tier.Name,
		Description:    tier.Description,
		Level:          int(tier.Level),
		ResponseTime:   tier.ResponseTimeMinutes,
		Contacts:       contacts,
		NextTierID:     nextTierID,
		CreatedAt:      tier.CreatedAt,
		UpdatedAt:      tier.UpdatedAt,
	}
}

// Helper method to convert contact model to response
func contactToResponse(contact *model.Contact) ContactResponse {
	var facilityID *string
	if contact.FacilityID != nil {
		facID := contact.FacilityID.String()
		facilityID = &facID
	}

	return ContactResponse{
		ID:           contact.ID.String(),
		Name:         contact.Name,
		Role:         contact.Role,
		Phone:        contact.Phone,
		Email:        contact.Email,
		FacilityID:   facilityID,
		IsEmergency:  contact.IsEmergency,
		IsEscalation: contact.IsEscalation,
		CreatedAt:    contact.CreatedAt,
		UpdatedAt:    contact.UpdatedAt,
	}
}

// Helper method to convert path model to response
func pathToResponse(path *model.EscalationPath) PathResponse {
	tierIDs := make([]string, 0, len(path.TierIDs))
	for _, tierID := range path.TierIDs {
		tierIDs = append(tierIDs, tierID.String())
	}

	var facilityID *string
	if path.FacilityID != nil {
		facID := path.FacilityID.String()
		facilityID = &facID
	}

	var districtID *string
	if path.DistrictID != nil {
		distID := path.DistrictID.String()
		districtID = &distID
	}

	return PathResponse{
		ID:          path.ID.String(),
		Name:        path.Name,
		Description: path.Description,
		TierIDs:     tierIDs,
		IsActive:    path.IsActive,
		FacilityID:  facilityID,
		DistrictID:  districtID,
		CreatedAt:   path.CreatedAt,
		UpdatedAt:   path.UpdatedAt,
	}
}

// LinkTiers handles linking two escalation tiers
func (h *EscalationHandler) LinkTiers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &LinkTiersRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse link tiers request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	linkReq := req.(*LinkTiersRequest)

	// Convert string IDs to UUIDs
	tierID, err := uuid.Parse(linkReq.TierID)
	if err != nil {
		h.logger.Error(ctx, "Invalid tier ID", logger.FieldsMap{
			"error":      err.Error(),
			"tier_id":    linkReq.TierID,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid tier ID", requestID)
		return
	}

	nextTierID, err := uuid.Parse(linkReq.NextTierID)
	if err != nil {
		h.logger.Error(ctx, "Invalid next tier ID", logger.FieldsMap{
			"error":        err.Error(),
			"next_tier_id": linkReq.NextTierID,
			"request_id":   requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid next tier ID", requestID)
		return
	}

	// Call service
	tier, err := h.escalationService.LinkTiers(ctx, tierID, nextTierID)
	if err != nil {
		h.logger.Error(ctx, "Failed to link tiers", logger.FieldsMap{
			"error":        err.Error(),
			"tier_id":      linkReq.TierID,
			"next_tier_id": linkReq.NextTierID,
			"request_id":   requestID,
		})

		var httpStatus int
		if errorx.IsOfType(err, errorx.NotFound) {
			httpStatus = http.StatusNotFound
		} else {
			httpStatus = http.StatusInternalServerError
		}

		response.WriteErrorResponse(w, httpStatus, err.Error(), requestID)
		return
	}

	// Prepare response
	resp := tierToResponse(tier)

	h.logger.Info(ctx, "Linked escalation tiers", logger.FieldsMap{
		"tier_id":      tier.ID.String(),
		"next_tier_id": nextTierID.String(),
		"request_id":   requestID,
	})

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}

// CreateContact handles creating a new contact
func (h *EscalationHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &CreateContactRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse create contact request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	contactReq := req.(*CreateContactRequest)

	// Convert string ID to UUID if provided
	var facilityID *uuid.UUID
	if contactReq.FacilityID != nil {
		facID, err := uuid.Parse(*contactReq.FacilityID)
		if err != nil {
			h.logger.Error(ctx, "Invalid facility ID", logger.FieldsMap{
				"error":       err.Error(),
				"facility_id": *contactReq.FacilityID,
				"request_id":  requestID,
			})
			response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid facility ID", requestID)
			return
		}
		facilityID = &facID
	}

	// Call service
	contact, err := h.escalationService.CreateContact(
		ctx,
		contactReq.Name,
		contactReq.Role,
		contactReq.Phone,
		contactReq.Email,
		contactReq.IsEmergency,
		contactReq.IsEscalation,
		facilityID,
	)
	if err != nil {
		h.logger.Error(ctx, "Failed to create contact", logger.FieldsMap{
			"error":      err.Error(),
			"name":       contactReq.Name,
			"request_id": requestID,
		})

		var httpStatus int
		if errorx.IsOfType(err, errorx.Validation) {
			httpStatus = http.StatusBadRequest
		} else {
			httpStatus = http.StatusInternalServerError
		}

		response.WriteErrorResponse(w, httpStatus, err.Error(), requestID)
		return
	}

	// Prepare response
	resp := contactToResponse(contact)

	h.logger.Info(ctx, "Created contact", logger.FieldsMap{
		"contact_id": contact.ID.String(),
		"name":       contact.Name,
		"request_id": requestID,
	})

	response.WriteJSONResponse(w, http.StatusCreated, resp, requestID)
}

// AddContactToTier handles adding a contact to an escalation tier
func (h *EscalationHandler) AddContactToTier(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &AddContactToTierRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse add contact to tier request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	addReq := req.(*AddContactToTierRequest)

	// Convert string IDs to UUIDs
	tierID, err := uuid.Parse(addReq.TierID)
	if err != nil {
		h.logger.Error(ctx, "Invalid tier ID", logger.FieldsMap{
			"error":      err.Error(),
			"tier_id":    addReq.TierID,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid tier ID", requestID)
		return
	}

	contactID, err := uuid.Parse(addReq.ContactID)
	if err != nil {
		h.logger.Error(ctx, "Invalid contact ID", logger.FieldsMap{
			"error":      err.Error(),
			"contact_id": addReq.ContactID,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid contact ID", requestID)
		return
	}

	// Call service
	tier, err := h.escalationService.AddContactToTier(ctx, tierID, contactID)
	if err != nil {
		h.logger.Error(ctx, "Failed to add contact to tier", logger.FieldsMap{
			"error":      err.Error(),
			"tier_id":    addReq.TierID,
			"contact_id": addReq.ContactID,
			"request_id": requestID,
		})

		var httpStatus int
		if errorx.IsOfType(err, errorx.NotFound) {
			httpStatus = http.StatusNotFound
		} else {
			httpStatus = http.StatusInternalServerError
		}

		response.WriteErrorResponse(w, httpStatus, err.Error(), requestID)
		return
	}

	// Prepare response
	resp := tierToResponse(tier)

	h.logger.Info(ctx, "Added contact to tier", logger.FieldsMap{
		"tier_id":    tierID.String(),
		"contact_id": contactID.String(),
		"request_id": requestID,
	})

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}

// CreateEscalationPath handles creating a new escalation path
func (h *EscalationHandler) CreateEscalationPath(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &CreateEscalationPathRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse create escalation path request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	pathReq := req.(*CreateEscalationPathRequest)

	// Convert string IDs to UUIDs
	tierIDs := make([]uuid.UUID, 0, len(pathReq.TierIDs))
	for _, tierIDStr := range pathReq.TierIDs {
		tierID, err := uuid.Parse(tierIDStr)
		if err != nil {
			h.logger.Error(ctx, "Invalid tier ID in path request", logger.FieldsMap{
				"error":      err.Error(),
				"tier_id":    tierIDStr,
				"request_id": requestID,
			})
			response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid tier ID: "+tierIDStr, requestID)
			return
		}
		tierIDs = append(tierIDs, tierID)
	}

	// Handle optional facility ID
	var facilityID *uuid.UUID
	if pathReq.FacilityID != nil {
		facID, err := uuid.Parse(*pathReq.FacilityID)
		if err != nil {
			h.logger.Error(ctx, "Invalid facility ID", logger.FieldsMap{
				"error":       err.Error(),
				"facility_id": *pathReq.FacilityID,
				"request_id":  requestID,
			})
			response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid facility ID", requestID)
			return
		}
		facilityID = &facID
	}

	// Handle optional district ID
	var districtID *uuid.UUID
	if pathReq.DistrictID != nil {
		distID, err := uuid.Parse(*pathReq.DistrictID)
		if err != nil {
			h.logger.Error(ctx, "Invalid district ID", logger.FieldsMap{
				"error":       err.Error(),
				"district_id": *pathReq.DistrictID,
				"request_id":  requestID,
			})
			response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid district ID", requestID)
			return
		}
		districtID = &distID
	}

	// Call service
	path, err := h.escalationService.CreateEscalationPath(
		ctx,
		pathReq.Name,
		pathReq.Description,
		tierIDs,
		facilityID,
		districtID,
	)
	if err != nil {
		h.logger.Error(ctx, "Failed to create escalation path", logger.FieldsMap{
			"error":      err.Error(),
			"name":       pathReq.Name,
			"tier_count": len(tierIDs),
			"request_id": requestID,
		})

		var httpStatus int
		if errorx.IsOfType(err, errorx.Validation) {
			httpStatus = http.StatusBadRequest
		} else if errorx.IsOfType(err, errorx.NotFound) {
			httpStatus = http.StatusNotFound
		} else {
			httpStatus = http.StatusInternalServerError
		}

		response.WriteErrorResponse(w, httpStatus, err.Error(), requestID)
		return
	}

	// Prepare response
	resp := pathToResponse(path)

	h.logger.Info(ctx, "Created escalation path", logger.FieldsMap{
		"path_id":     path.ID.String(),
		"name":        path.Name,
		"tier_count":  len(path.TierIDs),
		"request_id":  requestID,
	})

	response.WriteJSONResponse(w, http.StatusCreated, resp, requestID)
}

// GetEscalationPaths handles retrieving escalation paths for a facility
func (h *EscalationHandler) GetEscalationPaths(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &GetEscalationPathsRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse get escalation paths request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	pathsReq := req.(*GetEscalationPathsRequest)

	// Convert facility ID to UUID
	facilityID, err := uuid.Parse(pathsReq.FacilityID)
	if err != nil {
		h.logger.Error(ctx, "Invalid facility ID", logger.FieldsMap{
			"error":       err.Error(),
			"facility_id": pathsReq.FacilityID,
			"request_id":  requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid facility ID", requestID)
		return
	}

	// Call service
	paths, err := h.escalationService.GetEscalationPathsByFacility(ctx, facilityID)
	if err != nil {
		h.logger.Error(ctx, "Failed to get escalation paths", logger.FieldsMap{
			"error":       err.Error(),
			"facility_id": pathsReq.FacilityID,
			"request_id":  requestID,
		})

		var httpStatus int
		if errorx.IsOfType(err, errorx.NotFound) {
			httpStatus = http.StatusNotFound
		} else {
			httpStatus = http.StatusInternalServerError
		}

		response.WriteErrorResponse(w, httpStatus, err.Error(), requestID)
		return
	}

	// Prepare response
	pathResponses := make([]PathResponse, 0, len(paths))
	for _, path := range paths {
		pathResponses = append(pathResponses, pathToResponse(path))
	}

	resp := PathListResponse{
		Paths: pathResponses,
	}

	h.logger.Info(ctx, "Retrieved escalation paths", logger.FieldsMap{
		"facility_id":  facilityID.String(),
		"path_count":   len(paths),
		"request_id":   requestID,
	})

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}

// EscalateSOSEvent handles starting an escalation process for an SOS event
func (h *EscalationHandler) EscalateSOSEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &EscalateSOSEventRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse escalate SOS event request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	escReq := req.(*EscalateSOSEventRequest)

	// Convert IDs to UUIDs
	sosID, err := uuid.Parse(escReq.SOSID)
	if err != nil {
		h.logger.Error(ctx, "Invalid SOS ID", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     escReq.SOSID,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid SOS ID", requestID)
		return
	}

	pathID, err := uuid.Parse(escReq.PathID)
	if err != nil {
		h.logger.Error(ctx, "Invalid path ID", logger.FieldsMap{
			"error":      err.Error(),
			"path_id":    escReq.PathID,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid path ID", requestID)
		return
	}

	// Call service
	activeTier, err := h.escalationService.EscalateSOSEvent(ctx, sosID, pathID)
	if err != nil {
		h.logger.Error(ctx, "Failed to escalate SOS event", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     escReq.SOSID,
			"path_id":    escReq.PathID,
			"request_id": requestID,
		})

		var httpStatus int
		if errorx.IsOfType(err, errorx.NotFound) {
			httpStatus = http.StatusNotFound
		} else if errorx.IsOfType(err, errorx.Validation) {
			httpStatus = http.StatusBadRequest
		} else {
			httpStatus = http.StatusInternalServerError
		}

		response.WriteErrorResponse(w, httpStatus, err.Error(), requestID)
		return
	}

	// Prepare response
	resp := tierToResponse(activeTier)

	h.logger.Info(ctx, "Started escalation for SOS event", logger.FieldsMap{
		"sos_id":      sosID.String(),
		"path_id":     pathID.String(),
		"active_tier": activeTier.ID.String(),
		"request_id":  requestID,
	})

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}

// EscalateToNextTier handles escalating to the next tier in the path
func (h *EscalationHandler) EscalateToNextTier(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &EscalateToNextTierRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse escalate to next tier request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	escReq := req.(*EscalateToNextTierRequest)

	// Convert IDs to UUIDs
	sosID, err := uuid.Parse(escReq.SOSID)
	if err != nil {
		h.logger.Error(ctx, "Invalid SOS ID", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     escReq.SOSID,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid SOS ID", requestID)
		return
	}

	currentTierID, err := uuid.Parse(escReq.CurrentTierID)
	if err != nil {
		h.logger.Error(ctx, "Invalid current tier ID", logger.FieldsMap{
			"error":           err.Error(),
			"current_tier_id": escReq.CurrentTierID,
			"request_id":      requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid current tier ID", requestID)
		return
	}

	// Call service
	nextTier, err := h.escalationService.EscalateToNextTier(ctx, sosID, currentTierID)
	if err != nil {
		h.logger.Error(ctx, "Failed to escalate to next tier", logger.FieldsMap{
			"error":           err.Error(),
			"sos_id":          escReq.SOSID,
			"current_tier_id": escReq.CurrentTierID,
			"request_id":      requestID,
		})

		var httpStatus int
		if errorx.IsOfType(err, errorx.NotFound) {
			httpStatus = http.StatusNotFound
		} else {
			httpStatus = http.StatusInternalServerError
		}

		response.WriteErrorResponse(w, httpStatus, err.Error(), requestID)
		return
	}

	// Prepare response
	resp := tierToResponse(nextTier)

	h.logger.Info(ctx, "Escalated to next tier", logger.FieldsMap{
		"sos_id":          sosID.String(),
		"current_tier_id": currentTierID.String(),
		"next_tier_id":    nextTier.ID.String(),
		"request_id":      requestID,
	})

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}

// SendEscalationReminder handles sending a reminder to contacts in a tier
func (h *EscalationHandler) SendEscalationReminder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &SendEscalationReminderRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse send escalation reminder request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	remReq := req.(*SendEscalationReminderRequest)

	// Convert IDs to UUIDs
	sosID, err := uuid.Parse(remReq.SOSID)
	if err != nil {
		h.logger.Error(ctx, "Invalid SOS ID", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     remReq.SOSID,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid SOS ID", requestID)
		return
	}

	tierID, err := uuid.Parse(remReq.TierID)
	if err != nil {
		h.logger.Error(ctx, "Invalid tier ID", logger.FieldsMap{
			"error":      err.Error(),
			"tier_id":    remReq.TierID,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid tier ID", requestID)
		return
	}

	// Call service
	sentCount, err := h.escalationService.SendEscalationReminder(
		ctx,
		sosID,
		tierID,
		remReq.Attempts,
	)
	if err != nil {
		h.logger.Error(ctx, "Failed to send escalation reminder", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     remReq.SOSID,
			"tier_id":    remReq.TierID,
			"attempts":   remReq.Attempts,
			"request_id": requestID,
		})

		var httpStatus int
		if errorx.IsOfType(err, errorx.NotFound) {
			httpStatus = http.StatusNotFound
		} else {
			httpStatus = http.StatusInternalServerError
		}

		response.WriteErrorResponse(w, httpStatus, err.Error(), requestID)
		return
	}

	// Prepare response
	resp := struct {
		SentCount int `json:"sent_count"`
		SosID     string `json:"sos_id"`
		TierID    string `json:"tier_id"`
	}{
		SentCount: sentCount,
		SosID:     sosID.String(),
		TierID:    tierID.String(),
	}

	h.logger.Info(ctx, "Sent escalation reminders", logger.FieldsMap{
		"sos_id":      sosID.String(),
		"tier_id":     tierID.String(),
		"sent_count":  sentCount,
		"attempts":    remReq.Attempts,
		"request_id":  requestID,
	})

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}
