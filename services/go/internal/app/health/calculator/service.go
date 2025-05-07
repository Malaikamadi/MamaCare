package calculator

import (
	"math"
	"time"

	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// Service provides health-related calculation functionality
type Service struct {
	log logger.Logger
}

// NewService creates a new calculator service
func NewService(log logger.Logger) *Service {
	return &Service{
		log: log,
	}
}

// PregnancyDateInfo contains pregnancy related dates and durations
type PregnancyDateInfo struct {
	LMP                    time.Time `json:"lmp"`                     // Last menstrual period
	ConceptionDate         time.Time `json:"conception_date"`         // Estimated conception date
	ExpectedDeliveryDate   time.Time `json:"expected_delivery_date"`  // Expected delivery date
	GestationalAge         int       `json:"gestational_age"`         // In days
	GestationalAgeWeeks    int       `json:"gestational_age_weeks"`   // Completed weeks
	GestationalAgeDays     int       `json:"gestational_age_days"`    // Remaining days
	CurrentTrimester       int       `json:"current_trimester"`       // Current trimester (1, 2, or 3)
	WeeksRemaining         int       `json:"weeks_remaining"`         // Weeks remaining until due date
	IsPreTerm              bool      `json:"is_pre_term"`             // < 37 weeks
	IsFullTerm             bool      `json:"is_full_term"`            // 39-40 weeks
	DaysUntilFullTerm      int       `json:"days_until_full_term"`    // Days until full term
	PercentageComplete     float64   `json:"percentage_complete"`     // Percentage of pregnancy completed
}

// BMIInfo contains information about Body Mass Index
type BMIInfo struct {
	Height          float64 `json:"height"`           // In meters
	Weight          float64 `json:"weight"`           // In kilograms
	BMI             float64 `json:"bmi"`              // BMI value
	Category        string  `json:"category"`         // Underweight, Normal, Overweight, Obese
	RecommendedGain float64 `json:"recommended_gain"` // Recommended weight gain during pregnancy in kg
}

// CalculateGestationalAge calculates gestational age based on last menstrual period (LMP)
func (s *Service) CalculateGestationalAge(lmp time.Time, referenceDate time.Time) (int, error) {
	if lmp.After(referenceDate) {
		return 0, errorx.New(errorx.BadRequest, "last menstrual period cannot be after reference date")
	}

	// Calculate days between LMP and reference date
	durationDays := int(math.Floor(referenceDate.Sub(lmp).Hours() / 24))
	
	return durationDays, nil
}

// CalculatePregnancyDates calculates all pregnancy-related dates and durations
func (s *Service) CalculatePregnancyDates(lmp time.Time) (*PregnancyDateInfo, error) {
	if lmp.IsZero() {
		return nil, errorx.New(errorx.BadRequest, "invalid last menstrual period date")
	}

	now := time.Now()
	
	// Standard pregnancy calculations based on Naegele's rule
	// Average cycle is 28 days, ovulation around day 14, conception typically occurs around that time
	conceptionDate := lmp.AddDate(0, 0, 14)
	
	// EDD is 280 days (40 weeks) from LMP
	edd := lmp.AddDate(0, 0, 280)
	
	// Calculate gestational age in days
	gestationalDays, err := s.CalculateGestationalAge(lmp, now)
	if err != nil {
		return nil, err
	}
	
	// Convert to weeks and days
	gestationalWeeks := gestationalDays / 7
	remainingDays := gestationalDays % 7
	
	// Calculate trimester
	var trimester int
	switch {
	case gestationalWeeks < 13:
		trimester = 1
	case gestationalWeeks < 27:
		trimester = 2
	default:
		trimester = 3
	}
	
	// Calculate weeks remaining
	weeksRemaining := 40 - gestationalWeeks
	if remainingDays > 0 {
		weeksRemaining-- // Adjust if we have partial weeks
	}
	
	// Term classification
	isPreTerm := gestationalWeeks < 37
	isFullTerm := gestationalWeeks >= 39 && gestationalWeeks <= 40
	
	// Days until full term (39 weeks)
	daysUntilFullTerm := (39 * 7) - gestationalDays
	if daysUntilFullTerm < 0 {
		daysUntilFullTerm = 0
	}
	
	// Percentage complete
	percentageComplete := float64(gestationalDays) / 280.0 * 100.0
	if percentageComplete > 100.0 {
		percentageComplete = 100.0
	}
	
	return &PregnancyDateInfo{
		LMP:                  lmp,
		ConceptionDate:       conceptionDate,
		ExpectedDeliveryDate: edd,
		GestationalAge:       gestationalDays,
		GestationalAgeWeeks:  gestationalWeeks,
		GestationalAgeDays:   remainingDays,
		CurrentTrimester:     trimester,
		WeeksRemaining:       weeksRemaining,
		IsPreTerm:            isPreTerm,
		IsFullTerm:           isFullTerm,
		DaysUntilFullTerm:    daysUntilFullTerm,
		PercentageComplete:   percentageComplete,
	}, nil
}

// CalculateBMI calculates Body Mass Index and provides weight-related information
func (s *Service) CalculateBMI(heightCm, weightKg float64, isPregnant bool) (*BMIInfo, error) {
	if heightCm <= 0 || weightKg <= 0 {
		return nil, errorx.New(errorx.BadRequest, "height and weight must be positive values")
	}
	
	// Convert height from cm to meters
	heightM := heightCm / 100.0
	
	// Calculate BMI: weight (kg) / height² (m²)
	bmi := weightKg / (heightM * heightM)
	
	// Determine BMI category
	var category string
	var recommendedGain float64
	
	switch {
	case bmi < 18.5:
		category = "underweight"
		recommendedGain = 12.5 // 12.5-18 kg (using lower bound)
	case bmi < 25.0:
		category = "normal"
		recommendedGain = 11.5 // 11.5-16 kg (using lower bound)
	case bmi < 30.0:
		category = "overweight"
		recommendedGain = 7.0 // 7-11.5 kg (using lower bound)
	default:
		category = "obese"
		recommendedGain = 5.0 // 5-9 kg (using lower bound)
	}
	
	// If not pregnant, no recommended gain
	if !isPregnant {
		recommendedGain = 0.0
	}
	
	return &BMIInfo{
		Height:          heightM,
		Weight:          weightKg,
		BMI:             bmi,
		Category:        category,
		RecommendedGain: recommendedGain,
	}, nil
}

// CalculateExpectedDeliveryDate calculates the expected delivery date based on LMP
func (s *Service) CalculateExpectedDeliveryDate(lmp time.Time) (time.Time, error) {
	if lmp.IsZero() {
		return time.Time{}, errorx.New(errorx.BadRequest, "invalid last menstrual period date")
	}
	
	// Naegele's rule: EDD = LMP + 1 year - 3 months + 7 days
	// Simplified: EDD = LMP + 280 days
	return lmp.AddDate(0, 0, 280), nil
}

// EstimateFundusHeight estimates the expected fundus height based on gestational age
// Returns the height in centimeters and an acceptable range
func (s *Service) EstimateFundusHeight(gestationalWeeks int) (float64, float64, float64, error) {
	if gestationalWeeks < 16 || gestationalWeeks > 40 {
		return 0, 0, 0, errorx.New(errorx.BadRequest, "gestational age must be between 16 and 40 weeks")
	}
	
	// Simplified McDonald's rule: After 24 weeks, fundus height in cm often correlates with weeks of gestation
	// Before 24 weeks, this is less reliable but we'll provide estimates
	expectedHeight := float64(gestationalWeeks - 4) // Simple approximation
	if gestationalWeeks < 20 {
		expectedHeight = float64(gestationalWeeks - 6) // Adjusted for earlier weeks
	}
	
	// Define acceptable range: +/- 2cm is often considered normal variation
	minRange := expectedHeight - 2.0
	maxRange := expectedHeight + 2.0
	
	if minRange < 0 {
		minRange = 0
	}
	
	return expectedHeight, minRange, maxRange, nil
}

// EstimateFetusWeight estimates fetal weight based on gestational age
// Returns the estimated weight in grams and an acceptable range
func (s *Service) EstimateFetusWeight(gestationalWeeks int) (float64, float64, float64, error) {
	if gestationalWeeks < 10 || gestationalWeeks > 42 {
		return 0, 0, 0, errorx.New(errorx.BadRequest, "gestational age must be between 10 and 42 weeks")
	}
	
	// Simple estimation based on standard growth charts
	// This is a simplified approximation; actual calculations would use more complex formulas
	var estimatedWeight float64
	
	switch {
	case gestationalWeeks <= 12:
		estimatedWeight = float64((gestationalWeeks - 8) * 10) // ~10g gain per week in first trimester
	case gestationalWeeks <= 28:
		estimatedWeight = 100.0 + float64((gestationalWeeks - 12) * 100) // ~100g gain per week in second trimester
	default:
		// Third trimester: weight gain accelerates
		baseWeight := 1700.0 // Approximate weight at 28 weeks
		weeksPast28 := gestationalWeeks - 28
		weeklyGain := []float64{200, 200, 200, 225, 225, 225, 250, 250, 250, 225, 225, 200, 200, 175}
		
		for i := 0; i < weeksPast28 && i < len(weeklyGain); i++ {
			baseWeight += weeklyGain[i]
		}
		
		estimatedWeight = baseWeight
	}
	
	// Define normal range (10th to 90th percentile, approx +/- 15%)
	minRange := estimatedWeight * 0.85
	maxRange := estimatedWeight * 1.15
	
	return estimatedWeight, minRange, maxRange, nil
}
