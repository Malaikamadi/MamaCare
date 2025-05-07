package model

import (
	"time"

	"github.com/google/uuid"
)

// VitalSigns represents a collection of vital health measurements
type VitalSigns struct {
	BloodPressure   *BloodPressure `json:"blood_pressure,omitempty"`
	FetalHeartRate  *float64       `json:"fetal_heart_rate,omitempty"`
	FetalMovement   *float64       `json:"fetal_movement,omitempty"`
	BloodSugar      *float64       `json:"blood_sugar,omitempty"`
	HemoglobinLevel *float64       `json:"hemoglobin_level,omitempty"`
	IronLevel       *float64       `json:"iron_level,omitempty"`
	Weight          *float64       `json:"weight,omitempty"`
}

// BloodPressure represents a blood pressure measurement
type BloodPressure struct {
	Systolic  float64 `json:"systolic"`
	Diastolic float64 `json:"diastolic"`
}

// WeightRecord represents a weight measurement with a timestamp
type WeightRecord struct {
	Date   time.Time `json:"date"`
	Weight float64   `json:"weight"`
}

// ContractionReading represents a contraction reading
type ContractionReading struct {
	Duration      int `json:"duration"`       // seconds
	Interval      int `json:"interval"`       // seconds
	Intensity     int `json:"intensity"`      // 1-10 scale
	FrequencyHour int `json:"frequency_hour"` // per hour
}

// HealthMetric represents a health measurement record
type HealthMetric struct {
	ID            uuid.UUID  `json:"id"`
	MotherID      uuid.UUID  `json:"mother_id"`
	VisitID       *uuid.UUID `json:"visit_id,omitempty"`
	RecordedByID  *uuid.UUID `json:"recorded_by_id,omitempty"`
	RecordedAt    time.Time  `json:"recorded_at"`
	VitalSigns    VitalSigns `json:"vital_signs"`
	Contractions  *ContractionReading `json:"contractions,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	IsAbnormal    bool       `json:"is_abnormal"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// NewHealthMetric creates a new health metric record
func NewHealthMetric(id, motherID uuid.UUID) *HealthMetric {
	now := time.Now()
	return &HealthMetric{
		ID:         id,
		MotherID:   motherID,
		RecordedAt: now,
		VitalSigns: VitalSigns{},
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// WithRecordedBy sets who recorded the health metrics
func (h *HealthMetric) WithRecordedBy(recordedByID uuid.UUID) *HealthMetric {
	h.RecordedByID = &recordedByID
	return h
}

// WithVisit associates a visit with the health metric
func (h *HealthMetric) WithVisit(visitID uuid.UUID) *HealthMetric {
	h.VisitID = &visitID
	return h
}

// WithBloodPressure adds blood pressure measurements
func (h *HealthMetric) WithBloodPressure(systolic, diastolic float64) *HealthMetric {
	h.VitalSigns.BloodPressure = &BloodPressure{
		Systolic:  systolic,
		Diastolic: diastolic,
	}
	
	// Check if blood pressure is abnormal
	if systolic >= 140 || diastolic >= 90 || systolic < 90 || diastolic < 60 {
		h.IsAbnormal = true
	}
	
	return h
}

// WithFetalHeartRate adds fetal heart rate measurement
func (h *HealthMetric) WithFetalHeartRate(rate float64) *HealthMetric {
	h.VitalSigns.FetalHeartRate = &rate
	
	// Check if fetal heart rate is abnormal
	if rate < 110 || rate > 160 {
		h.IsAbnormal = true
	}
	
	return h
}

// WithFetalMovement adds fetal movement measurement
func (h *HealthMetric) WithFetalMovement(movement float64) *HealthMetric {
	h.VitalSigns.FetalMovement = &movement
	
	// Check if fetal movement is abnormal (less than 10 movements in counting session)
	if movement < 10 {
		h.IsAbnormal = true
	}
	
	return h
}

// WithBloodSugar adds blood sugar measurement
func (h *HealthMetric) WithBloodSugar(bloodSugar float64) *HealthMetric {
	h.VitalSigns.BloodSugar = &bloodSugar
	
	// Check if blood sugar is abnormal (above 95 mg/dL fasting)
	if bloodSugar > 95 {
		h.IsAbnormal = true
	}
	
	return h
}

// WithHemoglobinLevel adds hemoglobin level measurement
func (h *HealthMetric) WithHemoglobinLevel(hemoglobinLevel float64) *HealthMetric {
	h.VitalSigns.HemoglobinLevel = &hemoglobinLevel
	
	// Check if hemoglobin is abnormal (below 11 g/dL indicates anemia in pregnancy)
	if hemoglobinLevel < 11 {
		h.IsAbnormal = true
	}
	
	return h
}

// WithIronLevel adds iron level measurement
func (h *HealthMetric) WithIronLevel(ironLevel float64) *HealthMetric {
	h.VitalSigns.IronLevel = &ironLevel
	return h
}

// WithWeight adds weight measurement
func (h *HealthMetric) WithWeight(weight float64) *HealthMetric {
	h.VitalSigns.Weight = &weight
	
	// Abnormal if weight is less than 45kg (severe underweight for most adult women)
	if weight < 45.0 {
		h.IsAbnormal = true
	}
	
	return h
}

// WithContractions adds contraction measurements
func (h *HealthMetric) WithContractions(duration, interval, intensity, frequency int) *HealthMetric {
	h.Contractions = &ContractionReading{
		Duration:      duration,
		Interval:      interval,
		Intensity:     intensity,
		FrequencyHour: frequency,
	}
	
	// Check if contractions are abnormal
	if interval < 300 || duration > 90 || frequency > 5 {
		h.IsAbnormal = true
	}
	
	return h
}

// WithNotes adds notes to the health metric
func (h *HealthMetric) WithNotes(notes string) *HealthMetric {
	h.Notes = notes
	return h
}

// SetAbnormal explicitly sets the abnormal flag
func (h *HealthMetric) SetAbnormal(isAbnormal bool) *HealthMetric {
	h.IsAbnormal = isAbnormal
	return h
}

// IsBloodPressureNormal checks if blood pressure is within normal range
func (h *HealthMetric) IsBloodPressureNormal() bool {
	if h.VitalSigns.BloodPressure == nil {
		return true
	}
	
	// Normal range for pregnant women
	// Systolic: 90-140 mmHg
	// Diastolic: 60-90 mmHg
	return h.VitalSigns.BloodPressure.Systolic < 140 && 
	       h.VitalSigns.BloodPressure.Systolic >= 90 && 
	       h.VitalSigns.BloodPressure.Diastolic < 90 && 
	       h.VitalSigns.BloodPressure.Diastolic >= 60
}

// IsFetalHeartRateNormal checks if fetal heart rate is within normal range
func (h *HealthMetric) IsFetalHeartRateNormal() bool {
	if h.VitalSigns.FetalHeartRate == nil {
		return true
	}
	
	// Normal range for fetal heart rate: 110-160 BPM
	rate := *h.VitalSigns.FetalHeartRate
	return rate >= 110 && rate <= 160
}
