package auth

import (
	"context"
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v4" // Use v4 instead of v5
	"github.com/google/uuid"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/domain/repository"
	"github.com/mamacare/services/internal/infra/firebase"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// Config contains authentication service configuration
type Config struct {
	JWTSecret         string        // Secret for signing JWTs
	JWTExpiryDuration time.Duration // JWT token expiry duration
	HasuraNamespace   string        // Namespace for Hasura JWT claims
}

// Service handles authentication operations
type Service struct {
	config         Config
	firebaseAuth   *firebase.FirebaseAuth
	userRepository repository.UserRepository
	logger         logger.Logger
}

// NewService creates a new authentication service
func NewService(
	config Config,
	firebaseAuth *firebase.FirebaseAuth,
	userRepository repository.UserRepository,
	logger logger.Logger,
) *Service {
	return &Service{
		config:         config,
		firebaseAuth:   firebaseAuth,
		userRepository: userRepository,
		logger:         logger,
	}
}

// VerifyToken verifies a Firebase ID token and returns user information
func (s *Service) VerifyToken(ctx context.Context, token string) (*model.AuthUser, error) {
	// Check if Firebase auth is initialized
	if s.firebaseAuth == nil {
		return nil, errorx.New(errorx.InternalServerError, "Firebase authentication not initialized")
	}
	
	// Verify with Firebase
	decodedToken, err := s.firebaseAuth.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.Unauthorized, "invalid token")
	}

	// Extract user ID from token
	userID, err := uuid.Parse(decodedToken.UID)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "invalid user ID in token")
	}

	// Find user in our database
	user, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.Unauthorized, "user not found")
	}

	// Create auth user
	authUser := &model.AuthUser{
		ID:     user.ID.String(),
		Email:  user.Email,
		Role:   user.Role,
		Claims: decodedToken.Claims,
	}

	return authUser, nil
}

// GenerateHasuraJWT generates a JWT token with Hasura-specific claims
func (s *Service) GenerateHasuraJWT(ctx context.Context, user *model.User) (string, time.Time, error) {
	// Calculate expiry time
	expiresAt := time.Now().Add(s.config.JWTExpiryDuration)

	// Prepare claims for Hasura
	hasuraClaims := map[string]interface{}{
		"x-hasura-allowed-roles": []string{string(user.Role)},
		"x-hasura-default-role":  string(user.Role),
		"x-hasura-user-id":       user.ID.String(),
	}

	// Add additional claims based on role
	switch user.Role {
	case "mother": // Use string literals instead of model constants until they're defined
		// When user is a mother, the mother ID is the same as user ID
		hasuraClaims["x-hasura-mother-id"] = user.ID.String()
	case "chw":
		hasuraClaims["x-hasura-chw-id"] = user.ID.String()
	case "clinician":
		hasuraClaims["x-hasura-clinician-id"] = user.ID.String()
	}

	// Create the JWT token
	claims := jwt.MapClaims{
		"sub":                    user.ID.String(),
		"iat":                    time.Now().Unix(),
		"exp":                    expiresAt.Unix(),
		"https://hasura.io/jwt/claims": hasuraClaims,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Sign the token
	signedToken, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", time.Time{}, errorx.Wrap(err, errorx.InternalServerError, "failed to sign token")
	}

	return signedToken, expiresAt, nil
}

// GenerateHasuraClaims generates Hasura JWT claims for a user
func (s *Service) GenerateHasuraClaims(ctx context.Context, user *model.User) (json.RawMessage, error) {
	// Prepare claims for Hasura
	hasuraClaims := map[string]interface{}{
		"x-hasura-allowed-roles": []string{string(user.Role)},
		"x-hasura-default-role":  string(user.Role),
		"x-hasura-user-id":       user.ID.String(),
	}

	// Add additional claims based on role
	switch user.Role {
	case "mother":
		// When user is a mother, the mother ID is the same as user ID
		hasuraClaims["x-hasura-mother-id"] = user.ID.String()
	case "chw":
		hasuraClaims["x-hasura-chw-id"] = user.ID.String()
	case "clinician":
		hasuraClaims["x-hasura-clinician-id"] = user.ID.String()
	}

	// Convert to JSON
	claimsJSON, err := json.Marshal(hasuraClaims)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to marshal Hasura claims")
	}

	return claimsJSON, nil
}

// AuthenticateWithPhone authenticates a user using phone verification
func (s *Service) AuthenticateWithPhone(ctx context.Context, phoneNumber string, verificationCode string) (*model.User, string, error) {
	// In a real implementation, this would verify with Firebase phone auth
	// For now, just look up the user by phone number
	user, err := s.userRepository.FindByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		return nil, "", errorx.Wrap(err, errorx.Unauthorized, "invalid phone number or verification code")
	}

	// Generate custom token using Firebase (with defensive error handling)
	firebaseClient := s.firebaseAuth.GetClient()
	if firebaseClient == nil {
		return nil, "", errorx.New(errorx.InternalServerError, "Firebase client not initialized")
	}
	
	customToken, err := firebaseClient.CustomToken(ctx, user.ID.String())
	if err != nil {
		return nil, "", errorx.Wrap(err, errorx.InternalServerError, "failed to generate token")
	}

	return user, customToken, nil
}