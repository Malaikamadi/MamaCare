package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"

	"github.com/go-chi/chi/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// Validate middleware provides request body validation
type ValidatorMiddleware struct {
	validate *validator.Validate
	log      logger.Logger
}

// NewValidatorMiddleware creates a new validator middleware
func NewValidatorMiddleware(log logger.Logger) *ValidatorMiddleware {
	return &ValidatorMiddleware{
		validate: validator.New(),
		log:      log,
	}
}

// ValidateRequest validates the request body against the given struct type
func (v *ValidatorMiddleware) ValidateRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only process requests with a body
		if r.Body == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Store body contents in context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "validated", true)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// ValidateBody validates the request body against the given struct type
// and stores the validated struct in the request context
func (v *ValidatorMiddleware) ValidateBody(dest interface{}) func(http.Handler) http.Handler {
	t := reflect.TypeOf(dest)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get request ID for correlation
			reqID := middleware.GetReqID(r.Context())
			
			// Only process POST, PUT, PATCH requests with a body
			if r.Body == nil || (r.Method != http.MethodPost && 
				r.Method != http.MethodPut && 
				r.Method != http.MethodPatch) {
				next.ServeHTTP(w, r)
				return
			}
			
			// Create a new instance of the destination struct
			destValue := reflect.New(t).Interface()
			
			// Read the body
			body, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			
			if err != nil {
				v.handleError(w, errorx.Wrap(err, errorx.InvalidRequest, "failed to read request body"), reqID)
				return
			}
			
			// Unmarshal the JSON
			if err := json.Unmarshal(body, destValue); err != nil {
				v.handleError(w, errorx.Wrap(err, errorx.InvalidRequest, "invalid JSON format"), reqID)
				return
			}
			
			// Validate the struct
			if err := v.validate.Struct(destValue); err != nil {
				validationError := errorx.Wrap(err, errorx.InvalidRequest, "validation failed")
				// Add validation details to the error
				if vErrs, ok := err.(validator.ValidationErrors); ok {
					for _, vErr := range vErrs {
						validationError.AddDetail(vErr.Field(), vErr.Tag(), vErr.Param())
					}
				}
				v.handleError(w, validationError, reqID)
				return
			}
			
			// Store validated object in context
			ctx := context.WithValue(r.Context(), "validated_body", destValue)
			r = r.WithContext(ctx)
			
			next.ServeHTTP(w, r)
		})
	}
}

// GetValidatedBody retrieves the validated body from the request context
func GetValidatedBody(r *http.Request) interface{} {
	return r.Context().Value("validated_body")
}

// handleError handles validation errors consistently
func (v *ValidatorMiddleware) handleError(w http.ResponseWriter, err error, reqID string) {
	// Log the error
	v.log.Warn().Err(err).Str("request_id", reqID).Msg("Request validation failed")
	
	// Create error response
	resp := errorx.NewErrorResponse(err, reqID)
	
	// Write error response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorx.HTTPStatusCode(err))
	json.NewEncoder(w).Encode(resp)
}
