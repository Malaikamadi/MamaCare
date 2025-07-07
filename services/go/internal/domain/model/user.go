package model

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role of a user in the system
type UserRole string

const (
	// RoleAnonymous represents an unauthenticated user
	RoleAnonymous UserRole = "anonymous"
	// RoleUser represents a basic authenticated user
	RoleUser UserRole = "user"
	// RoleMother represents a pregnant woman enrolled in the system
	RoleMother UserRole = "mother"
	// RoleCHW represents a community health worker
	RoleCHW UserRole = "chw"
	// RoleClinician represents a healthcare professional
	RoleClinician UserRole = "clinician"
	// RoleAdmin represents a system administrator
	RoleAdmin UserRole = "admin"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone_number"`
	Role      UserRole  `json:"role"`
	District  string    `json:"district,omitempty"`
	FacilityID *uuid.UUID `json:"facility_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewUser creates a new user with default values
func NewUser(id uuid.UUID, name, email, phone string, role UserRole) *User {
	now := time.Now()
	return &User{
		ID:        id,
		Name:      name,
		Email:     email,
		Phone:     phone,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// WithDistrict assigns a district to the user
func (u *User) WithDistrict(district string) *User {
	u.District = district
	return u
}

// WithFacility assigns a facility to the user
func (u *User) WithFacility(facilityID uuid.UUID) *User {
	u.FacilityID = &facilityID
	return u
}

// IsHealthcareProvider checks if the user is a healthcare provider (CHW or clinician)
func (u *User) IsHealthcareProvider() bool {
	return u.Role == RoleCHW || u.Role == RoleClinician
}

// HasAdminAccess checks if the user has admin access
func (u *User) HasAdminAccess() bool {
	return u.Role == RoleAdmin
}

// CanAccessMothersData checks if the user can access a mother's data
func (u *User) CanAccessMothersData(motherID uuid.UUID) bool {
	// Admins can access all data
	if u.Role == RoleAdmin {
		return true
	}

	// Mothers can only access their own data
	if u.Role == RoleMother {
		return u.ID == motherID
	}

	// Healthcare providers can access data for mothers in their district
	// Note: Further check against assigned mothers would be implemented in repository layer
	return u.IsHealthcareProvider()
}
