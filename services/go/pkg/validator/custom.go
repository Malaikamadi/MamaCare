package validator

import (
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
)

// RegisterCustomValidators registers all custom validators
func (v *Validator) RegisterCustomValidators() {
	// Phone number validation (basic international format)
	_ = v.RegisterCustomValidation("phone", isValidPhone)
	
	// Future date validation (for expected delivery dates)
	_ = v.RegisterCustomValidation("future_date", isFutureDate)
	
	// Geolocation validation (for latitude and longitude)
	_ = v.RegisterCustomValidation("latitude", isValidLatitude)
	_ = v.RegisterCustomValidation("longitude", isValidLongitude)
	
	// Time in 24-hour format validation (for operating hours)
	_ = v.RegisterCustomValidation("time24h", isValidTime24h)
	
	// UUID validation
	_ = v.RegisterCustomValidation("uuid", isValidUUID)
	
	// Blood pressure validation
	_ = v.RegisterCustomValidation("systolic", isValidSystolic)
	_ = v.RegisterCustomValidation("diastolic", isValidDiastolic)
}

// isValidPhone validates a phone number
func isValidPhone(fl validator.FieldLevel) bool {
	// Basic international phone number format (e.g., +1234567890)
	// You can make this more complex based on your requirements
	re := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	return re.MatchString(fl.Field().String())
}

// isFutureDate validates that a date is in the future
func isFutureDate(fl validator.FieldLevel) bool {
	date, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}
	
	// Allow dates from today onwards
	return date.After(time.Now().AddDate(0, 0, -1))
}

// isValidLatitude validates a latitude value
func isValidLatitude(fl validator.FieldLevel) bool {
	lat := fl.Field().Float()
	return lat >= -90 && lat <= 90
}

// isValidLongitude validates a longitude value
func isValidLongitude(fl validator.FieldLevel) bool {
	lng := fl.Field().Float()
	return lng >= -180 && lng <= 180
}

// isValidTime24h validates a time string in 24-hour format (e.g., "14:30")
func isValidTime24h(fl validator.FieldLevel) bool {
	timeStr := fl.Field().String()
	_, err := time.Parse("15:04", timeStr)
	return err == nil
}

// isValidUUID validates a UUID string
func isValidUUID(fl validator.FieldLevel) bool {
	uuidStr := fl.Field().String()
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return re.MatchString(uuidStr)
}

// isValidSystolic validates a systolic blood pressure value
func isValidSystolic(fl validator.FieldLevel) bool {
	value := fl.Field().Int()
	// Normal range for systolic: 90-180 mmHg (wider range for pregnancy)
	return value >= 80 && value <= 200
}

// isValidDiastolic validates a diastolic blood pressure value
func isValidDiastolic(fl validator.FieldLevel) bool {
	value := fl.Field().Int()
	// Normal range for diastolic: 60-110 mmHg (wider range for pregnancy)
	return value >= 40 && value <= 120
}
