package trend

import (
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// TrendType defines the type of trend detected
type TrendType string

const (
	// TrendIncreasing indicates an increasing trend
	TrendIncreasing TrendType = "increasing"
	// TrendDecreasing indicates a decreasing trend
	TrendDecreasing TrendType = "decreasing"
	// TrendStable indicates a stable trend
	TrendStable TrendType = "stable"
	// TrendFluctuating indicates a fluctuating trend
	TrendFluctuating TrendType = "fluctuating"
	// TrendInsufficient indicates insufficient data to determine a trend
	TrendInsufficient TrendType = "insufficient_data"
)

// AlertLevel defines the level of alert for a detected trend
type AlertLevel string

const (
	// AlertNone indicates no alert needed
	AlertNone AlertLevel = "none"
	// AlertMonitor indicates the trend should be monitored
	AlertMonitor AlertLevel = "monitor"
	// AlertConcern indicates the trend is concerning
	AlertConcern AlertLevel = "concern"
	// AlertUrgent indicates the trend requires urgent attention
	AlertUrgent AlertLevel = "urgent"
)

// TrendResult represents the detected trend in a specific metric
type TrendResult struct {
	MetricName       string    `json:"metric_name"`
	TrendType        TrendType `json:"trend_type"`
	AlertLevel       AlertLevel `json:"alert_level"`
	Description      string    `json:"description"`
	RecommendedAction string    `json:"recommended_action,omitempty"`
	FirstValue       float64   `json:"first_value"`
	LastValue        float64   `json:"last_value"`
	ChangeRate       float64   `json:"change_rate"`       // Percentage change
	ChangePerDay     float64   `json:"change_per_day"`    // Absolute change per day
}

// TrendAnalysis represents a comprehensive trend analysis for a mother's health metrics
type TrendAnalysis struct {
	MotherID       uuid.UUID     `json:"mother_id"`
	AnalysisDate   time.Time     `json:"analysis_date"`
	DataStartDate  time.Time     `json:"data_start_date"`
	DataEndDate    time.Time     `json:"data_end_date"`
	DataPoints     int           `json:"data_points"`
	Trends         []TrendResult `json:"trends"`
	HighestAlert   AlertLevel    `json:"highest_alert"`
}

// TimeSeriesData represents time series data for analysis
type TimeSeriesData struct {
	Dates  []time.Time
	Values []float64
}

// Service provides trend detection functionality
type Service struct {
	log logger.Logger
	
	// Configuration values for trend detection
	minDataPoints           int
	significantChangePercent float64
	significantChangeBP     float64
	significantChangeWeight float64
	significantChangeFHR    float64
}

// NewService creates a new trend detection service
func NewService(log logger.Logger) *Service {
	return &Service{
		log:                     log,
		minDataPoints:           3,
		significantChangePercent: 10.0,
		significantChangeBP:     10.0,  // 10 mmHg change in blood pressure is significant
		significantChangeWeight: 2.0,   // 2 kg change in weight in short period is significant
		significantChangeFHR:    10.0,  // 10 bpm change in fetal heart rate is significant
	}
}

// AnalyzeTrends performs trend analysis on a series of health metrics
func (s *Service) AnalyzeTrends(motherID uuid.UUID, metrics []*model.HealthMetric) (*TrendAnalysis, error) {
	if len(metrics) < s.minDataPoints {
		return nil, errorx.New(errorx.BadRequest, "insufficient data points for trend analysis")
	}

	// Sort metrics by date (oldest first for time series analysis)
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].RecordedAt.Before(metrics[j].RecordedAt)
	})

	analysis := &TrendAnalysis{
		MotherID:      motherID,
		AnalysisDate:  time.Now(),
		DataStartDate: metrics[0].RecordedAt,
		DataEndDate:   metrics[len(metrics)-1].RecordedAt,
		DataPoints:    len(metrics),
		Trends:        []TrendResult{},
		HighestAlert:  AlertNone,
	}

	// Extract time series data for various metrics
	bpSystolicData := s.extractBloodPressureSystolic(metrics)
	bpDiastolicData := s.extractBloodPressureDiastolic(metrics)
	weightData := s.extractWeight(metrics)
	fetalHeartRateData := s.extractFetalHeartRate(metrics)
	
	// Analyze individual metrics
	if len(bpSystolicData.Values) >= s.minDataPoints {
		trend := s.analyzeSystolicTrend(bpSystolicData)
		analysis.Trends = append(analysis.Trends, trend)
		s.updateHighestAlert(analysis, trend.AlertLevel)
	}
	
	if len(bpDiastolicData.Values) >= s.minDataPoints {
		trend := s.analyzeDiastolicTrend(bpDiastolicData)
		analysis.Trends = append(analysis.Trends, trend)
		s.updateHighestAlert(analysis, trend.AlertLevel)
	}
	
	if len(weightData.Values) >= s.minDataPoints {
		trend := s.analyzeWeightTrend(weightData)
		analysis.Trends = append(analysis.Trends, trend)
		s.updateHighestAlert(analysis, trend.AlertLevel)
	}
	
	if len(fetalHeartRateData.Values) >= s.minDataPoints {
		trend := s.analyzeFetalHeartRateTrend(fetalHeartRateData)
		analysis.Trends = append(analysis.Trends, trend)
		s.updateHighestAlert(analysis, trend.AlertLevel)
	}

	return analysis, nil
}

// extractBloodPressureSystolic extracts systolic blood pressure data
func (s *Service) extractBloodPressureSystolic(metrics []*model.HealthMetric) TimeSeriesData {
	var result TimeSeriesData
	
	for _, metric := range metrics {
		if metric.VitalSigns.BloodPressure != nil {
			result.Dates = append(result.Dates, metric.RecordedAt)
			result.Values = append(result.Values, metric.VitalSigns.BloodPressure.Systolic)
		}
	}
	
	return result
}

// extractBloodPressureDiastolic extracts diastolic blood pressure data
func (s *Service) extractBloodPressureDiastolic(metrics []*model.HealthMetric) TimeSeriesData {
	var result TimeSeriesData
	
	for _, metric := range metrics {
		if metric.VitalSigns.BloodPressure != nil {
			result.Dates = append(result.Dates, metric.RecordedAt)
			result.Values = append(result.Values, metric.VitalSigns.BloodPressure.Diastolic)
		}
	}
	
	return result
}

// extractWeight extracts weight measurement data
func (s *Service) extractWeight(metrics []*model.HealthMetric) TimeSeriesData {
	var result TimeSeriesData
	
	for _, metric := range metrics {
		if metric.VitalSigns.Weight != nil {
			result.Dates = append(result.Dates, metric.RecordedAt)
			result.Values = append(result.Values, *metric.VitalSigns.Weight)
		}
	}
	
	return result
}

// extractFetalHeartRate extracts fetal heart rate data
func (s *Service) extractFetalHeartRate(metrics []*model.HealthMetric) TimeSeriesData {
	var result TimeSeriesData
	
	for _, metric := range metrics {
		if metric.VitalSigns.FetalHeartRate != nil {
			result.Dates = append(result.Dates, metric.RecordedAt)
			result.Values = append(result.Values, *metric.VitalSigns.FetalHeartRate)
		}
	}
	
	return result
}

// analyzeSystolicTrend analyzes trends in systolic blood pressure
func (s *Service) analyzeSystolicTrend(data TimeSeriesData) TrendResult {
	result := TrendResult{
		MetricName: "systolic_blood_pressure",
		FirstValue: data.Values[0],
		LastValue:  data.Values[len(data.Values)-1],
	}
	
	// Calculate trend type
	trendType, changeRate, changePerDay := s.calculateTrend(data)
	result.TrendType = trendType
	result.ChangeRate = changeRate
	result.ChangePerDay = changePerDay
	
	// Determine alert level and recommendations based on trend
	if trendType == TrendIncreasing {
		// Increasing systolic pressure is concerning, especially if already high
		if result.LastValue >= 140 {
			result.AlertLevel = AlertUrgent
			result.Description = "Systolic blood pressure increasing and has reached hypertensive levels"
			result.RecommendedAction = "Seek immediate medical attention for hypertension"
		} else if result.LastValue >= 130 {
			result.AlertLevel = AlertConcern
			result.Description = "Systolic blood pressure increasing and approaching hypertensive levels"
			result.RecommendedAction = "Consult healthcare provider about rising blood pressure"
		} else if changeRate > 5 {
			result.AlertLevel = AlertMonitor
			result.Description = "Systolic blood pressure showing significant upward trend"
			result.RecommendedAction = "Monitor blood pressure closely and report continued increases"
		} else {
			result.AlertLevel = AlertNone
			result.Description = "Mild increase in systolic blood pressure, still within normal range"
		}
	} else if trendType == TrendDecreasing {
		// Decreasing is generally good unless it's too low
		if result.LastValue < 90 {
			result.AlertLevel = AlertConcern
			result.Description = "Systolic blood pressure decreasing and has reached hypotensive levels"
			result.RecommendedAction = "Consult healthcare provider about low blood pressure"
		} else {
			result.AlertLevel = AlertNone
			result.Description = "Decreasing systolic blood pressure, trending toward normal range"
		}
	} else {
		// Stable trend
		if result.LastValue >= 140 {
			result.AlertLevel = AlertConcern
			result.Description = "Consistently high systolic blood pressure"
			result.RecommendedAction = "Follow up with healthcare provider about hypertension"
		} else if result.LastValue < 90 {
			result.AlertLevel = AlertMonitor
			result.Description = "Consistently low systolic blood pressure"
			result.RecommendedAction = "Monitor for symptoms of hypotension"
		} else {
			result.AlertLevel = AlertNone
			result.Description = "Stable systolic blood pressure within normal range"
		}
	}
	
	return result
}

// analyzeDiastolicTrend analyzes trends in diastolic blood pressure
func (s *Service) analyzeDiastolicTrend(data TimeSeriesData) TrendResult {
	result := TrendResult{
		MetricName: "diastolic_blood_pressure",
		FirstValue: data.Values[0],
		LastValue:  data.Values[len(data.Values)-1],
	}
	
	// Calculate trend type
	trendType, changeRate, changePerDay := s.calculateTrend(data)
	result.TrendType = trendType
	result.ChangeRate = changeRate
	result.ChangePerDay = changePerDay
	
	// Determine alert level and recommendations based on trend
	if trendType == TrendIncreasing {
		// Increasing diastolic pressure is concerning, especially if already high
		if result.LastValue >= 90 {
			result.AlertLevel = AlertUrgent
			result.Description = "Diastolic blood pressure increasing and has reached hypertensive levels"
			result.RecommendedAction = "Seek immediate medical attention for hypertension"
		} else if result.LastValue >= 85 {
			result.AlertLevel = AlertConcern
			result.Description = "Diastolic blood pressure increasing and approaching hypertensive levels"
			result.RecommendedAction = "Consult healthcare provider about rising blood pressure"
		} else if changeRate > 5 {
			result.AlertLevel = AlertMonitor
			result.Description = "Diastolic blood pressure showing significant upward trend"
			result.RecommendedAction = "Monitor blood pressure closely and report continued increases"
		} else {
			result.AlertLevel = AlertNone
			result.Description = "Mild increase in diastolic blood pressure, still within normal range"
		}
	} else if trendType == TrendDecreasing {
		// Decreasing is generally good unless it's too low
		if result.LastValue < 60 {
			result.AlertLevel = AlertConcern
			result.Description = "Diastolic blood pressure decreasing and has reached hypotensive levels"
			result.RecommendedAction = "Consult healthcare provider about low blood pressure"
		} else {
			result.AlertLevel = AlertNone
			result.Description = "Decreasing diastolic blood pressure, trending toward normal range"
		}
	} else {
		// Stable trend
		if result.LastValue >= 90 {
			result.AlertLevel = AlertConcern
			result.Description = "Consistently high diastolic blood pressure"
			result.RecommendedAction = "Follow up with healthcare provider about hypertension"
		} else if result.LastValue < 60 {
			result.AlertLevel = AlertMonitor
			result.Description = "Consistently low diastolic blood pressure"
			result.RecommendedAction = "Monitor for symptoms of hypotension"
		} else {
			result.AlertLevel = AlertNone
			result.Description = "Stable diastolic blood pressure within normal range"
		}
	}
	
	return result
}

// analyzeWeightTrend analyzes trends in weight
func (s *Service) analyzeWeightTrend(data TimeSeriesData) TrendResult {
	result := TrendResult{
		MetricName: "weight",
		FirstValue: data.Values[0],
		LastValue:  data.Values[len(data.Values)-1],
	}
	
	// Calculate trend type
	trendType, changeRate, changePerDay := s.calculateTrend(data)
	result.TrendType = trendType
	result.ChangeRate = changeRate
	result.ChangePerDay = changePerDay
	
	// Determine alert level and recommendations for weight trends
	// For pregnant women, weight should increase gradually
	if trendType == TrendDecreasing {
		result.AlertLevel = AlertConcern
		result.Description = "Weight decreasing during pregnancy"
		result.RecommendedAction = "Consult healthcare provider about weight loss during pregnancy"
	} else if trendType == TrendIncreasing {
		// Check if weight gain is too rapid
		if changePerDay > 0.2 { // More than 0.2 kg per day is very rapid
			result.AlertLevel = AlertConcern
			result.Description = "Rapid weight gain"
			result.RecommendedAction = "Consult healthcare provider about rapid weight gain"
		} else if changePerDay > 0.1 { // More than 0.1 kg per day is somewhat rapid
			result.AlertLevel = AlertMonitor
			result.Description = "Accelerated weight gain"
			result.RecommendedAction = "Monitor weight gain and discuss with healthcare provider at next visit"
		} else {
			result.AlertLevel = AlertNone
			result.Description = "Steady weight gain"
		}
	} else {
		// No significant change
		result.AlertLevel = AlertMonitor
		result.Description = "Weight stable (minimal change)"
		result.RecommendedAction = "Discuss weight progression with healthcare provider"
	}
	
	return result
}

// analyzeFetalHeartRateTrend analyzes trends in fetal heart rate
func (s *Service) analyzeFetalHeartRateTrend(data TimeSeriesData) TrendResult {
	result := TrendResult{
		MetricName: "fetal_heart_rate",
		FirstValue: data.Values[0],
		LastValue:  data.Values[len(data.Values)-1],
	}
	
	// Calculate trend type
	trendType, changeRate, changePerDay := s.calculateTrend(data)
	result.TrendType = trendType
	result.ChangeRate = changeRate
	result.ChangePerDay = changePerDay
	
	// Normal fetal heart rate range is typically 110-160 bpm
	if trendType == TrendDecreasing {
		if result.LastValue < 110 {
			result.AlertLevel = AlertUrgent
			result.Description = "Fetal heart rate decreasing and below normal range"
			result.RecommendedAction = "Seek immediate medical attention"
		} else if result.LastValue < 120 && changeRate > 5 {
			result.AlertLevel = AlertConcern
			result.Description = "Fetal heart rate decreasing and approaching lower limit of normal range"
			result.RecommendedAction = "Consult healthcare provider promptly"
		} else if changeRate > 10 {
			result.AlertLevel = AlertMonitor
			result.Description = "Significant decrease in fetal heart rate, still within normal range"
			result.RecommendedAction = "Monitor fetal movement and heart rate closely"
		} else {
			result.AlertLevel = AlertNone
			result.Description = "Mild decrease in fetal heart rate, within normal variation"
		}
	} else if trendType == TrendIncreasing {
		if result.LastValue > 160 {
			result.AlertLevel = AlertConcern
			result.Description = "Fetal heart rate increasing and above normal range"
			result.RecommendedAction = "Consult healthcare provider promptly"
		} else if result.LastValue > 150 && changeRate > 5 {
			result.AlertLevel = AlertMonitor
			result.Description = "Fetal heart rate increasing and approaching upper limit of normal range"
			result.RecommendedAction = "Monitor fetal heart rate closely"
		} else {
			result.AlertLevel = AlertNone
			result.Description = "Mild increase in fetal heart rate, within normal variation"
		}
	} else {
		// Stable trend
		if result.LastValue < 110 || result.LastValue > 160 {
			result.AlertLevel = AlertConcern
			result.Description = "Fetal heart rate stable but outside normal range"
			result.RecommendedAction = "Consult healthcare provider promptly"
		} else {
			result.AlertLevel = AlertNone
			result.Description = "Stable fetal heart rate within normal range"
		}
	}
	
	return result
}

// calculateTrend determines the type of trend in the time series data
func (s *Service) calculateTrend(data TimeSeriesData) (TrendType, float64, float64) {
	if len(data.Values) < s.minDataPoints {
		return TrendInsufficient, 0, 0
	}
	
	// Calculate simple linear regression to determine trend
	n := len(data.Values)
	firstValue := data.Values[0]
	lastValue := data.Values[n-1]
	
	// Calculate total days elapsed
	totalDays := data.Dates[n-1].Sub(data.Dates[0]).Hours() / 24
	if totalDays < 1 {
		totalDays = 1 // Avoid division by zero
	}
	
	// Calculate absolute and percentage change
	absoluteChange := lastValue - firstValue
	changePerDay := absoluteChange / totalDays
	
	var percentageChange float64
	if firstValue != 0 {
		percentageChange = (absoluteChange / firstValue) * 100
	}
	
	// Calculate mean and standard deviation to assess stability
	mean := calculateMean(data.Values)
	stdDev := calculateStdDev(data.Values, mean)
	
	// Coefficient of variation as a measure of stability
	var cv float64
	if mean != 0 {
		cv = stdDev / mean
	}
	
	// Determine trend type based on changes and variation
	if math.Abs(percentageChange) < s.significantChangePercent && cv < 0.1 {
		return TrendStable, percentageChange, changePerDay
	} else if cv > 0.2 {
		return TrendFluctuating, percentageChange, changePerDay
	} else if percentageChange > 0 {
		return TrendIncreasing, percentageChange, changePerDay
	} else {
		return TrendDecreasing, percentageChange, changePerDay
	}
}

// updateHighestAlert updates the highest alert level in the analysis
func (s *Service) updateHighestAlert(analysis *TrendAnalysis, alertLevel AlertLevel) {
	// Order of alert levels from highest to lowest:
	// AlertUrgent > AlertConcern > AlertMonitor > AlertNone
	
	if alertLevel == AlertUrgent {
		analysis.HighestAlert = AlertUrgent
	} else if alertLevel == AlertConcern && analysis.HighestAlert != AlertUrgent {
		analysis.HighestAlert = AlertConcern
	} else if alertLevel == AlertMonitor && 
			  analysis.HighestAlert != AlertUrgent && 
			  analysis.HighestAlert != AlertConcern {
		analysis.HighestAlert = AlertMonitor
	}
}

// calculateMean calculates the mean of a series of values
func calculateMean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateStdDev calculates the standard deviation of a series of values
func calculateStdDev(values []float64, mean float64) float64 {
	sumSquaredDiff := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}
	variance := sumSquaredDiff / float64(len(values))
	return math.Sqrt(variance)
}
