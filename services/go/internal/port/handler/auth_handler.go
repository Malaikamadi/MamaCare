package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/auth"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/domain/repository"
	"github.com/mamacare/services/internal/port/middleware"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// LoginRequest represents a login request payload
type LoginRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required,e164"`
	VerificationCode string `json:"verification_code" validate:"required"`
}

// LoginResponse represents a login response payload
type LoginResponse struct {
	User *model.User `json:"user"`
	Token string `json:"token"`
	ExpiresAt int64 `json:"expires_at,omitempty"`
}

// TokenResponse represents a JWT token response
type TokenResponse struct {
	Token string `json:"token"`
	ExpiresAt int64 `json:"expires_at,omitempty"`
}

// AuthHandler handles authentication requests
type AuthHandler struct {
	authService    *auth.Service
	userRepository repository.UserRepository
	validate       *validator.Validate
	logger         logger.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(
	authService *auth.Service,
	userRepository repository.UserRepository,
	logger logger.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		userRepository: userRepository,
		validate:       validator.New(),
		logger:         logger,
	}
}

// Login handles user authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, errorx.Wrap(err, errorx.BadRequest, "invalid request payload"))
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.respondWithError(w, errorx.Wrap(err, errorx.BadRequest, "validation failed"))
		return
	}

	// Authenticate user
	user, token, err := h.authService.AuthenticateWithPhone(r.Context(), req.PhoneNumber, req.VerificationCode)
	if err != nil {
		h.respondWithError(w, err)
		return
	}

	// Create and send response
	response := LoginResponse{
		User:  user,
		Token: token,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// RefreshToken handles token refresh requests
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	tokenString := extractTokenFromHeader(r)
	if tokenString == "" {
		h.respondWithError(w, errorx.New(errorx.Unauthorized, "missing authorization token"))
		return
	}

	// Verify token
	authUser, err := h.authService.VerifyToken(r.Context(), tokenString)
	if err != nil {
		h.respondWithError(w, err)
		return
	}

	// Find user in our database
	userID, err := uuid.Parse(authUser.ID)
	if err != nil {
		h.respondWithError(w, errorx.Wrap(err, errorx.InternalServerError, "invalid user ID"))
		return
	}

	user, err := h.userRepository.FindByID(r.Context(), userID)
	if err != nil {
		h.respondWithError(w, errorx.Wrap(err, errorx.Unauthorized, "user not found"))
		return
	}

	// Generate new token
	token, expiresAt, err := h.authService.GenerateHasuraJWT(r.Context(), user)
	if err != nil {
		h.respondWithError(w, err)
		return
	}

	// Create and send response
	response := TokenResponse{
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GenerateHasuraJWT generates a new Hasura JWT for an authenticated user
func (h *AuthHandler) GenerateHasuraJWT(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	authUser, err := middleware.GetAuthUser(r.Context())
	if err != nil {
		h.respondWithError(w, err)
		return
	}

	// Parse user ID
	userID, err := uuid.Parse(authUser.ID)
	if err != nil {
		h.respondWithError(w, errorx.Wrap(err, errorx.InternalServerError, "invalid user ID"))
		return
	}

	// Find user in database
	user, err := h.userRepository.FindByID(r.Context(), userID)
	if err != nil {
		h.respondWithError(w, errorx.Wrap(err, errorx.NotFound, "user not found"))
		return
	}

	// Generate new Hasura JWT
	token, expiresAt, err := h.authService.GenerateHasuraJWT(r.Context(), user)
	if err != nil {
		h.respondWithError(w, err)
		return
	}

	// Create and send response
	response := TokenResponse{
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// Helper methods

// respondWithError sends an error response
func (h *AuthHandler) respondWithError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	var status int
	var response map[string]interface{}

	if appErr, ok := err.(*errorx.Error); ok {
		switch appErr.GetType() {
		case errorx.Unauthorized:
			status = http.StatusUnauthorized
		case errorx.Forbidden:
			status = http.StatusForbidden
		case errorx.NotFound:
			status = http.StatusNotFound
		case errorx.BadRequest:
			status = http.StatusBadRequest
		default:
			status = http.StatusInternalServerError
		}

		response = map[string]interface{}{
			"error": map[string]interface{}{
				"code":    fmt.Sprintf("%d", appErr.GetType()),
				"message": appErr.Error(),
			},
		}
		h.logger.Error("API error", appErr)
	} else {
		status = http.StatusInternalServerError
		response = map[string]interface{}{
			"error": map[string]interface{}{
				"code":    errorx.InternalServerError,
				"message": "Internal server error",
			},
		}
		h.logger.Error("Unexpected error", err)
	}

	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// respondWithJSON sends a JSON response
func (h *AuthHandler) respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// extractTokenFromHeader extracts the JWT token from the Authorization header
func extractTokenFromHeader(r *http.Request) string {
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		return ""
	}

	parts := strings.SplitN(authorization, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}