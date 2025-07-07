package action

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/visit/report"
	"github.com/mamacare/services/internal/port/hasura"
	"github.com/mamacare/services/internal/port/response"
	"github.com/mamacare/services/internal/port/validation"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// GenerateVisitReportRequest is the request for generating a visit report
type GenerateVisitReportRequest struct {
	VisitID string `json:"visit_id" validate:"required,uuid"`
}

// GenerateFacilitySummaryRequest is the request for generating a facility summary
type GenerateFacilitySummaryRequest struct {
	FacilityID string `json:"facility_id" validate:"required,uuid"`
	StartDate  string `json:"start_date" validate:"required,rfc3339"`
	EndDate    string `json:"end_date" validate:"required,rfc3339"`
}

// GenerateCHWSummaryRequest is the request for generating a CHW summary
type GenerateCHWSummaryRequest struct {
	CHWID     string `json:"chw_id" validate:"required,uuid"`
	StartDate string `json:"start_date" validate:"required,rfc3339"`
	EndDate   string `json:"end_date" validate:"required,rfc3339"`
}

// GenerateDistrictSummaryRequest is the request for generating a district summary
type GenerateDistrictSummaryRequest struct {
	District  string `json:"district" validate:"required"`
	StartDate string `json:"start_date" validate:"required,rfc3339"`
	EndDate   string `json:"end_date" validate:"required,rfc3339"`
}

// ReportHandler handles visit report requests
type ReportHandler struct {
	hasura.BaseActionHandler
	reportService *report.Service
	validator     *validation.Validator
	log           logger.Logger
}

// NewReportHandler creates a new report handler
func NewReportHandler(
	log logger.Logger,
	reportService *report.Service,
	validator *validation.Validator,
) *ReportHandler {
	return &ReportHandler{
		BaseActionHandler: hasura.BaseActionHandler{},
		reportService:     reportService,
		validator:         validator,
		log:               log,
	}
}

// GenerateVisitReport generates a detailed report for a visit
func (h *ReportHandler) GenerateVisitReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req GenerateVisitReportRequest
	if err := h.ParseRequest(r, &req); err != nil {
		h.log.Error("Failed to parse request", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		h.log.Error("Invalid request", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Parse UUID
	visitID, err := uuid.Parse(req.VisitID)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid visit ID"))
		return
	}

	// Generate report
	visitReport, err := h.reportService.GenerateVisitReport(ctx, visitID)
	if err != nil {
		h.log.Error("Failed to generate visit report", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"visit_id":   req.VisitID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Generated visit report", logger.Fields{
		"request_id": reqID,
		"visit_id":   req.VisitID,
	})

	response.WriteJSONResponse(w, reqID, visitReport)
}

// GenerateFacilitySummary generates a summary report for a facility
func (h *ReportHandler) GenerateFacilitySummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req GenerateFacilitySummaryRequest
	if err := h.ParseRequest(r, &req); err != nil {
		h.log.Error("Failed to parse request", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		h.log.Error("Invalid request", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Parse UUID
	facilityID, err := uuid.Parse(req.FacilityID)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid facility ID"))
		return
	}

	// Parse dates
	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid start date format"))
		return
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid end date format"))
		return
	}

	// Generate summary
	summary, err := h.reportService.GenerateFacilitySummary(ctx, facilityID, startDate, endDate)
	if err != nil {
		h.log.Error("Failed to generate facility summary", logger.Fields{
			"request_id":  reqID,
			"error":       err.Error(),
			"facility_id": req.FacilityID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Generated facility summary", logger.Fields{
		"request_id":  reqID,
		"facility_id": req.FacilityID,
		"start_date":  req.StartDate,
		"end_date":    req.EndDate,
	})

	response.WriteJSONResponse(w, reqID, summary)
}

// GenerateCHWSummary generates a summary report for a CHW
func (h *ReportHandler) GenerateCHWSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req GenerateCHWSummaryRequest
	if err := h.ParseRequest(r, &req); err != nil {
		h.log.Error("Failed to parse request", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		h.log.Error("Invalid request", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Parse UUID
	chwID, err := uuid.Parse(req.CHWID)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid CHW ID"))
		return
	}

	// Parse dates
	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid start date format"))
		return
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid end date format"))
		return
	}

	// Generate summary
	summary, err := h.reportService.GenerateCHWSummary(ctx, chwID, startDate, endDate)
	if err != nil {
		h.log.Error("Failed to generate CHW summary", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"chw_id":     req.CHWID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Generated CHW summary", logger.Fields{
		"request_id": reqID,
		"chw_id":     req.CHWID,
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
	})

	response.WriteJSONResponse(w, reqID, summary)
}

// GenerateDistrictSummary generates a summary report for a district
func (h *ReportHandler) GenerateDistrictSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req GenerateDistrictSummaryRequest
	if err := h.ParseRequest(r, &req); err != nil {
		h.log.Error("Failed to parse request", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		h.log.Error("Invalid request", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Parse dates
	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid start date format"))
		return
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid end date format"))
		return
	}

	// Generate summary
	summary, err := h.reportService.GenerateDistrictSummary(ctx, req.District, startDate, endDate)
	if err != nil {
		h.log.Error("Failed to generate district summary", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"district":   req.District,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Generated district summary", logger.Fields{
		"request_id": reqID,
		"district":   req.District,
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
	})

	response.WriteJSONResponse(w, reqID, summary)
}
