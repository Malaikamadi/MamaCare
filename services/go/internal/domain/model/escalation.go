package model

import (
	"time"

	"github.com/google/uuid"
)

// EscalationLevel represents the severity level of an escalation
type EscalationLevel int

const (
	// EscalationLevelLow represents a low-priority escalation
	EscalationLevelLow EscalationLevel = 1
	// EscalationLevelMedium represents a medium-priority escalation
	EscalationLevelMedium EscalationLevel = 2
	// EscalationLevelHigh represents a high-priority escalation
	EscalationLevelHigh EscalationLevel = 3
	// EscalationLevelCritical represents a critical-priority escalation
	EscalationLevelCritical EscalationLevel = 4
)

// EscalationTier represents a tier in the escalation path
type EscalationTier struct {
	ID           uuid.UUID      `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Level        EscalationLevel `json:"level"`
	ResponseTime int            `json:"response_time_minutes"` // Expected response time in minutes
	Contacts     []Contact      `json:"contacts"`
	NextTierID   *uuid.UUID     `json:"next_tier_id,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// Contact represents a contact person in an escalation tier
type Contact struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	Phone        string    `json:"phone"`
	Email        string    `json:"email,omitempty"`
	FacilityID   *uuid.UUID `json:"facility_id,omitempty"`
	IsEmergency  bool      `json:"is_emergency"`
	IsEscalation bool      `json:"is_escalation"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewEscalationTier creates a new escalation tier
func NewEscalationTier(id uuid.UUID, name, description string, level EscalationLevel, responseTime int) *EscalationTier {
	now := time.Now()
	return &EscalationTier{
		ID:           id,
		Name:         name,
		Description:  description,
		Level:        level,
		ResponseTime: responseTime,
		Contacts:     make([]Contact, 0),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// WithNextTier links this tier to the next escalation tier
func (et *EscalationTier) WithNextTier(nextTierID uuid.UUID) *EscalationTier {
	et.NextTierID = &nextTierID
	et.UpdatedAt = time.Now()
	return et
}

// AddContact adds a contact to the escalation tier
func (et *EscalationTier) AddContact(contact Contact) *EscalationTier {
	et.Contacts = append(et.Contacts, contact)
	et.UpdatedAt = time.Time{}
	return et
}

// RemoveContact removes a contact from the escalation tier
func (et *EscalationTier) RemoveContact(contactID uuid.UUID) *EscalationTier {
	for i, contact := range et.Contacts {
		if contact.ID == contactID {
			et.Contacts = append(et.Contacts[:i], et.Contacts[i+1:]...)
			break
		}
	}
	et.UpdatedAt = time.Now()
	return et
}

// NewContact creates a new contact
func NewContact(id uuid.UUID, name, role, phone string) *Contact {
	now := time.Now()
	return &Contact{
		ID:        id,
		Name:      name,
		Role:      role,
		Phone:     phone,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// WithEmail adds an email to the contact
func (c *Contact) WithEmail(email string) *Contact {
	c.Email = email
	c.UpdatedAt = time.Now()
	return c
}

// WithFacility associates the contact with a facility
func (c *Contact) WithFacility(facilityID uuid.UUID) *Contact {
	c.FacilityID = &facilityID
	c.UpdatedAt = time.Now()
	return c
}

// MarkAsEmergency marks the contact as an emergency contact
func (c *Contact) MarkAsEmergency() *Contact {
	c.IsEmergency = true
	c.UpdatedAt = time.Now()
	return c
}

// MarkAsEscalation marks the contact as an escalation contact
func (c *Contact) MarkAsEscalation() *Contact {
	c.IsEscalation = true
	c.UpdatedAt = time.Now()
	return c
}

// EscalationPath represents a complete escalation path for emergencies
type EscalationPath struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	TierIDs     []uuid.UUID     `json:"tier_ids"`
	IsActive    bool            `json:"is_active"`
	FacilityID  *uuid.UUID      `json:"facility_id,omitempty"`
	DistrictID  *uuid.UUID      `json:"district_id,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// NewEscalationPath creates a new escalation path
func NewEscalationPath(id uuid.UUID, name, description string) *EscalationPath {
	now := time.Now()
	return &EscalationPath{
		ID:          id,
		Name:        name,
		Description: description,
		TierIDs:     make([]uuid.UUID, 0),
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// WithFacility associates the escalation path with a facility
func (ep *EscalationPath) WithFacility(facilityID uuid.UUID) *EscalationPath {
	ep.FacilityID = &facilityID
	ep.UpdatedAt = time.Now()
	return ep
}

// WithDistrict associates the escalation path with a district
func (ep *EscalationPath) WithDistrict(districtID uuid.UUID) *EscalationPath {
	ep.DistrictID = &districtID
	ep.UpdatedAt = time.Now()
	return ep
}

// AddTier adds a tier to the escalation path
func (ep *EscalationPath) AddTier(tierID uuid.UUID) *EscalationPath {
	ep.TierIDs = append(ep.TierIDs, tierID)
	ep.UpdatedAt = time.Now()
	return ep
}

// RemoveTier removes a tier from the escalation path
func (ep *EscalationPath) RemoveTier(tierID uuid.UUID) *EscalationPath {
	for i, id := range ep.TierIDs {
		if id == tierID {
			ep.TierIDs = append(ep.TierIDs[:i], ep.TierIDs[i+1:]...)
			break
		}
	}
	ep.UpdatedAt = time.Now()
	return ep
}

// Activate activates the escalation path
func (ep *EscalationPath) Activate() *EscalationPath {
	ep.IsActive = true
	ep.UpdatedAt = time.Now()
	return ep
}

// Deactivate deactivates the escalation path
func (ep *EscalationPath) Deactivate() *EscalationPath {
	ep.IsActive = false
	ep.UpdatedAt = time.Now()
	return ep
}
