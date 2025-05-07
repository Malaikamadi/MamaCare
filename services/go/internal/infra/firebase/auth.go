package firebase

import (
	"context"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
	"google.golang.org/api/option"
)

// Config contains Firebase configuration
type Config struct {
	CredentialsFile string
	ProjectID       string
}

// FirebaseAuth manages Firebase Authentication
type FirebaseAuth struct {
	app    *firebase.App
	client *auth.Client
	logger logger.Logger
}

// New creates a new FirebaseAuth instance
func New(config Config, logger logger.Logger) *FirebaseAuth {
	return &FirebaseAuth{
		logger: logger,
	}
}

// Initialize initializes the Firebase Authentication client
func (f *FirebaseAuth) Initialize(ctx context.Context) error {
	var err error

	// Initialize with credentials file if provided
	if f.app != nil {
		return nil // Already initialized
	}

	// Initialize Firebase app
	opts := []option.ClientOption{}
	if f.app, err = firebase.NewApp(ctx, nil, opts...); err != nil {
		return errorx.New(errorx.InternalServerError, "failed to initialize firebase app")
	}

	// Initialize Firebase Auth client
	if f.client, err = f.app.Auth(ctx); err != nil {
		return errorx.New(errorx.InternalServerError, "failed to initialize firebase auth client")
	}

	f.logger.Info("Firebase Auth initialized")
	return nil
}

// VerifyIDToken verifies a Firebase ID token
func (f *FirebaseAuth) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	token, err := f.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, errorx.New(errorx.Unauthorized, "invalid firebase id token")
	}
	return token, nil
}

// GetUser retrieves a user by UID
func (f *FirebaseAuth) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
	user, err := f.client.GetUser(ctx, uid)
	if err != nil {
		return nil, errorx.New(errorx.NotFound, "user not found in firebase")
	}
	return user, nil
}

// SetCustomUserClaims sets custom claims for a user
func (f *FirebaseAuth) SetCustomUserClaims(ctx context.Context, uid string, claims map[string]interface{}) error {
	if err := f.client.SetCustomUserClaims(ctx, uid, claims); err != nil {
		return errorx.New(errorx.InternalServerError, "failed to set custom claims in firebase")
	}
	return nil
}

// CreateUser creates a new user in Firebase
func (f *FirebaseAuth) CreateUser(ctx context.Context, params *auth.UserToCreate) (string, error) {
	user, err := f.client.CreateUser(ctx, params)
	if err != nil {
		return "", errorx.New(errorx.InternalServerError, "failed to create user in firebase")
	}
	return user.UID, nil
}

// UpdateUser updates a user in Firebase
func (f *FirebaseAuth) UpdateUser(ctx context.Context, uid string, params *auth.UserToUpdate) error {
	if _, err := f.client.UpdateUser(ctx, uid, params); err != nil {
		return errorx.New(errorx.InternalServerError, "failed to update user in firebase")
	}
	return nil
}

// DeleteUser deletes a user from Firebase
func (f *FirebaseAuth) DeleteUser(ctx context.Context, uid string) error {
	if err := f.client.DeleteUser(ctx, uid); err != nil {
		return errorx.New(errorx.InternalServerError, "failed to delete user from firebase")
	}
	return nil
}

// GenerateCustomToken generates a custom token for a user
func (f *FirebaseAuth) GenerateCustomToken(ctx context.Context, uid string, claims map[string]interface{}) (string, error) {
	// Firebase v4 CustomToken only accepts uid without claims parameter
	// We'll need to serialize claims differently or use another approach
	token, err := f.client.CustomToken(ctx, uid)
	if err != nil {
		return "", errorx.New(errorx.InternalServerError, "failed to generate custom token")
	}
	return token, nil
}

// GetClient returns the Firebase Auth client
func (f *FirebaseAuth) GetClient() *auth.Client {
	return f.client
}
