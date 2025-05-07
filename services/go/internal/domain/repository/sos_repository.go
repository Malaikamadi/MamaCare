package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/incognito25/mamacare/services/go/internal/domain/model"
)

// SOSRepository defines the interface for SOS event data access
type SOSRepository interface {
	// Create creates a new SOS event in the repository
	Create(ctx context.Context, sosEvent *model.SOSEvent) error

	// GetByID retrieves an SOS event by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*model.SOSEvent, error)

	// GetByMotherID retrieves SOS events for a specific mother
	GetByMotherID(ctx context.Context, motherID uuid.UUID) ([]*model.SOSEvent, error)

	// GetActive retrieves all active SOS events (reported or dispatched)
	GetActive(ctx context.Context) ([]*model.SOSEvent, error)

	// GetByStatus retrieves SOS events with a specific status
	GetByStatus(ctx context.Context, status model.SOSEventStatus) ([]*model.SOSEvent, error)

	// GetByFacility retrieves SOS events associated with a specific facility
	GetByFacility(ctx context.Context, facilityID uuid.UUID) ([]*model.SOSEvent, error)

	// GetByTimeRange retrieves SOS events within a specific time range
	GetByTimeRange(ctx context.Context, start, end time.Time) ([]*model.SOSEvent, error)

	// GetInRadius retrieves active SOS events within a specified radius of coordinates
	GetInRadius(ctx context.Context, lat, lng float64, radiusKm float64) ([]*model.SOSEvent, error)

	// Update updates an existing SOS event in the repository
	Update(ctx context.Context, sosEvent *model.SOSEvent) error

	// Delete deletes an SOS event from the repository
	Delete(ctx context.Context, id uuid.UUID) error
}
