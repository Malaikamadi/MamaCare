package model

// AuthUser represents an authenticated user with their permissions
type AuthUser struct {
	// ID is the unique identifier for the user
	ID string `json:"id"`

	// Email is the user's email address
	Email string `json:"email"`

	// Role represents the user's role in the system
	Role UserRole `json:"role"`

	// Claims contains additional Firebase authentication claims
	Claims map[string]interface{} `json:"claims"`
}

// HasRole checks if the authenticated user has the specified role
func (a *AuthUser) HasRole(role UserRole) bool {
	return a.Role == role
}

// HasAnyRole checks if the authenticated user has any of the specified roles
func (a *AuthUser) HasAnyRole(roles []UserRole) bool {
	for _, role := range roles {
		if a.Role == role {
			return true
		}
	}
	return false
}