package action

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/health/risk"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/domain/repository"
	"github.com/mamacare/services/internal/port/hasura"
	"github.com/mamacare/services/internal/port/response"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// RiskAssessmentRequest is the request for risk assessment
type RiskAssessmentRequest struct {
	MotherID uuid.UUID `json:"mother_id" validate:"required"`
}

// RiskAssessmentResult is the response for risk assessment
type RiskAssessmentResult struct {
	RiskLevel   string           `json:"risk_level"`
	RiskScore   int              `json:"risk_score"`
	RiskFactors json.RawMessage  `json:"risk_factors"`
}

// RiskHandler handles risk assessment actions
type RiskHandler struct {
	hasura.BaseActionHandler
	riskService    *risk.Service
	motherRepo     repository.MotherRepository
	healthMetricRepo repository.HealthMetricRepository
	log           logger.Logger
}

// NewRiskHandler creates a new risk handler
func NewRiskHandler(
	log logger.Logger,
	riskService *risk.Service,
	motherRepo repository.MotherRepository,
	healthMetricRepo repository.HealthMetricRepository,
) *RiskHandler {
	return &RiskHandler{
		BaseActionHandler: hasura.BaseActionHandler{},
		riskService:      riskService,
		motherRepo:       motherRepo,
		healthMetricRepo: healthMetricRepo,
		log:             log,
	}
}

// CalculateRisk calculates risk for a mother
func (h *RiskHandler) CalculateRisk(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req RiskAssessmentRequest
	if err := h.ParseRequest(r, &req); err != nil {
		h.log.Error("Failed to parse request", logger.FieldsMap{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	riskAssessment, err := h.calculateRiskForMother(ctx, req.MotherID)
	if err != nil {
		h.log.Error("Failed to calculate risk", logger.FieldsMap{
			"request_id": reqID,
			"mother_id":  req.MotherID.String(),
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Convert risk factors to JSON
	riskFactorsJSON, err := json.Marshal(riskAssessment.RiskFactors)
	if err != nil {
		h.log.Error("Failed to marshal risk factors", logger.FieldsMap{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.Internal, "Internal server error"))
		return
	}

	result := RiskAssessmentResult{
		RiskLevel:   string(riskAssessment.RiskLevel),
		RiskScore:   riskAssessment.RiskScore,
		RiskFactors: riskFactorsJSON,
	}

	h.log.Info("Risk assessment completed", logger.FieldsMap{
		"request_id": reqID,
		"mother_id":  req.MotherID.String(),
		"risk_level": result.RiskLevel,
	})

	response.WriteJSONResponse(w, reqID, result)
}

// calculateRiskForMother retrieves mother data and calculates risk
func (h *RiskHandler) calculateRiskForMother(ctx context.Context, motherID uuid.UUID) (*risk.RiskAssessment, error) {
	// Retrieve mother information
	mother, err := h.motherRepo.GetByID(ctx, motherID)
	if err != nil {
		return nil, errorx.New(errorx.NotFound, "Mother not found")
	}

	// Retrieve recent health metrics
	metrics, err := h.healthMetricRepo.GetRecentByMotherID(ctx, motherID, 10)
	if err != nil {
		h.log.Warn("Failed to retrieve health metrics", logger.FieldsMap{
			"mother_id": motherID.String(),
			"error":     err.Error(),
		})
		// Continue with empty metrics rather than failing
		metrics = []*model.HealthMetric{}
	}

	// Calculate risk assessment
	assessment, err := h.riskService.CalculateRisk(mother, metrics)
	if err != nil {
		return nil, errorx.New(errorx.Internal, "Failed to calculate risk assessment")
	}

	return assessment, nil
}
