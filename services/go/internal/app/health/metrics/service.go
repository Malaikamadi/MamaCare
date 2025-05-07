package metrics

import (
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// ReferenceRange defines normal ranges for a specific vital sign
type ReferenceRange struct {
	Min float64
	Max float64
}

// MetricAnalysis represents the analysis result of a health metric
type MetricAnalysis struct {
	MetricID       uuid.UUID          `json:"metric_id"`
	MotherID       uuid.UUID          `json:"mother_id"`
	AnalysisDate   time.Time          `json:"analysis_date"`
	Abnormalities  map[string]string  `json:"abnormalities,omitempty"`
	Trends         map[string]string  `json:"trends,omitempty"`
	RecommendedActions []string       `json:"recommended_actions,omitempty"`
	SeverityLevel  string             `json:"severity_level,omitempty"`
}

// Service provides health metric analysis functionality
type Service struct {
	log logger.Logger
}

// NewService creates a new metrics analysis service
func NewService(log logger.Logger) *Service {
	return &Service{
		log: log,
	}
}

// AnalyzeMetric performs comprehensive analysis on a single health metric
func (s *Service) AnalyzeMetric(metric *model.HealthMetric, gestationalAge int) (*MetricAnalysis, error) {
	if metric == nil {
		return nil, errorx.New(errorx.BadRequest, "metric data required for analysis")
	}

	analysis := &MetricAnalysis{
		MetricID:      metric.ID,
		MotherID:      metric.MotherID,
		AnalysisDate:  time.Now(),
		Abnormalities: make(map[string]string),
		Trends:        make(map[string]string),
		RecommendedActions: []string{},
		SeverityLevel: "normal",
	}

	// Check vital signs for abnormalities
	s.analyzeBloodPressure(metric, analysis, gestationalAge)
	s.analyzeFetalHeartRate(metric, analysis, gestationalAge)
	s.analyzeFetalMovement(metric, analysis, gestationalAge)
	s.analyzeBloodSugar(metric, analysis, gestationalAge)
	s.analyzeHemoglobin(metric, analysis, gestationalAge)
	s.analyzeWeight(metric, analysis, gestationalAge)

	// Determine overall severity level
	s.determineSeverityLevel(analysis)

	return analysis, nil
}

// AnalyzeMetricHistory analyzes a series of health metrics to identify trends
func (s *Service) AnalyzeMetricHistory(metrics []*model.HealthMetric, gestationalAge int) (*MetricAnalysis, error) {
	if len(metrics) == 0 {
		return nil, errorx.New(errorx.BadRequest, "metric history required for trend analysis")
	}

	// Sort metrics by recording date (newest first)
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].RecordedAt.After(metrics[j].RecordedAt)
	})

	// Use the most recent metric as the basis for analysis
	latestMetric := metrics[0]
	analysis, err := s.AnalyzeMetric(latestMetric, gestationalAge)
	if err != nil {
		return nil, err
	}

	// Analyze trends over time
	s.analyzeBloodPressureTrend(metrics, analysis)
	s.analyzeWeightTrend(metrics, analysis)
	s.analyzeFetalHeartRateTrend(metrics, analysis)

	return analysis, nil
}

// analyzeBloodPressure analyzes blood pressure against reference ranges
func (s *Service) analyzeBloodPressure(metric *model.HealthMetric, analysis *MetricAnalysis, gestationalAge int) {
	if metric.VitalSigns.BloodPressure == nil {
		return
	}

	bp := metric.VitalSigns.BloodPressure
	systolic := bp.Systolic
	diastolic := bp.Diastolic

	// Check for hypertension
	if systolic >= 140 || diastolic >= 90 {
		if systolic >= 160 || diastolic >= 110 {
			analysis.Abnormalities["blood_pressure"] = "severe hypertension"
			analysis.RecommendedActions = append(analysis.RecommendedActions, 
				"Seek immediate medical attention for severe high blood pressure")
		} else {
			analysis.Abnormalities["blood_pressure"] = "hypertension"
			analysis.RecommendedActions = append(analysis.RecommendedActions, 
				"Schedule follow-up appointment to monitor blood pressure")
		}
	}

	// Check for hypotension
	if systolic < 90 || diastolic < 60 {
		analysis.Abnormalities["blood_pressure"] = "hypotension"
		analysis.RecommendedActions = append(analysis.RecommendedActions, 
			"Monitor for dizziness and ensure adequate hydration")
	}
}

// analyzeFetalHeartRate analyzes fetal heart rate against reference ranges
func (s *Service) analyzeFetalHeartRate(metric *model.HealthMetric, analysis *MetricAnalysis, gestationalAge int) {
	if metric.VitalSigns.FetalHeartRate == nil {
		return
	}

	fhr := *metric.VitalSigns.FetalHeartRate

	// Normal fetal heart rate: 110-160 bpm
	if fhr < 110 {
		analysis.Abnormalities["fetal_heart_rate"] = "bradycardia"
		analysis.RecommendedActions = append(analysis.RecommendedActions, 
			"Seek immediate medical attention for low fetal heart rate")
	} else if fhr > 160 {
		analysis.Abnormalities["fetal_heart_rate"] = "tachycardia"
		analysis.RecommendedActions = append(analysis.RecommendedActions, 
			"Seek medical attention for high fetal heart rate")
	}
}

// analyzeFetalMovement analyzes fetal movement against reference ranges
func (s *Service) analyzeFetalMovement(metric *model.HealthMetric, analysis *MetricAnalysis, gestationalAge int) {
	// Skip if no fetal movement data or gestational age < 24 weeks (movements typically felt after 20-24 weeks)
	if metric.VitalSigns.FetalMovement == nil || gestationalAge < 24 {
		return
	}

	movement := *metric.VitalSigns.FetalMovement

	// Normal is at least 10 movements in counting session
	if movement < 10 {
		analysis.Abnormalities["fetal_movement"] = "reduced"
		analysis.RecommendedActions = append(analysis.RecommendedActions, 
			"Continue monitoring fetal movement; seek medical attention if consistently decreased")
	}
	
	// Very low movement is a concern
	if movement < 3 {
		analysis.Abnormalities["fetal_movement"] = "severely reduced"
		analysis.RecommendedActions = append(analysis.RecommendedActions, 
			"Seek immediate medical attention for severely reduced fetal movement")
	}
}

// analyzeBloodSugar analyzes blood sugar against reference ranges
func (s *Service) analyzeBloodSugar(metric *model.HealthMetric, analysis *MetricAnalysis, gestationalAge int) {
	if metric.VitalSigns.BloodSugar == nil {
		return
	}

	bloodSugar := *metric.VitalSigns.BloodSugar

	// Fasting blood sugar thresholds for pregnancy
	// Different thresholds exist for different testing methods and conditions
	// Using general guidelines here
	if bloodSugar > 95 { // Fasting
		analysis.Abnormalities["blood_sugar"] = "elevated"
		analysis.RecommendedActions = append(analysis.RecommendedActions, 
			"Follow up with healthcare provider to discuss blood sugar management")
	}
	
	// Very high values require immediate attention
	if bloodSugar > 180 {
		analysis.Abnormalities["blood_sugar"] = "severely elevated"
		analysis.RecommendedActions = append(analysis.RecommendedActions, 
			"Seek medical attention for very high blood sugar")
	}
}

// analyzeHemoglobin analyzes hemoglobin against reference ranges
func (s *Service) analyzeHemoglobin(metric *model.HealthMetric, analysis *MetricAnalysis, gestationalAge int) {
	if metric.VitalSigns.HemoglobinLevel == nil {
		return
	}

	hemoglobin := *metric.VitalSigns.HemoglobinLevel

	// Hemoglobin thresholds for anemia in pregnancy
	if hemoglobin < 11 {
		if hemoglobin < 7 {
			analysis.Abnormalities["hemoglobin"] = "severe anemia"
			analysis.RecommendedActions = append(analysis.RecommendedActions, 
				"Seek medical attention for severe anemia")
		} else {
			analysis.Abnormalities["hemoglobin"] = "anemia"
			analysis.RecommendedActions = append(analysis.RecommendedActions, 
				"Discuss iron supplementation with healthcare provider")
		}
	}
}

// analyzeWeight analyzes weight and weight changes
func (s *Service) analyzeWeight(metric *model.HealthMetric, analysis *MetricAnalysis, gestationalAge int) {
	if metric.VitalSigns.Weight == nil {
		return
	}

	weight := *metric.VitalSigns.Weight

	// Simple weight check for now - in a real implementation, this would
	// consider pre-pregnancy weight, BMI, and expected weight gain by trimester
	if weight < 45 {
		analysis.Abnormalities["weight"] = "underweight"
		analysis.RecommendedActions = append(analysis.RecommendedActions, 
			"Discuss nutrition and weight gain with healthcare provider")
	}
}

// analyzeBloodPressureTrend analyzes blood pressure trends over time
func (s *Service) analyzeBloodPressureTrend(metrics []*model.HealthMetric, analysis *MetricAnalysis) {
	if len(metrics) < 3 {
		return // Need at least 3 readings to establish a trend
	}

	// Check for consistently rising blood pressure
	risingSystolic := true
	risingDiastolic := true

	for i := 0; i < len(metrics)-1; i++ {
		current := metrics[i].VitalSigns.BloodPressure
		previous := metrics[i+1].VitalSigns.BloodPressure

		if current == nil || previous == nil {
			continue
		}

		if current.Systolic <= previous.Systolic {
			risingSystolic = false
		}

		if current.Diastolic <= previous.Diastolic {
			risingDiastolic = false
		}
	}

	if risingSystolic && risingDiastolic {
		analysis.Trends["blood_pressure"] = "consistently rising"
		analysis.RecommendedActions = append(analysis.RecommendedActions, 
			"Monitor increasing blood pressure trend closely")
	}
}

// analyzeWeightTrend analyzes weight trends over time
func (s *Service) analyzeWeightTrend(metrics []*model.HealthMetric, analysis *MetricAnalysis) {
	if len(metrics) < 3 {
		return // Need at least 3 readings to establish a trend
	}

	// Check if weight is consistently declining or not increasing appropriately
	weights := []float64{}
	for _, metric := range metrics {
		if metric.VitalSigns.Weight != nil {
			weights = append(weights, *metric.VitalSigns.Weight)
		}
	}

	if len(weights) < 3 {
		return
	}

	// Check for weight loss or insufficient gain during pregnancy
	decreasingWeight := true
	for i := 0; i < len(weights)-1; i++ {
		if weights[i] >= weights[i+1] {
			decreasingWeight = false
			break
		}
	}

	if decreasingWeight {
		analysis.Trends["weight"] = "decreasing"
		analysis.RecommendedActions = append(analysis.RecommendedActions, 
			"Consult healthcare provider about weight loss during pregnancy")
	}
}

// analyzeFetalHeartRateTrend analyzes fetal heart rate trends
func (s *Service) analyzeFetalHeartRateTrend(metrics []*model.HealthMetric, analysis *MetricAnalysis) {
	if len(metrics) < 3 {
		return
	}

	// Extract fetal heart rates
	rates := []float64{}
	for _, metric := range metrics {
		if metric.VitalSigns.FetalHeartRate != nil {
			rates = append(rates, *metric.VitalSigns.FetalHeartRate)
		}
	}

	if len(rates) < 3 {
		return
	}

	// Check for consistently declining fetal heart rate
	decliningRate := true
	for i := 0; i < len(rates)-1; i++ {
		if rates[i] >= rates[i+1] {
			decliningRate = false
			break
		}
	}

	if decliningRate {
		analysis.Trends["fetal_heart_rate"] = "declining"
		analysis.RecommendedActions = append(analysis.RecommendedActions, 
			"Consult healthcare provider about decreasing fetal heart rate")
	}
}

// determineSeverityLevel sets the overall severity level based on abnormalities
func (s *Service) determineSeverityLevel(analysis *MetricAnalysis) {
	// Start with normal and escalate based on findings
	severity := "normal"

	// Check for urgent conditions
	for _, abnormality := range analysis.Abnormalities {
		if abnormality == "severe hypertension" || 
		   abnormality == "severely reduced" || 
		   abnormality == "severe anemia" ||
		   abnormality == "bradycardia" {
			severity = "urgent"
			break
		}
	}

	// If not urgent, check for concerning conditions
	if severity == "normal" {
		for _, abnormality := range analysis.Abnormalities {
			if abnormality == "hypertension" || 
			   abnormality == "reduced" || 
			   abnormality == "anemia" ||
			   abnormality == "tachycardia" {
				severity = "concerning"
				break
			}
		}
	}

	// If we have concerning trends, might elevate severity
	if severity == "normal" {
		for _, trend := range analysis.Trends {
			if trend == "consistently rising" || 
			   trend == "decreasing" || 
			   trend == "declining" {
				severity = "monitor"
				break
			}
		}
	}

	analysis.SeverityLevel = severity
}
