package validator

import (
	"time"

	"github.com/google/uuid"
)

// DTOs for authentication
type (
	// UserRegistrationDTO represents user registration data
	UserRegistrationDTO struct {
		Name     string `json:"name" validate:"required,min=3,max=100"`
		Email    string `json:"email" validate:"required,email"`
		Phone    string `json:"phone" validate:"required,phone"`
		Password string `json:"password" validate:"required,min=8,max=100"`
		Role     string `json:"role" validate:"required,oneof=user mother chw clinician admin"`
		District string `json:"district" validate:"omitempty,max=100"`
	}

	// UserLoginDTO represents user login data
	UserLoginDTO struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	// UserUpdateDTO represents user update data
	UserUpdateDTO struct {
		Name     string `json:"name" validate:"omitempty,min=3,max=100"`
		Phone    string `json:"phone" validate:"omitempty,phone"`
		District string `json:"district" validate:"omitempty,max=100"`
	}
)

// DTOs for mother/pregnancy data
type (
	// MotherRegistrationDTO represents mother registration data
	MotherRegistrationDTO struct {
		UserID                uuid.UUID `json:"user_id" validate:"required,uuid"`
		ExpectedDeliveryDate  time.Time `json:"expected_delivery_date" validate:"required,future_date"`
		BloodType             string    `json:"blood_type" validate:"required,oneof=A+ A- B+ B- AB+ AB- O+ O- unknown"`
		HealthConditions      []string  `json:"health_conditions" validate:"omitempty,dive,max=100"`
		PreviousPregnancies   int       `json:"previous_pregnancies" validate:"min=0,max=20"`
		PreviousDeliveries    int       `json:"previous_deliveries" validate:"min=0,max=20"`
		PreviousCaesareans    int       `json:"previous_caesareans" validate:"min=0,max=20"`
		PreviousComplications []string  `json:"previous_complications" validate:"omitempty,dive,max=100"`
	}

	// MotherUpdateDTO represents mother update data
	MotherUpdateDTO struct {
		ExpectedDeliveryDate  *time.Time `json:"expected_delivery_date" validate:"omitempty,future_date"`
		BloodType             *string    `json:"blood_type" validate:"omitempty,oneof=A+ A- B+ B- AB+ AB- O+ O- unknown"`
		HealthConditions      []string   `json:"health_conditions" validate:"omitempty,dive,max=100"`
		PreviousPregnancies   *int       `json:"previous_pregnancies" validate:"omitempty,min=0,max=20"`
		PreviousDeliveries    *int       `json:"previous_deliveries" validate:"omitempty,min=0,max=20"`
		PreviousCaesareans    *int       `json:"previous_caesareans" validate:"omitempty,min=0,max=20"`
		PreviousComplications []string   `json:"previous_complications" validate:"omitempty,dive,max=100"`
	}
)

// DTOs for healthcare facility data
type (
	// FacilityRegistrationDTO represents facility registration data
	FacilityRegistrationDTO struct {
		Name          string   `json:"name" validate:"required,min=3,max=100"`
		Address       string   `json:"address" validate:"required,min=5,max=200"`
		District      string   `json:"district" validate:"required,max=100"`
		Latitude      float64  `json:"latitude" validate:"required,latitude"`
		Longitude     float64  `json:"longitude" validate:"required,longitude"`
		FacilityType  string   `json:"facility_type" validate:"required,oneof=hospital clinic health_center health_post"`
		Capacity      int      `json:"capacity" validate:"required,min=1,max=10000"`
		OperatingHours string   `json:"operating_hours" validate:"required,json"`
		Services      []string `json:"services" validate:"omitempty,dive,max=100"`
	}

	// FacilityUpdateDTO represents facility update data
	FacilityUpdateDTO struct {
		Name          *string   `json:"name" validate:"omitempty,min=3,max=100"`
		Address       *string   `json:"address" validate:"omitempty,min=5,max=200"`
		District      *string   `json:"district" validate:"omitempty,max=100"`
		Latitude      *float64  `json:"latitude" validate:"omitempty,latitude"`
		Longitude     *float64  `json:"longitude" validate:"omitempty,longitude"`
		FacilityType  *string   `json:"facility_type" validate:"omitempty,oneof=hospital clinic health_center health_post"`
		Capacity      *int      `json:"capacity" validate:"omitempty,min=1,max=10000"`
		OperatingHours *string   `json:"operating_hours" validate:"omitempty,json"`
		Services      []string  `json:"services" validate:"omitempty,dive,max=100"`
	}
)

// DTOs for health metrics data
type (
	// BloodPressureDTO represents blood pressure data
	BloodPressureDTO struct {
		Systolic  int    `json:"systolic" validate:"required,systolic"`
		Diastolic int    `json:"diastolic" validate:"required,diastolic"`
		Notes     string `json:"notes" validate:"omitempty,max=500"`
	}

	// WeightDTO represents weight data
	WeightDTO struct {
		Value float64 `json:"value" validate:"required,min=20,max=250"`
		Unit  string  `json:"unit" validate:"required,oneof=kg lb"`
		Notes string  `json:"notes" validate:"omitempty,max=500"`
	}

	// BloodSugarDTO represents blood sugar data
	BloodSugarDTO struct {
		Value float64 `json:"value" validate:"required,min=0,max=500"`
		Unit  string  `json:"unit" validate:"required,oneof=mg/dL mmol/L"`
		Notes string  `json:"notes" validate:"omitempty,max=500"`
	}

	// FetalMovementDTO represents fetal movement data
	FetalMovementDTO struct {
		Count    int    `json:"count" validate:"required,min=0,max=100"`
		Duration int    `json:"duration" validate:"required,min=1,max=480"` // minutes
		Notes    string `json:"notes" validate:"omitempty,max=500"`
	}

	// ContractionDTO represents contraction data
	ContractionDTO struct {
		Duration      int    `json:"duration" validate:"required,min=1,max=300"` // seconds
		Interval      int    `json:"interval" validate:"required,min=0,max=3600"` // seconds
		Intensity     int    `json:"intensity" validate:"required,min=1,max=10"`
		FrequencyHour int    `json:"frequency_hour" validate:"required,min=0,max=60"`
		Notes         string `json:"notes" validate:"omitempty,max=500"`
	}

	// HealthMetricDTO is a generic container for health metric data
	HealthMetricDTO struct {
		MotherID    uuid.UUID     `json:"mother_id" validate:"required,uuid"`
		RecordedBy  uuid.UUID     `json:"recorded_by" validate:"required,uuid"`
		VisitID     *uuid.UUID    `json:"visit_id" validate:"omitempty,uuid"`
		MetricType  string        `json:"metric_type" validate:"required,oneof=blood_pressure weight blood_sugar fetal_movement fetal_heart_rate contractions"`
		RecordedAt  time.Time     `json:"recorded_at" validate:"required"`
		NumericValue *float64      `json:"numeric_value" validate:"omitempty"`
		BloodPressure *BloodPressureDTO `json:"blood_pressure" validate:"omitempty"`
		Contractions *ContractionDTO   `json:"contractions" validate:"omitempty"`
		Notes       string        `json:"notes" validate:"omitempty,max=500"`
	}
)

// DTOs for visit/appointment data
type (
	// VisitScheduleDTO represents appointment scheduling data
	VisitScheduleDTO struct {
		MotherID      uuid.UUID  `json:"mother_id" validate:"required,uuid"`
		FacilityID    uuid.UUID  `json:"facility_id" validate:"required,uuid"`
		CHWID         *uuid.UUID `json:"chw_id" validate:"omitempty,uuid"`
		ClinicianID   *uuid.UUID `json:"clinician_id" validate:"omitempty,uuid"`
		ScheduledTime time.Time  `json:"scheduled_time" validate:"required,future_date"`
		VisitType     string     `json:"visit_type" validate:"required,oneof=routine emergency follow_up"`
		Notes         string     `json:"notes" validate:"omitempty,max=500"`
	}

	// VisitUpdateDTO represents appointment update data
	VisitUpdateDTO struct {
		FacilityID    *uuid.UUID `json:"facility_id" validate:"omitempty,uuid"`
		CHWID         *uuid.UUID `json:"chw_id" validate:"omitempty,uuid"`
		ClinicianID   *uuid.UUID `json:"clinician_id" validate:"omitempty,uuid"`
		ScheduledTime *time.Time `json:"scheduled_time" validate:"omitempty,future_date"`
		VisitType     *string    `json:"visit_type" validate:"omitempty,oneof=routine emergency follow_up"`
		Notes         *string    `json:"notes" validate:"omitempty,max=500"`
	}

	// VisitCompletionDTO represents visit completion data
	VisitCompletionDTO struct {
		Notes string `json:"notes" validate:"required,max=1000"`
	}
)

// DTOs for emergency (SOS) data
type (
	// SOSRequestDTO represents an emergency request
	SOSRequestDTO struct {
		MotherID    uuid.UUID `json:"mother_id" validate:"required,uuid"`
		ReportedBy  uuid.UUID `json:"reported_by" validate:"required,uuid"`
		Latitude    float64   `json:"latitude" validate:"required,latitude"`
		Longitude   float64   `json:"longitude" validate:"required,longitude"`
		Nature      string    `json:"nature" validate:"required,oneof=labor bleeding accident other"`
		Description string    `json:"description" validate:"omitempty,max=1000"`
	}

	// SOSDispatchDTO represents ambulance dispatch data
	SOSDispatchDTO struct {
		AmbulanceID uuid.UUID `json:"ambulance_id" validate:"required,uuid"`
		ETA         time.Time `json:"eta" validate:"required,future_date"`
	}

	// SOSResolveDTO represents SOS resolution data
	SOSResolveDTO struct {
		FacilityID uuid.UUID `json:"facility_id" validate:"required,uuid"`
		Notes      string    `json:"notes" validate:"omitempty,max=1000"`
	}
)
