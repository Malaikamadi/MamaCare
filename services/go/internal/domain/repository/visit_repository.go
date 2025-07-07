package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/domain/model"
)

// VisitRepository defines the interface for visit data access
type VisitRepository interface {
	// Create creates a new visit
	Create(ctx context.Context, visit *model.Visit) error

	// GetByID retrieves a visit by ID
	GetByID(ctx context.Context, id uuid.UUID) (*model.Visit, error)

	// GetByMotherID retrieves visits for a mother
	GetByMotherID(ctx context.Context, motherID uuid.UUID, options *VisitQueryOptions) ([]*model.Visit, error)

	// GetByFacilityID retrieves visits for a facility
	GetByFacilityID(ctx context.Context, facilityID uuid.UUID, options *VisitQueryOptions) ([]*model.Visit, error)

	// GetByCHW retrieves visits for a CHW
	GetByCHW(ctx context.Context, chwID uuid.UUID, options *VisitQueryOptions) ([]*model.Visit, error)

	// GetUpcoming retrieves upcoming visits
	GetUpcoming(ctx context.Context, options *VisitQueryOptions) ([]*model.Visit, error)

	// GetOverdue retrieves overdue visits (missed)
	GetOverdue(ctx context.Context, options *VisitQueryOptions) ([]*model.Visit, error)

	// Update updates a visit
	Update(ctx context.Context, visit *model.Visit) error

	// Delete deletes a visit
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByDateRange retrieves visits within a date range
	GetByDateRange(ctx context.Context, startDate, endDate time.Time, options *VisitQueryOptions) ([]*model.Visit, error)
}

// VisitQueryOptions contains options for querying visits
type VisitQueryOptions struct {
	// Status filters visits by status
	Status *model.VisitStatus `json:"status,omitempty"`

	// Type filters visits by type
	Type *model.VisitType `json:"type,omitempty"`

	// Limit limits the number of results
	Limit int `json:"limit,omitempty"`

	// Offset specifies the offset for pagination
	Offset int `json:"offset,omitempty"`

	// OrderBy specifies the field to order by
	OrderBy string `json:"order_by,omitempty"`

	// OrderDirection specifies the order direction (ASC or DESC)
	OrderDirection string `json:"order_direction,omitempty"`

	// StartDate specifies the start date for filtering
	StartDate *time.Time `json:"start_date,omitempty"`

	// EndDate specifies the end date for filtering
	EndDate *time.Time `json:"end_date,omitempty"`
}

// NewVisitQueryOptions creates a new instance of VisitQueryOptions
func NewVisitQueryOptions() *VisitQueryOptions {
	return &VisitQueryOptions{
		Limit:          50,
		Offset:         0,
		OrderBy:        "scheduled_time",
		OrderDirection: "ASC",
	}
}

// WithStatus sets the status filter
func (o *VisitQueryOptions) WithStatus(status model.VisitStatus) *VisitQueryOptions {
	o.Status = &status
	return o
}

// WithType sets the type filter
func (o *VisitQueryOptions) WithType(visitType model.VisitType) *VisitQueryOptions {
	o.Type = &visitType
	return o
}

// WithLimit sets the limit
func (o *VisitQueryOptions) WithLimit(limit int) *VisitQueryOptions {
	if limit > 0 {
		o.Limit = limit
	}
	return o
}

// WithOffset sets the offset
func (o *VisitQueryOptions) WithOffset(offset int) *VisitQueryOptions {
	if offset >= 0 {
		o.Offset = offset
	}
	return o
}

// WithOrder sets the order
func (o *VisitQueryOptions) WithOrder(orderBy, direction string) *VisitQueryOptions {
	if orderBy != "" {
		o.OrderBy = orderBy
	}
	if direction == "ASC" || direction == "DESC" {
		o.OrderDirection = direction
	}
	return o
}

// WithDateRange sets the date range
func (o *VisitQueryOptions) WithDateRange(startDate, endDate time.Time) *VisitQueryOptions {
	o.StartDate = &startDate
	o.EndDate = &endDate
	return o
}
