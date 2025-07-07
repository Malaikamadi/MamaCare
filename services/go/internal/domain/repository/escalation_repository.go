package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/incognito25/mamacare/services/go/internal/domain/model"
)

// EscalationTierRepository defines the interface for escalation tier data access
type EscalationTierRepository interface {
	// Create creates a new escalation tier
	Create(ctx context.Context, tier *model.EscalationTier) error

	// GetByID retrieves an escalation tier by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*model.EscalationTier, error)

	// GetAll retrieves all escalation tiers
	GetAll(ctx context.Context) ([]*model.EscalationTier, error)

	// GetByLevel retrieves escalation tiers of a specific level
	GetByLevel(ctx context.Context, level model.EscalationLevel) ([]*model.EscalationTier, error)

	// Update updates an existing escalation tier
	Update(ctx context.Context, tier *model.EscalationTier) error

	// Delete deletes an escalation tier
	Delete(ctx context.Context, id uuid.UUID) error
}

// ContactRepository defines the interface for contact data access
type ContactRepository interface {
	// Create creates a new contact
	Create(ctx context.Context, contact *model.Contact) error

	// GetByID retrieves a contact by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*model.Contact, error)

	// GetAll retrieves all contacts
	GetAll(ctx context.Context) ([]*model.Contact, error)

	// GetByFacility retrieves contacts associated with a specific facility
	GetByFacility(ctx context.Context, facilityID uuid.UUID) ([]*model.Contact, error)

	// GetEmergencyContacts retrieves all emergency contacts
	GetEmergencyContacts(ctx context.Context) ([]*model.Contact, error)

	// GetEscalationContacts retrieves all escalation contacts
	GetEscalationContacts(ctx context.Context) ([]*model.Contact, error)

	// Update updates an existing contact
	Update(ctx context.Context, contact *model.Contact) error

	// Delete deletes a contact
	Delete(ctx context.Context, id uuid.UUID) error
}

// EscalationPathRepository defines the interface for escalation path data access
type EscalationPathRepository interface {
	// Create creates a new escalation path
	Create(ctx context.Context, path *model.EscalationPath) error

	// GetByID retrieves an escalation path by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*model.EscalationPath, error)

	// GetAll retrieves all escalation paths
	GetAll(ctx context.Context) ([]*model.EscalationPath, error)

	// GetActive retrieves all active escalation paths
	GetActive(ctx context.Context) ([]*model.EscalationPath, error)

	// GetByFacility retrieves escalation paths for a specific facility
	GetByFacility(ctx context.Context, facilityID uuid.UUID) ([]*model.EscalationPath, error)

	// GetByDistrict retrieves escalation paths for a specific district
	GetByDistrict(ctx context.Context, districtID uuid.UUID) ([]*model.EscalationPath, error)

	// Update updates an existing escalation path
	Update(ctx context.Context, path *model.EscalationPath) error

	// Delete deletes an escalation path
	Delete(ctx context.Context, id uuid.UUID) error
}
