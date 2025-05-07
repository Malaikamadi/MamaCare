package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/incognito25/mamacare/services/go/internal/domain/model"
)

// AmbulanceRepository defines the interface for ambulance data access
type AmbulanceRepository interface {
	// Create creates a new ambulance in the repository
	Create(ctx context.Context, ambulance *model.Ambulance) error

	// GetByID retrieves an ambulance by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*model.Ambulance, error)

	// GetAll retrieves all ambulances in the system
	GetAll(ctx context.Context) ([]*model.Ambulance, error)

	// GetByStatus retrieves ambulances with a specific status
	GetByStatus(ctx context.Context, status model.AmbulanceStatus) ([]*model.Ambulance, error)

	// GetByFacility retrieves ambulances associated with a specific facility
	GetByFacility(ctx context.Context, facilityID uuid.UUID) ([]*model.Ambulance, error)

	// GetAvailableInRadius retrieves available ambulances within a specified radius
	GetAvailableInRadius(ctx context.Context, lat, lng, radiusKm float64) ([]*model.Ambulance, error)

	// GetByType retrieves ambulances of a specific type
	GetByType(ctx context.Context, ambulanceType model.AmbulanceType) ([]*model.Ambulance, error)

	// Update updates an existing ambulance in the repository
	Update(ctx context.Context, ambulance *model.Ambulance) error

	// UpdateLocation updates just the location of an ambulance
	UpdateLocation(ctx context.Context, id uuid.UUID, lat, lng float64) error

	// UpdateStatus updates just the status of an ambulance
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.AmbulanceStatus) error

	// Delete deletes an ambulance from the repository
	Delete(ctx context.Context, id uuid.UUID) error
}
