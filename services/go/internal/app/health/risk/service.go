package risk

import (
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// Service provides risk calculation functionality for maternal health
type Service struct {
	log logger.Logger
}

// NewService creates a new risk calculation service
func NewService(log logger.Logger) *Service {
	return &Service{
		log: log,
	}
}

// RiskFactors represents identified risk factors for a mother
type RiskFactors struct {
	AgeRelated       []string `json:"age_related,omitempty"`
	MedicalHistory   []string `json:"medical_history,omitempty"`
	ObstetricHistory []string `json:"obstetric_history,omitempty"`
	CurrentVitals    []string `json:"current_vitals,omitempty"`
	Lifestyle        []string `json:"lifestyle,omitempty"`
}

// RiskAssessment represents the result of a risk assessment
type RiskAssessment struct {
	MotherID    uuid.UUID   `json:"mother_id"`
	RiskLevel   model.RiskLevel `json:"risk_level"`
	RiskScore   int         `json:"risk_score"`
	RiskFactors RiskFactors `json:"risk_factors"`
	AssessedAt  time.Time   `json:"assessed_at"`
}

// CalculateRisk performs a comprehensive risk assessment for a mother
func (s *Service) CalculateRisk(mother *model.Mother, recentMetrics []*model.HealthMetric) (*RiskAssessment, error) {
	if mother == nil {
		return nil, errorx.New(errorx.BadRequest, "mother data required for risk assessment")
	}

	assessment := &RiskAssessment{
		MotherID:   mother.ID,
		RiskFactors: RiskFactors{
			AgeRelated:       []string{},
			MedicalHistory:   []string{},
			ObstetricHistory: []string{},
			CurrentVitals:    []string{},
			Lifestyle:        []string{},
		},
		AssessedAt: time.Now(),
	}

	// Calculate risk score based on various factors
	score := 0

	// Age-related risk factors
	score += s.evaluateAgeRisk(mother, assessment)

	// Medical history risk factors
	score += s.evaluateMedicalHistoryRisk(mother, assessment)

	// Obstetric history risk factors
	score += s.evaluateObstetricHistoryRisk(mother, assessment)

	// Current health metrics risk factors
	score += s.evaluateCurrentHealthRisk(mother, recentMetrics, assessment)

	// Set final risk score and risk level
	assessment.RiskScore = score
	assessment.RiskLevel = s.determineRiskLevel(score)

	return assessment, nil
}

// evaluateAgeRisk assesses risks based on maternal age
func (s *Service) evaluateAgeRisk(mother *model.Mother, assessment *RiskAssessment) int {
	// TODO: Get mother's age from user profile
	// For now, we'll use a placeholder implementation
	
	// Example age-related risk factors:
	// - Age < 18: Teenage pregnancy
	// - Age > 35: Advanced maternal age

	riskScore := 0
	
	// For MVP, we'll return a placeholder score
	// In a complete implementation, we would calculate based on actual age
	
	return riskScore
}

// evaluateMedicalHistoryRisk assesses risks based on pre-existing health conditions
func (s *Service) evaluateMedicalHistoryRisk(mother *model.Mother, assessment *RiskAssessment) int {
	riskScore := 0
	
	// Check for high-risk health conditions
	highRiskConditions := map[string]int{
		"diabetes":           3,
		"hypertension":       3,
		"heart disease":      4,
		"kidney disease":     3,
		"thyroid disorder":   2,
		"autoimmune disease": 2,
		"hiv":                3,
		"hepatitis":          2,
		"malaria":            2,
		"anemia":             2,
		"sickle cell":        3,
	}
	
	for _, condition := range mother.HealthConditions {
		if score, exists := highRiskConditions[condition]; exists {
			riskScore += score
			assessment.RiskFactors.MedicalHistory = append(
				assessment.RiskFactors.MedicalHistory,
				condition,
			)
		}
	}
	
	// Rh factor incompatibility check
	if mother.BloodType == model.BloodTypeANeg || 
	   mother.BloodType == model.BloodTypeBNeg || 
	   mother.BloodType == model.BloodTypeABNeg || 
	   mother.BloodType == model.BloodTypeONeg {
		riskScore += 2
		assessment.RiskFactors.MedicalHistory = append(
			assessment.RiskFactors.MedicalHistory,
			"rh negative blood type",
		)
	}
	
	return riskScore
}

// evaluateObstetricHistoryRisk assesses risks based on previous pregnancies
func (s *Service) evaluateObstetricHistoryRisk(mother *model.Mother, assessment *RiskAssessment) int {
	riskScore := 0
	history := mother.PregnancyHistory
	
	// Previous cesarean sections
	if history.PreviousCaesareans > 0 {
		riskScore += history.PreviousCaesareans
		assessment.RiskFactors.ObstetricHistory = append(
			assessment.RiskFactors.ObstetricHistory,
			"previous cesarean delivery",
		)
	}
	
	// Grand multiparity (5+ previous deliveries)
	if history.PreviousDeliveries >= 5 {
		riskScore += 2
		assessment.RiskFactors.ObstetricHistory = append(
			assessment.RiskFactors.ObstetricHistory,
			"grand multiparity",
		)
	}
	
	// No previous pregnancies (nulliparity)
	if history.PreviousPregnancies == 0 {
		riskScore += 1
		assessment.RiskFactors.ObstetricHistory = append(
			assessment.RiskFactors.ObstetricHistory,
			"first pregnancy",
		)
	}
	
	// Previous pregnancy complications
	complicationRisks := map[string]int{
		"preeclampsia":        3,
		"eclampsia":           4,
		"gestational diabetes": 3,
		"preterm birth":       3,
		"placenta previa":     3,
		"placental abruption": 4,
		"postpartum hemorrhage": 3,
		"stillbirth":          4,
		"miscarriage":         2,
	}
	
	for _, complication := range history.PreviousComplications {
		if score, exists := complicationRisks[complication]; exists {
			riskScore += score
			assessment.RiskFactors.ObstetricHistory = append(
				assessment.RiskFactors.ObstetricHistory,
				"history of "+complication,
			)
		}
	}
	
	return riskScore
}

// evaluateCurrentHealthRisk assesses risks based on current health metrics
func (s *Service) evaluateCurrentHealthRisk(mother *model.Mother, metrics []*model.HealthMetric, assessment *RiskAssessment) int {
	riskScore := 0
	
	if len(metrics) == 0 {
		return riskScore
	}
	
	// Use the most recent health metric
	latestMetric := metrics[0]
	
	// Check blood pressure
	if latestMetric.VitalSigns.BloodPressure != nil {
		bp := latestMetric.VitalSigns.BloodPressure
		
		// Check for hypertension
		if bp.Systolic >= 140 || bp.Diastolic >= 90 {
			riskScore += 3
			assessment.RiskFactors.CurrentVitals = append(
				assessment.RiskFactors.CurrentVitals,
				"elevated blood pressure",
			)
		}
		
		// Check for hypotension
		if bp.Systolic < 90 || bp.Diastolic < 60 {
			riskScore += 2
			assessment.RiskFactors.CurrentVitals = append(
				assessment.RiskFactors.CurrentVitals,
				"low blood pressure",
			)
		}
	}
	
	// Check fetal heart rate
	if latestMetric.VitalSigns.FetalHeartRate != nil {
		fhr := *latestMetric.VitalSigns.FetalHeartRate
		
		if fhr < 110 || fhr > 160 {
			riskScore += 3
			assessment.RiskFactors.CurrentVitals = append(
				assessment.RiskFactors.CurrentVitals,
				"abnormal fetal heart rate",
			)
		}
	}
	
	// Check hemoglobin (anemia)
	if latestMetric.VitalSigns.HemoglobinLevel != nil {
		hemoglobin := *latestMetric.VitalSigns.HemoglobinLevel
		
		if hemoglobin < 11 {
			riskScore += 2
			assessment.RiskFactors.CurrentVitals = append(
				assessment.RiskFactors.CurrentVitals,
				"anemia",
			)
		}
		
		if hemoglobin < 7 {
			riskScore += 3 // Severe anemia is higher risk
		}
	}
	
	// Check blood sugar
	if latestMetric.VitalSigns.BloodSugar != nil {
		bloodSugar := *latestMetric.VitalSigns.BloodSugar
		
		if bloodSugar > 95 { // Fasting blood sugar
			riskScore += 2
			assessment.RiskFactors.CurrentVitals = append(
				assessment.RiskFactors.CurrentVitals,
				"elevated blood sugar",
			)
		}
	}
	
	return riskScore
}

// determineRiskLevel converts a risk score to a risk level
func (s *Service) determineRiskLevel(score int) model.RiskLevel {
	switch {
	case score >= 10:
		return model.RiskLevelHigh
	case score >= 5:
		return model.RiskLevelMedium
	default:
		return model.RiskLevelLow
	}
}
