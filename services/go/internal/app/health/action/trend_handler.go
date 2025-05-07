package action

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/health/metrics"
	"github.com/mamacare/services/internal/app/health/trend"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/domain/repository"
	"github.com/mamacare/services/internal/port/hasura"
	"github.com/mamacare/services/internal/port/response"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// TrendAnalysisRequest is the request for trend analysis
type TrendAnalysisRequest struct {
	MotherID     uuid.UUID `json:"mother_id" validate:"required"`
	TimeRangeDays int      `json:"time_range_days" validate:"required,min=1,max=365"`
}

// TrendResult represents a single detected trend
type TrendResult struct {
	MetricName        string  `json:"metric_name"`
	TrendType         string  `json:"trend_type"`
	AlertLevel        string  `json:"alert_level"`
	Description       string  `json:"description"`
	RecommendedAction string  `json:"recommended_action,omitempty"`
	FirstValue        float64 `json:"first_value"`
	LastValue         float64 `json:"last_value"`
	ChangeRate        float64 `json:"change_rate"`
	ChangePerDay      float64 `json:"change_per_day"`
}

// TrendAnalysisResult is the response for trend analysis
type TrendAnalysisResult struct {
	MotherID      uuid.UUID     `json:"mother_id"`
	AnalysisDate  string        `json:"analysis_date"`
	DataStartDate string        `json:"data_start_date"`
	DataEndDate   string        `json:"data_end_date"`
	DataPoints    int           `json:"data_points"`
	Trends        []TrendResult `json:"trends"`
	HighestAlert  string        `json:"highest_alert"`
}

// TrendHandler handles trend analysis actions
type TrendHandler struct {
	*hasura.BaseActionHandler
	trendService     *trend.Service
	healthMetricRepo repository.HealthMetricRepository
	log              logger.Logger
}

// NewTrendHandler creates a new trend analysis handler
func NewTrendHandler(
	log logger.Logger,
	trendService *trend.Service,
	healthMetricRepo repository.HealthMetricRepository,
) *TrendHandler {
	return &TrendHandler{
		BaseActionHandler: hasura.NewBaseActionHandler(log),
		trendService:      trendService,
		healthMetricRepo:  healthMetricRepo,
		log:              log,
	}
}

// AnalyzeTrends analyzes health metric trends for a mother
func (h *TrendHandler) AnalyzeTrends(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req TrendAnalysisRequest
	_, err := h.ParseRequest(r, &req)
	if err != nil {
		h.log.Error("Failed to parse request", logger.FieldsMap{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Get date range for metrics
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -req.TimeRangeDays)

	// Retrieve health metrics for the given time range
	metrics, err := h.retrieveHealthMetrics(ctx, req.MotherID, startDate, endDate)
	if err != nil {
		h.log.Error("Failed to retrieve health metrics", logger.Fields{
			"request_id": reqID,
			"mother_id":  req.MotherID.String(),
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	if len(metrics) < 3 {
		h.log.Warn("Insufficient data for trend analysis", logger.Fields{
			"request_id": reqID,
			"mother_id":  req.MotherID.String(),
			"metrics_count": len(metrics),
		})
		response.WriteError(w, reqID, 
			errorx.NewError(errorx.BadRequest, "Insufficient data for trend analysis. At least 3 measurements are required"))
		return
	}

	// Perform trend analysis
	analysis, err := h.trendService.AnalyzeTrends(req.MotherID, metrics)
	if err != nil {
		h.log.Error("Failed to analyze trends", logger.Fields{
			"request_id": reqID,
			"mother_id":  req.MotherID.String(),
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Convert trend results to response format
	trends := make([]TrendResult, 0, len(analysis.Trends))
	for _, t := range analysis.Trends {
		trends = append(trends, TrendResult{
			MetricName:        t.MetricName,
			TrendType:         string(t.TrendType),
			AlertLevel:        string(t.AlertLevel),
			Description:       t.Description,
			RecommendedAction: t.RecommendedAction,
			FirstValue:        t.FirstValue,
			LastValue:         t.LastValue,
			ChangeRate:        t.ChangeRate,
			ChangePerDay:      t.ChangePerDay,
		})
	}

	// Prepare the response
	result := TrendAnalysisResult{
		MotherID:      analysis.MotherID,
		AnalysisDate:  analysis.AnalysisDate.Format(time.RFC3339),
		DataStartDate: analysis.DataStartDate.Format(time.RFC3339),
		DataEndDate:   analysis.DataEndDate.Format(time.RFC3339),
		DataPoints:    analysis.DataPoints,
		Trends:        trends,
		HighestAlert:  string(analysis.HighestAlert),
	}

	h.log.Info("Trend analysis completed", logger.Fields{
		"request_id":    reqID,
		"mother_id":     req.MotherID.String(),
		"trends_count":  len(trends),
		"highest_alert": string(analysis.HighestAlert),
	})

	response.WriteJSONResponse(w, reqID, result)
}

// retrieveHealthMetrics retrieves health metrics for a given mother and time range
func (h *TrendHandler) retrieveHealthMetrics(
	ctx context.Context,
	motherID uuid.UUID,
	startDate, endDate time.Time,
) ([]*model.HealthMetric, error) {
	// Assuming this method should be implemented on the repository
	return h.healthMetricRepo.FindByMotherIDAndTimeRange(ctx, motherID, startDate, endDate)
}

// GetMetricsAnalysis processes and analyzes a single health metric
type MetricsAnalysisRequest struct {
	MetricID        uuid.UUID `json:"metric_id" validate:"required"`
	GestationalAge  int       `json:"gestational_age" validate:"required,min=0,max=45"`
}

// MetricsAnalysisResult is the response for metrics analysis
type MetricsAnalysisResult struct {
	MetricID           uuid.UUID              `json:"metric_id"`
	MotherID           uuid.UUID              `json:"mother_id"`
	AnalysisDate       string                 `json:"analysis_date"`
	Abnormalities      map[string]string      `json:"abnormalities,omitempty"`
	RecommendedActions []string               `json:"recommended_actions,omitempty"`
	SeverityLevel      string                 `json:"severity_level"`
}

// AnalyzeMetric analyzes a single health metric
func (h *TrendHandler) AnalyzeMetric(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req MetricsAnalysisRequest
	if err := h.ParseRequest(r, &req); err != nil {
		h.log.Error("Failed to parse request", logger.FieldsMap{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Retrieve the health metric
	metric, err := h.healthMetricRepo.GetByID(ctx, req.MetricID)
	if err != nil {
		h.log.Error("Failed to retrieve health metric", logger.FieldsMap{
			"request_id": reqID,
			"metric_id":  req.MetricID.String(),
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.NotFound, "Health metric not found"))
		return
	}

	// Analyze the metric
	analysisService := metrics.NewService(h.log)
	analysis, err := analysisService.AnalyzeMetric(metric, req.GestationalAge)
	if err != nil {
		h.log.Error("Failed to analyze metric", logger.FieldsMap{
			"request_id": reqID,
			"metric_id":  req.MetricID.String(),
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Prepare the response
	result := MetricsAnalysisResult{
		MetricID:           analysis.MetricID,
		MotherID:           analysis.MotherID,
		AnalysisDate:       analysis.AnalysisDate.Format(time.RFC3339),
		Abnormalities:      analysis.Abnormalities,
		RecommendedActions: analysis.RecommendedActions,
		SeverityLevel:      analysis.SeverityLevel,
	}

	h.log.Info("Metric analysis completed", logger.FieldsMap{
		"request_id":     reqID,
		"metric_id":      req.MetricID.String(),
		"severity_level": analysis.SeverityLevel,
	})

	response.WriteJSONResponse(w, reqID, result)
}
