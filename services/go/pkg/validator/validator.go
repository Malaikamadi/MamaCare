package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mamacare/services/pkg/errorx"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

// Validator encapsulates validator functionality
type Validator struct {
	validate *validator.Validate
}

// New creates a new validator
func New() *Validator {
	v := validator.New()

	// Register function to get json tag names
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{
		validate: v,
	}
}

// RegisterCustomValidation registers a custom validation function
func (v *Validator) RegisterCustomValidation(tag string, fn validator.Func) error {
	return v.validate.RegisterValidation(tag, fn)
}

// Validate validates a struct and returns a list of validation errors
func (v *Validator) Validate(i interface{}) error {
	err := v.validate.Struct(i)
	if err == nil {
		return nil
	}

	var errors []ValidationError
	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, ValidationError{
			Field: err.Field(),
			Tag:   err.Tag(),
			Value: fmt.Sprintf("%v", err.Value()),
		})
	}

	// Return a custom error with BadRequest type and context
	customErr := errorx.New(errorx.BadRequest, "Invalid input data")
	for _, e := range errors {
		customErr.WithContext(fmt.Sprintf("validation_%s", e.Field), fmt.Sprintf("%s: %s", e.Tag, e.Value))
	}
	
	return customErr
}

// ValidateVar validates a variable
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	err := v.validate.Var(field, tag)
	if err == nil {
		return nil
	}

	// Return a custom error
	return errorx.Newf(errorx.BadRequest, "Validation failed for field with tag '%s'", tag)
}
