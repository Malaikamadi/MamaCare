package model

import (
	"time"

	"github.com/google/uuid"
)

// BloodType represents the blood type of a mother
type BloodType string

const (
	// BloodTypeAPos represents A+ blood type
	BloodTypeAPos BloodType = "A+"
	// BloodTypeANeg represents A- blood type
	BloodTypeANeg BloodType = "A-"
	// BloodTypeBPos represents B+ blood type
	BloodTypeBPos BloodType = "B+"
	// BloodTypeBNeg represents B- blood type
	BloodTypeBNeg BloodType = "B-"
	// BloodTypeABPos represents AB+ blood type
	BloodTypeABPos BloodType = "AB+"
	// BloodTypeABNeg represents AB- blood type
	BloodTypeABNeg BloodType = "AB-"
	// BloodTypeOPos represents O+ blood type
	BloodTypeOPos BloodType = "O+"
	// BloodTypeONeg represents O- blood type
	BloodTypeONeg BloodType = "O-"
	// BloodTypeUnknown represents unknown blood type
	BloodTypeUnknown BloodType = "unknown"
)

// RiskLevel represents the risk level of a pregnancy
type RiskLevel string

const (
	// RiskLevelLow represents low risk level
	RiskLevelLow RiskLevel = "low"
	// RiskLevelMedium represents medium risk level
	RiskLevelMedium RiskLevel = "medium"
	// RiskLevelHigh represents high risk level
	RiskLevelHigh RiskLevel = "high"
)

// PregnancyHistory represents the pregnancy history of a mother
type PregnancyHistory struct {
	PreviousPregnancies   int      `json:"previous_pregnancies"`
	PreviousDeliveries    int      `json:"previous_deliveries"`
	PreviousCaesareans    int      `json:"previous_caesareans"`
	PreviousComplications []string `json:"previous_complications"`
}

// Mother represents a mother enrolled in the system
type Mother struct {
	ID                   uuid.UUID        `json:"id"`
	UserID               uuid.UUID        `json:"user_id"`
	ExpectedDeliveryDate time.Time        `json:"expected_delivery_date"`
	BloodType            BloodType        `json:"blood_type"`
	HealthConditions     []string         `json:"health_conditions"`
	PregnancyHistory     PregnancyHistory `json:"pregnancy_history"`
	RiskLevel            RiskLevel        `json:"risk_level"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
}

// NewMother creates a new mother
func NewMother(id, userID uuid.UUID, expectedDeliveryDate time.Time) *Mother {
	now := time.Now()
	return &Mother{
		ID:                   id,
		UserID:               userID,
		ExpectedDeliveryDate: expectedDeliveryDate,
		BloodType:            BloodTypeUnknown,
		HealthConditions:     []string{},
		RiskLevel:            RiskLevelLow,
		CreatedAt:            now,
		UpdatedAt:            now,
		PregnancyHistory:     PregnancyHistory{},
	}
}

// WithBloodType sets the blood type of the mother
func (m *Mother) WithBloodType(bloodType BloodType) *Mother {
	m.BloodType = bloodType
	return m
}

// WithHealthConditions sets the health conditions of the mother
func (m *Mother) WithHealthConditions(conditions []string) *Mother {
	m.HealthConditions = conditions
	return m
}

// WithPregnancyHistory sets the pregnancy history of the mother
func (m *Mother) WithPregnancyHistory(history PregnancyHistory) *Mother {
	m.PregnancyHistory = history
	return m
}

// WithRiskLevel sets the risk level of the mother
func (m *Mother) WithRiskLevel(riskLevel RiskLevel) *Mother {
	m.RiskLevel = riskLevel
	return m
}

// AddHealthCondition adds a health condition to the mother
func (m *Mother) AddHealthCondition(condition string) *Mother {
	// Check if the condition already exists
	for _, c := range m.HealthConditions {
		if c == condition {
			return m
		}
	}

	// Add the condition
	m.HealthConditions = append(m.HealthConditions, condition)
	return m
}

// GetWeeksPregnant calculates the number of weeks pregnant based on the expected delivery date
func (m *Mother) GetWeeksPregnant(referenceDate time.Time) int {
	// Assuming a standard 40-week pregnancy
	totalPregnancyWeeks := 40

	// Calculate weeks remaining until delivery
	weeksToDelivery := int(m.ExpectedDeliveryDate.Sub(referenceDate).Hours() / 24 / 7)

	// Calculate weeks pregnant
	weeksPregnant := totalPregnancyWeeks - weeksToDelivery

	// Ensure the value is within reasonable range
	if weeksPregnant < 0 {
		return 0
	}
	if weeksPregnant > totalPregnancyWeeks {
		return totalPregnancyWeeks
	}

	return weeksPregnant
}

// IsHighRisk checks if the mother is classified as high risk
func (m *Mother) IsHighRisk() bool {
	return m.RiskLevel == RiskLevelHigh
}
