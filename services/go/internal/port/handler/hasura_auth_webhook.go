package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/auth"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/pkg/logger"
)

// HasuraAuthWebhook handles authentication webhooks from Hasura
type HasuraAuthWebhook struct {
	authService *auth.Service
	logger      logger.Logger
}

// HasuraAuthResponse represents the response format expected by Hasura
type HasuraAuthResponse struct {
	// X-Hasura-Role is the user's role for Hasura RBAC
	Role string `json:"X-Hasura-Role,omitempty"`

	// X-Hasura-User-Id is the user's ID
	UserID string `json:"X-Hasura-User-Id,omitempty"`

	// X-Hasura-* are any additional custom claims
	Claims json.RawMessage `json:"claims,omitempty"`
}

// NewHasuraAuthWebhook creates a new Hasura authentication webhook handler
func NewHasuraAuthWebhook(authService *auth.Service, logger logger.Logger) *HasuraAuthWebhook {
	return &HasuraAuthWebhook{
		authService: authService,
		logger:      logger,
	}
}

// Authenticate handles Hasura authentication requests
func (h *HasuraAuthWebhook) Authenticate(w http.ResponseWriter, r *http.Request) {
	// Get authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// No authorization header, return anonymous role
		h.respondWithJSON(w, http.StatusOK, HasuraAuthResponse{
			Role: "anonymous",
		})
		return
	}

	// Extract token
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		h.respondWithError(w, http.StatusUnauthorized, "invalid authorization format")
		return
	}
	token := tokenParts[1]

	// Verify token
	authUser, err := h.authService.VerifyToken(r.Context(), token)
	if err != nil {
		h.logger.Error("Failed to verify token", err)
		h.respondWithError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	// Get additional Hasura-specific claims
	claims, err := h.authService.GenerateHasuraClaims(r.Context(), &model.User{
		ID:    uuid.MustParse(authUser.ID),
		Role:  authUser.Role,
		Email: authUser.Email,
	})
	if err != nil {
		h.logger.Error("Failed to generate Hasura claims", err)
		h.respondWithError(w, http.StatusInternalServerError, "failed to generate claims")
		return
	}

	// Create response
	response := HasuraAuthResponse{
		Role:   string(authUser.Role),
		UserID: authUser.ID,
		Claims: claims,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// Helper methods

// respondWithJSON sends a JSON response
func (h *HasuraAuthWebhook) respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", err)
	}
}

// respondWithError sends an error response
func (h *HasuraAuthWebhook) respondWithError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	response := map[string]interface{}{
		"error": message,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode error response", err)
	}
}