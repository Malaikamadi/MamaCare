package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/infra/firebase"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// AuthMiddleware handles authentication and authorization
type AuthMiddleware struct {
	firebaseAuth *firebase.FirebaseAuth
	logger       logger.Logger
}

// AuthUserCtxKey is the context key for the authenticated user
type AuthUserCtxKey struct{}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(firebaseAuth *firebase.FirebaseAuth, logger logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		firebaseAuth: firebaseAuth,
		logger:       logger,
	}
}

// Authenticate verifies the JWT token in the request
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		tokenString := extractTokenFromHeader(r)
		if tokenString == "" {
			respondWithError(w, errorx.New(errorx.Unauthorized, "missing or invalid authorization token"))
			return
		}

		// Verify token with Firebase
		decodedToken, err := m.firebaseAuth.VerifyIDToken(r.Context(), tokenString)
		if err != nil {
			m.logger.Error("Failed to verify ID token", err)
			respondWithError(w, errorx.Wrap(err, errorx.Unauthorized, "invalid token"))
			return
		}

		// Extract user ID from token
		userID, err := extractUserID(decodedToken.UID)
		if err != nil {
			m.logger.Error("Invalid user ID in token", err)
			respondWithError(w, errorx.Wrap(err, errorx.InternalServerError, "invalid user ID"))
			return
		}

		// Extract user role from token claims
		role, err := extractUserRole(decodedToken.Claims)
		if err != nil {
			m.logger.Error("Failed to extract user role", err)
			respondWithError(w, errorx.Wrap(err, errorx.InternalServerError, "invalid user role"))
			return
		}

		// Create auth user
		authUser := &model.AuthUser{
			ID:     userID.String(),
			Role:   role,
			Claims: decodedToken.Claims,
		}

		// Add auth user to context
		ctx := context.WithValue(r.Context(), AuthUserCtxKey{}, authUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole ensures the user has one of the required roles
func (m *AuthMiddleware) RequireRole(allowedRoles []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get auth user from context
			authUser, ok := r.Context().Value(AuthUserCtxKey{}).(*model.AuthUser)
			if !ok {
				respondWithError(w, errorx.New(errorx.Unauthorized, "user not authenticated"))
				return
			}

			// Check if user has an allowed role
			hasAllowedRole := false
			for _, role := range allowedRoles {
				if string(authUser.Role) == role {
					hasAllowedRole = true
					break
				}
			}

			if !hasAllowedRole {
				respondWithError(w, errorx.New(errorx.Forbidden, "insufficient permissions"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetAuthUser retrieves the authenticated user from context
func GetAuthUser(ctx context.Context) (*model.AuthUser, error) {
	authUser, ok := ctx.Value(AuthUserCtxKey{}).(*model.AuthUser)
	if !ok || authUser == nil {
		return nil, errorx.New(errorx.Unauthorized, "user not authenticated")
	}
	return authUser, nil
}

// Helper functions

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

// extractUserID extracts and validates the user ID from the token
func extractUserID(uid string) (uuid.UUID, error) {
	userID, err := uuid.Parse(uid)
	if err != nil {
		return uuid.Nil, errorx.Wrap(err, errorx.InternalServerError, "invalid user ID format")
	}
	return userID, nil
}

// extractUserRole extracts the user role from token claims
func extractUserRole(claims map[string]interface{}) (model.UserRole, error) {
	// Try to get role from custom claims
	roleValue, ok := claims["role"]
	if !ok {
		// Default to GUEST if no role claim is present
		return model.UserRoleGuest, nil
	}

	roleStr, ok := roleValue.(string)
	if !ok {
		return "", errorx.New(errorx.InternalServerError, "invalid role format in claims")
	}

	// Validate role
	role := model.UserRole(roleStr)
	switch role {
	case model.UserRoleAdmin, model.UserRoleCHW, model.UserRoleClinician, model.UserRoleMother, model.UserRoleGuest:
		return role, nil
	default:
		return "", errorx.New(errorx.InternalServerError, "unknown user role")
	}
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	var status int
	var response map[string]interface{}

	if appErr, ok := err.(*errorx.Error); ok {
		switch appErr.Code() {
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
				"code":    appErr.Code(),
				"message": appErr.Error(),
			},
		}
	} else {
		status = http.StatusInternalServerError
		response = map[string]interface{}{
			"error": map[string]interface{}{
				"code":    errorx.InternalServerError,
				"message": "Internal server error",
			},
		}
	}

	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}