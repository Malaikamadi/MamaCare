package action

import (
	"net/http"
	"time"

	"github.com/mamacare/services/internal/app/health/calculator"
	"github.com/mamacare/services/internal/port/hasura"
	"github.com/mamacare/services/internal/port/response"
	"github.com/mamacare/services/internal/port/validation"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// PregnancyCalcRequest is the request for pregnancy date calculations
type PregnancyCalcRequest struct {
	LastMenstrualPeriod string `json:"last_menstrual_period" validate:"required"`
}

// PregnancyCalcResult is the response for pregnancy calculations
type PregnancyCalcResult struct {
	LastMenstrualPeriod   string `json:"last_menstrual_period"`
	ConceptionDate        string `json:"conception_date"`
	ExpectedDeliveryDate  string `json:"expected_delivery_date"`
	GestationalAge        int    `json:"gestational_age"`
	GestationalAgeWeeks   int    `json:"gestational_age_weeks"`
	GestationalAgeDays    int    `json:"gestational_age_days"`
	CurrentTrimester      int    `json:"current_trimester"`
	WeeksRemaining        int    `json:"weeks_remaining"`
	IsPreTerm             bool   `json:"is_pre_term"`
	IsFullTerm            bool   `json:"is_full_term"`
	DaysUntilFullTerm     int    `json:"days_until_full_term"`
	PercentageComplete    float64 `json:"percentage_complete"`
}

// PregnancyHandler handles pregnancy calculation actions
type PregnancyHandler struct {
	hasura.BaseActionHandler
	calcService *calculator.Service
	validator   *validation.Validator
	log         logger.Logger
}

// NewPregnancyHandler creates a new pregnancy calculation handler
func NewPregnancyHandler(
	log logger.Logger,
	calcService *calculator.Service,
	validator *validation.Validator,
) *PregnancyHandler {
	return &PregnancyHandler{
		BaseActionHandler: hasura.BaseActionHandler{},
		calcService:      calcService,
		validator:        validator,
		log:             log,
	}
}

// CalculatePregnancyDates calculates pregnancy-related dates and durations
func (h *PregnancyHandler) CalculatePregnancyDates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req PregnancyCalcRequest
	if err := h.ParseRequest(r, &req); err != nil {
		h.log.Error("Failed to parse request", logger.FieldsMap{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		h.log.Error("Invalid request", logger.FieldsMap{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Parse LMP date
	lmp, err := time.Parse("2006-01-02", req.LastMenstrualPeriod)
	if err != nil {
		h.log.Error("Invalid date format", logger.FieldsMap{
			"request_id": reqID,
			"lmp":        req.LastMenstrualPeriod,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid date format. Use YYYY-MM-DD"))
		return
	}

	// Calculate pregnancy dates
	dateInfo, err := h.calcService.CalculatePregnancyDates(lmp)
	if err != nil {
		h.log.Error("Failed to calculate pregnancy dates", logger.FieldsMap{
			"request_id": reqID,
			"lmp":        req.LastMenstrualPeriod,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Format response
	result := PregnancyCalcResult{
		LastMenstrualPeriod:   dateInfo.LMP.Format("2006-01-02"),
		ConceptionDate:        dateInfo.ConceptionDate.Format("2006-01-02"),
		ExpectedDeliveryDate:  dateInfo.ExpectedDeliveryDate.Format("2006-01-02"),
		GestationalAge:        dateInfo.GestationalAge,
		GestationalAgeWeeks:   dateInfo.GestationalAgeWeeks,
		GestationalAgeDays:    dateInfo.GestationalAgeDays,
		CurrentTrimester:      dateInfo.CurrentTrimester,
		WeeksRemaining:        dateInfo.WeeksRemaining,
		IsPreTerm:             dateInfo.IsPreTerm,
		IsFullTerm:            dateInfo.IsFullTerm,
		DaysUntilFullTerm:     dateInfo.DaysUntilFullTerm,
		PercentageComplete:    dateInfo.PercentageComplete,
	}

	h.log.Info("Pregnancy date calculation completed", logger.FieldsMap{
		"request_id": reqID,
		"lmp":        req.LastMenstrualPeriod,
	})

	response.WriteJSONResponse(w, reqID, result)
}

// CalculateBMI calculates BMI and weight recommendations
type BMICalcRequest struct {
	HeightCm   float64 `json:"height_cm" validate:"required,gt=0"`
	WeightKg   float64 `json:"weight_kg" validate:"required,gt=0"`
	IsPregnant bool    `json:"is_pregnant"`
}

// BMICalcResult is the response for BMI calculations
type BMICalcResult struct {
	BMI              float64 `json:"bmi"`
	Category         string  `json:"category"`
	RecommendedGain  float64 `json:"recommended_gain,omitempty"`
}

// CalculateBMI calculates BMI and weight recommendations
func (h *PregnancyHandler) CalculateBMI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req BMICalcRequest
	if err := h.ParseRequest(r, &req); err != nil {
		h.log.Error("Failed to parse request", logger.FieldsMap{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		h.log.Error("Invalid request", logger.FieldsMap{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Calculate BMI
	bmiInfo, err := h.calcService.CalculateBMI(req.HeightCm, req.WeightKg, req.IsPregnant)
	if err != nil {
		h.log.Error("Failed to calculate BMI", logger.FieldsMap{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Format response
	result := BMICalcResult{
		BMI:             bmiInfo.BMI,
		Category:        bmiInfo.Category,
		RecommendedGain: bmiInfo.RecommendedGain,
	}

	h.log.Info("BMI calculation completed", logger.FieldsMap{
		"request_id": reqID,
		"bmi":        result.BMI,
		"category":   result.Category,
	})

	response.WriteJSONResponse(w, reqID, result)
}
