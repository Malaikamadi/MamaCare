package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mamacare/services/pkg/errorx"
)

// Validator provides request validation functionality
type Validator struct {
	validate *validator.Validate
}

// ValidationError represents a validation error for a specific field
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	v := validator.New()
	
	// Register custom name resolver that uses the "json" tag
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	
	return &Validator{validate: v}
}

// Validate validates a struct and returns a detailed error with field information
func (v *Validator) Validate(data interface{}) error {
	err := v.validate.Struct(data)
	if err == nil {
		return nil
	}
	
	// Extract validation errors
	validationErrors := make([]ValidationError, 0)
	
	if errors, ok := err.(validator.ValidationErrors); ok {
		for _, err := range errors {
			field := err.Field()
			tag := err.Tag()
			value := fmt.Sprintf("%v", err.Value())
			
			message := fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", field, tag)
			
			// Add more specific messages based on validation tag
			switch tag {
			case "required":
				message = fmt.Sprintf("'%s' is required", field)
			case "email":
				message = fmt.Sprintf("'%s' must be a valid email address", field)
			case "min":
				param := err.Param()
				message = fmt.Sprintf("'%s' must be at least %s characters long", field, param)
			case "max":
				param := err.Param()
				message = fmt.Sprintf("'%s' must be at most %s characters long", field, param)
			}
			
			validationErrors = append(validationErrors, ValidationError{
				Field:   field,
				Tag:     tag,
				Value:   value,
				Message: message,
			})
		}
	}
	
	// Create a structured error
	validationErr := errorx.New(errorx.BadRequest, "Validation failed")
	
	// Add validation details to error
	for _, ve := range validationErrors {
		validationErr = validationErr.AddContext(ve.Field, ve.Tag, ve.Message)
	}
	
	return validationErr
}

// ValidateRequest validates a request object and returns a detailed validation error
func (v *Validator) ValidateRequest(data interface{}) error {
	return v.Validate(data)
}

// Register registers custom validators
func (v *Validator) Register() error {
	// Example: register custom validator for phone numbers
	err := v.validate.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		// Simple check: at least 10 digits
		value := fl.Field().String()
		cleanPhone := strings.ReplaceAll(strings.ReplaceAll(value, " ", ""), "-", "")
		return len(cleanPhone) >= 10
	})
	
	if err != nil {
		return errors.New("failed to register custom validator: " + err.Error())
	}
	
	return nil
}
