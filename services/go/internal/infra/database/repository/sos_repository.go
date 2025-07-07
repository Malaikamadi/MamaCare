package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/domain/repository"
	"github.com/mamacare/services/internal/infra/database"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// SOSRepository implements repository.SOSRepository interface
type SOSRepository struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

// NewSOSRepository creates a new SOS repository
func NewSOSRepository(pool *pgxpool.Pool, logger logger.Logger) repository.SOSRepository {
	return &SOSRepository{
		pool:   pool,
		logger: logger,
	}
}

// scanSOSEvent scans an SOS event from a row
func scanSOSEvent(row pgx.Row) (*model.SOSEvent, error) {
	var event model.SOSEvent
	var lat, lng float64
	var ambulanceID, facilityID *uuid.UUID
	var eta *time.Time

	err := row.Scan(
		&event.ID,
		&event.MotherID,
		&event.ReportedBy,
		&lng,
		&lat,
		&event.Nature,
		&event.Description,
		&event.Status,
		&ambulanceID,
		&facilityID,
		&event.Priority,
		&eta,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errorx.New(errorx.NotFound, "SOS event not found")
		}
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan SOS event")
	}

	// Set location
	event.Location = model.Location{
		Latitude:  lat,
		Longitude: lng,
	}

	// Set optional fields
	event.AmbulanceID = ambulanceID
	event.FacilityID = facilityID
	event.ETA = eta

	return &event, nil
}

// FindByID retrieves an SOS event by ID
func (r *SOSRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.SOSEvent, error) {
	query := `
		SELECT 
			s.id, 
			s.mother_id, 
			s.reported_by, 
			ST_X(s.location::geometry) as longitude, 
			ST_Y(s.location::geometry) as latitude, 
			s.nature, 
			s.description, 
			s.status, 
			s.ambulance_id, 
			s.facility_id, 
			s.priority,
			s.eta,
			s.created_at, 
			s.updated_at
		FROM sos_events s
		WHERE s.id = $1
	`

	row := database.GetQuerier(ctx, r.pool).QueryRow(ctx, query, id)
	return scanSOSEvent(row)
}

// FindByMother retrieves SOS events for a mother
func (r *SOSRepository) FindByMother(ctx context.Context, motherID uuid.UUID) ([]*model.SOSEvent, error) {
	query := `
		SELECT 
			s.id, 
			s.mother_id, 
			s.reported_by, 
			ST_X(s.location::geometry) as longitude, 
			ST_Y(s.location::geometry) as latitude, 
			s.nature, 
			s.description, 
			s.status, 
			s.ambulance_id, 
			s.facility_id, 
			s.priority,
			s.eta,
			s.created_at, 
			s.updated_at
		FROM sos_events s
		WHERE s.mother_id = $1
		ORDER BY s.created_at DESC
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, motherID)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query SOS events by mother")
	}
	defer rows.Close()

	return scanSOSEvents(rows)
}

// FindActive retrieves all active SOS events
func (r *SOSRepository) FindActive(ctx context.Context) ([]*model.SOSEvent, error) {
	query := `
		SELECT 
			s.id, 
			s.mother_id, 
			s.reported_by, 
			ST_X(s.location::geometry) as longitude, 
			ST_Y(s.location::geometry) as latitude, 
			s.nature, 
			s.description, 
			s.status, 
			s.ambulance_id, 
			s.facility_id, 
			s.priority,
			s.eta,
			s.created_at, 
			s.updated_at
		FROM sos_events s
		WHERE s.status IN ('reported', 'dispatched')
		ORDER BY s.priority DESC, s.created_at ASC
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query active SOS events")
	}
	defer rows.Close()

	return scanSOSEvents(rows)
}

// FindByStatus retrieves SOS events by status
func (r *SOSRepository) FindByStatus(ctx context.Context, status model.SOSEventStatus) ([]*model.SOSEvent, error) {
	query := `
		SELECT 
			s.id, 
			s.mother_id, 
			s.reported_by, 
			ST_X(s.location::geometry) as longitude, 
			ST_Y(s.location::geometry) as latitude, 
			s.nature, 
			s.description, 
			s.status, 
			s.ambulance_id, 
			s.facility_id, 
			s.priority,
			s.eta,
			s.created_at, 
			s.updated_at
		FROM sos_events s
		WHERE s.status = $1
		ORDER BY s.created_at DESC
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, status)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query SOS events by status")
	}
	defer rows.Close()

	return scanSOSEvents(rows)
}

// FindByDateRange retrieves SOS events created in a date range
func (r *SOSRepository) FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*model.SOSEvent, error) {
	query := `
		SELECT 
			s.id, 
			s.mother_id, 
			s.reported_by, 
			ST_X(s.location::geometry) as longitude, 
			ST_Y(s.location::geometry) as latitude, 
			s.nature, 
			s.description, 
			s.status, 
			s.ambulance_id, 
			s.facility_id, 
			s.priority,
			s.eta,
			s.created_at, 
			s.updated_at
		FROM sos_events s
		WHERE s.created_at BETWEEN $1 AND $2
		ORDER BY s.created_at DESC
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query SOS events by date range")
	}
	defer rows.Close()

	return scanSOSEvents(rows)
}

// FindByDistrict retrieves SOS events in a district
func (r *SOSRepository) FindByDistrict(ctx context.Context, district string) ([]*model.SOSEvent, error) {
	query := `
		SELECT 
			s.id, 
			s.mother_id, 
			s.reported_by, 
			ST_X(s.location::geometry) as longitude, 
			ST_Y(s.location::geometry) as latitude, 
			s.nature, 
			s.description, 
			s.status, 
			s.ambulance_id, 
			s.facility_id, 
			s.priority,
			s.eta,
			s.created_at, 
			s.updated_at
		FROM sos_events s
		JOIN users u ON s.mother_id = u.id
		WHERE u.district = $1
		ORDER BY s.created_at DESC
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, district)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query SOS events by district")
	}
	defer rows.Close()

	return scanSOSEvents(rows)
}

// FindByNature retrieves SOS events by nature
func (r *SOSRepository) FindByNature(ctx context.Context, nature model.SOSEventNature) ([]*model.SOSEvent, error) {
	query := `
		SELECT 
			s.id, 
			s.mother_id, 
			s.reported_by, 
			ST_X(s.location::geometry) as longitude, 
			ST_Y(s.location::geometry) as latitude, 
			s.nature, 
			s.description, 
			s.status, 
			s.ambulance_id, 
			s.facility_id, 
			s.priority,
			s.eta,
			s.created_at, 
			s.updated_at
		FROM sos_events s
		WHERE s.nature = $1
		ORDER BY s.created_at DESC
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, nature)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query SOS events by nature")
	}
	defer rows.Close()

	return scanSOSEvents(rows)
}

// FindNearby retrieves SOS events near a location
func (r *SOSRepository) FindNearby(ctx context.Context, lat, lng float64, radiusKm float64) ([]*model.SOSEvent, error) {
	// Convert km to meters for ST_DWithin
	radiusMeters := radiusKm * 1000.0

	query := `
		SELECT 
			s.id, 
			s.mother_id, 
			s.reported_by, 
			ST_X(s.location::geometry) as longitude, 
			ST_Y(s.location::geometry) as latitude, 
			s.nature, 
			s.description, 
			s.status, 
			s.ambulance_id, 
			s.facility_id, 
			s.priority,
			s.eta,
			s.created_at, 
			s.updated_at
		FROM sos_events s
		WHERE ST_DWithin(
			s.location::geography,
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
			$3
		)
		AND s.status IN ('reported', 'dispatched')
		ORDER BY s.location <-> ST_SetSRID(ST_MakePoint($1, $2), 4326)
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, lng, lat, radiusMeters)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query nearby SOS events")
	}
	defer rows.Close()

	return scanSOSEvents(rows)
}

// Save creates or updates an SOS event
func (r *SOSRepository) Save(ctx context.Context, sosEvent *model.SOSEvent) error {
	// Update timestamp for modifications
	sosEvent.UpdatedAt = time.Now()

	// Use upsert to handle both insert and update
	query := `
		INSERT INTO sos_events (
			id, mother_id, reported_by, location, nature, description, status, 
			ambulance_id, facility_id, priority, eta, created_at, updated_at
		) VALUES (
			$1, $2, $3, ST_SetSRID(ST_MakePoint($4, $5), 4326), $6, $7, $8, $9, $10, $11, $12, $13, $14
		) ON CONFLICT (id) DO UPDATE SET
			mother_id = EXCLUDED.mother_id,
			reported_by = EXCLUDED.reported_by,
			location = EXCLUDED.location,
			nature = EXCLUDED.nature,
			description = EXCLUDED.description,
			status = EXCLUDED.status,
			ambulance_id = EXCLUDED.ambulance_id,
			facility_id = EXCLUDED.facility_id,
			priority = EXCLUDED.priority,
			eta = EXCLUDED.eta,
			updated_at = EXCLUDED.updated_at
	`

	_, err := database.GetQuerier(ctx, r.pool).Exec(ctx, query,
		sosEvent.ID,
		sosEvent.MotherID,
		sosEvent.ReportedBy,
		sosEvent.Location.Longitude, // PostGIS expects longitude first
		sosEvent.Location.Latitude,
		sosEvent.Nature,
		sosEvent.Description,
		sosEvent.Status,
		sosEvent.AmbulanceID,
		sosEvent.FacilityID,
		sosEvent.Priority,
		sosEvent.ETA,
		sosEvent.CreatedAt,
		sosEvent.UpdatedAt,
	)

	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to save SOS event")
	}

	return nil
}

// Delete removes an SOS event
func (r *SOSRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sos_events WHERE id = $1`

	cmdTag, err := database.GetQuerier(ctx, r.pool).Exec(ctx, query, id)
	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to delete SOS event")
	}

	if cmdTag.RowsAffected() == 0 {
		return errorx.New(errorx.NotFound, "SOS event not found")
	}

	return nil
}

// CountByStatusAndDate counts SOS events by status and date
func (r *SOSRepository) CountByStatusAndDate(ctx context.Context, date time.Time) (map[model.SOSEventStatus]int, error) {
	// Get the start and end of the specified date
	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endDate := startDate.Add(24 * time.Hour)

	query := `
		SELECT status, COUNT(id) as count
		FROM sos_events
		WHERE created_at BETWEEN $1 AND $2
		GROUP BY status
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to count SOS events by status and date")
	}
	defer rows.Close()

	counts := make(map[model.SOSEventStatus]int)
	for rows.Next() {
		var status model.SOSEventStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan SOS status count")
		}
		counts[status] = count
	}

	if err := rows.Err(); err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "error iterating over SOS status count rows")
	}

	return counts, nil
}

// scanSOSEvents scans multiple SOS events from rows
func scanSOSEvents(rows pgx.Rows) ([]*model.SOSEvent, error) {
	var events []*model.SOSEvent

	for rows.Next() {
		var event model.SOSEvent
		var lat, lng float64
		var ambulanceID, facilityID *uuid.UUID
		var eta *time.Time

		err := rows.Scan(
			&event.ID,
			&event.MotherID,
			&event.ReportedBy,
			&lng,
			&lat,
			&event.Nature,
			&event.Description,
			&event.Status,
			&ambulanceID,
			&facilityID,
			&event.Priority,
			&eta,
			&event.CreatedAt,
			&event.UpdatedAt,
		)

		if err != nil {
			return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan SOS event")
		}

		// Set location
		event.Location = model.Location{
			Latitude:  lat,
			Longitude: lng,
		}

		// Set optional fields
		event.AmbulanceID = ambulanceID
		event.FacilityID = facilityID
		event.ETA = eta

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "error iterating over SOS event rows")
	}

	return events, nil
}
