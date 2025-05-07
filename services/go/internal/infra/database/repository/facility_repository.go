package repository

import (
	"context"
	"encoding/json"
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

// FacilityRepository implements repository.FacilityRepository interface
type FacilityRepository struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

// NewFacilityRepository creates a new facility repository
func NewFacilityRepository(pool *pgxpool.Pool, logger logger.Logger) repository.FacilityRepository {
	return &FacilityRepository{
		pool:   pool,
		logger: logger,
	}
}

// scanFacility scans a facility from a row
func scanFacility(row pgx.Row) (*model.HealthcareFacility, error) {
	var facility model.HealthcareFacility
	var lat, lng float64
	var operatingHoursJSON []byte
	var servicesOffered []string

	err := row.Scan(
		&facility.ID,
		&facility.Name,
		&facility.Address,
		&facility.District,
		&lat,
		&lng,
		&facility.FacilityType,
		&facility.Capacity,
		&operatingHoursJSON,
		&servicesOffered,
		&facility.CreatedAt,
		&facility.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errorx.New(errorx.NotFound, "facility not found")
		}
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan facility")
	}

	// Set location
	facility.Location = model.Location{
		Latitude:  lat,
		Longitude: lng,
	}

	// Unmarshal operating hours JSON
	if operatingHoursJSON != nil {
		if err := json.Unmarshal(operatingHoursJSON, &facility.OperatingHours); err != nil {
			return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to unmarshal operating hours")
		}
	}

	// Set services offered
	facility.ServicesOffered = servicesOffered

	return &facility, nil
}

// FindByID retrieves a facility by ID
func (r *FacilityRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.HealthcareFacility, error) {
	query := `
		SELECT 
			f.id, 
			f.name, 
			f.address, 
			f.district, 
			ST_X(f.location::geometry) as longitude, 
			ST_Y(f.location::geometry) as latitude, 
			f.facility_type, 
			f.capacity, 
			f.operating_hours, 
			f.services_offered, 
			f.created_at, 
			f.updated_at
		FROM facilities f
		WHERE f.id = $1
	`

	row := database.GetQuerier(ctx, r.pool).QueryRow(ctx, query, id)
	return scanFacility(row)
}

// FindAll retrieves all facilities
func (r *FacilityRepository) FindAll(ctx context.Context) ([]*model.HealthcareFacility, error) {
	query := `
		SELECT 
			f.id, 
			f.name, 
			f.address, 
			f.district, 
			ST_X(f.location::geometry) as longitude, 
			ST_Y(f.location::geometry) as latitude, 
			f.facility_type, 
			f.capacity, 
			f.operating_hours, 
			f.services_offered, 
			f.created_at, 
			f.updated_at
		FROM facilities f
		ORDER BY f.name
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query facilities")
	}
	defer rows.Close()

	return scanFacilities(rows)
}

// FindByDistrict retrieves facilities by district
func (r *FacilityRepository) FindByDistrict(ctx context.Context, district string) ([]*model.HealthcareFacility, error) {
	query := `
		SELECT 
			f.id, 
			f.name, 
			f.address, 
			f.district, 
			ST_X(f.location::geometry) as longitude, 
			ST_Y(f.location::geometry) as latitude, 
			f.facility_type, 
			f.capacity, 
			f.operating_hours, 
			f.services_offered, 
			f.created_at, 
			f.updated_at
		FROM facilities f
		WHERE f.district = $1
		ORDER BY f.name
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, district)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query facilities by district")
	}
	defer rows.Close()

	return scanFacilities(rows)
}

// FindByType retrieves facilities by type
func (r *FacilityRepository) FindByType(ctx context.Context, facilityType model.FacilityType) ([]*model.HealthcareFacility, error) {
	query := `
		SELECT 
			f.id, 
			f.name, 
			f.address, 
			f.district, 
			ST_X(f.location::geometry) as longitude, 
			ST_Y(f.location::geometry) as latitude, 
			f.facility_type, 
			f.capacity, 
			f.operating_hours, 
			f.services_offered, 
			f.created_at, 
			f.updated_at
		FROM facilities f
		WHERE f.facility_type = $1
		ORDER BY f.name
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, facilityType)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query facilities by type")
	}
	defer rows.Close()

	return scanFacilities(rows)
}

// FindByService retrieves facilities that offer a specific service
func (r *FacilityRepository) FindByService(ctx context.Context, service string) ([]*model.HealthcareFacility, error) {
	query := `
		SELECT 
			f.id, 
			f.name, 
			f.address, 
			f.district, 
			ST_X(f.location::geometry) as longitude, 
			ST_Y(f.location::geometry) as latitude, 
			f.facility_type, 
			f.capacity, 
			f.operating_hours, 
			f.services_offered, 
			f.created_at, 
			f.updated_at
		FROM facilities f
		WHERE $1 = ANY(f.services_offered)
		ORDER BY f.name
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, service)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query facilities by service")
	}
	defer rows.Close()

	return scanFacilities(rows)
}

// FindNearby retrieves facilities within a radius of a location
func (r *FacilityRepository) FindNearby(ctx context.Context, lat, lng float64, radiusKm float64) ([]*model.HealthcareFacility, error) {
	query := `
		SELECT 
			f.id, 
			f.name, 
			f.address, 
			f.district, 
			ST_X(f.location::geometry) as longitude, 
			ST_Y(f.location::geometry) as latitude, 
			f.facility_type, 
			f.capacity, 
			f.operating_hours, 
			f.services_offered, 
			f.created_at, 
			f.updated_at
		FROM facilities f
		WHERE ST_DWithin(
			f.location, 
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, 
			$3 * 1000
		)
		ORDER BY f.location <-> ST_SetSRID(ST_MakePoint($1, $2), 4326)
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, lng, lat, radiusKm)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query nearby facilities")
	}
	defer rows.Close()

	return scanFacilities(rows)
}

// Save creates or updates a facility
func (r *FacilityRepository) Save(ctx context.Context, facility *model.HealthcareFacility) error {
	// Update timestamp for modifications
	facility.UpdatedAt = time.Now()

	// Marshal operating hours to JSON
	operatingHoursJSON, err := json.Marshal(facility.OperatingHours)
	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to marshal operating hours")
	}

	// Use upsert to handle both insert and update
	query := `
		INSERT INTO facilities (
			id, name, address, district, location, facility_type, capacity, 
			operating_hours, services_offered, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, ST_SetSRID(ST_MakePoint($5, $6), 4326), $7, $8, $9, $10, $11, $12
		) ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			address = EXCLUDED.address,
			district = EXCLUDED.district,
			location = EXCLUDED.location,
			facility_type = EXCLUDED.facility_type,
			capacity = EXCLUDED.capacity,
			operating_hours = EXCLUDED.operating_hours,
			services_offered = EXCLUDED.services_offered,
			updated_at = EXCLUDED.updated_at
	`

	_, err = database.GetQuerier(ctx, r.pool).Exec(ctx, query,
		facility.ID,
		facility.Name,
		facility.Address,
		facility.District,
		facility.Location.Longitude, // PostGIS expects longitude first
		facility.Location.Latitude,
		facility.FacilityType,
		facility.Capacity,
		operatingHoursJSON,
		facility.ServicesOffered,
		facility.CreatedAt,
		facility.UpdatedAt,
	)

	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to save facility")
	}

	return nil
}

// Delete removes a facility
func (r *FacilityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM facilities WHERE id = $1`

	_, err := database.GetQuerier(ctx, r.pool).Exec(ctx, query, id)
	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to delete facility")
	}

	return nil
}

// FindByName retrieves a facility by name
func (r *FacilityRepository) FindByName(ctx context.Context, name string) (*model.HealthcareFacility, error) {
	query := `
		SELECT 
			f.id, 
			f.name, 
			f.address, 
			f.district, 
			ST_X(f.location::geometry) as longitude, 
			ST_Y(f.location::geometry) as latitude, 
			f.facility_type, 
			f.capacity, 
			f.operating_hours, 
			f.services_offered, 
			f.created_at, 
			f.updated_at
		FROM facilities f
		WHERE f.name = $1
	`

	row := database.GetQuerier(ctx, r.pool).QueryRow(ctx, query, name)
	return scanFacility(row)
}

// CountByType counts facilities by type
func (r *FacilityRepository) CountByType(ctx context.Context) (map[model.FacilityType]int, error) {
	query := `
		SELECT 
			f.facility_type, 
			COUNT(*) as count
		FROM facilities f
		GROUP BY f.facility_type
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to count facilities by type")
	}
	defer rows.Close()

	result := make(map[model.FacilityType]int)
	for rows.Next() {
		var facilityType model.FacilityType
		var count int

		if err := rows.Scan(&facilityType, &count); err != nil {
			return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan facility count")
		}

		result[facilityType] = count
	}

	if err := rows.Err(); err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "error iterating facility count rows")
	}

	return result, nil
}

// scanFacilities scans multiple facilities from rows
func scanFacilities(rows pgx.Rows) ([]*model.HealthcareFacility, error) {
	var facilities []*model.HealthcareFacility

	for rows.Next() {
		var facility model.HealthcareFacility
		var lat, lng float64
		var operatingHoursJSON []byte
		var servicesOffered []string

		err := rows.Scan(
			&facility.ID,
			&facility.Name,
			&facility.Address,
			&facility.District,
			&lng,
			&lat,
			&facility.FacilityType,
			&facility.Capacity,
			&operatingHoursJSON,
			&servicesOffered,
			&facility.CreatedAt,
			&facility.UpdatedAt,
		)

		if err != nil {
			return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan facility")
		}

		// Set location
		facility.Location = model.Location{
			Latitude:  lat,
			Longitude: lng,
		}

		// Unmarshal operating hours JSON
		if operatingHoursJSON != nil {
			if err := json.Unmarshal(operatingHoursJSON, &facility.OperatingHours); err != nil {
				return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to unmarshal operating hours")
			}
		}

		// Set services offered
		facility.ServicesOffered = servicesOffered

		facilities = append(facilities, &facility)
	}

	if err := rows.Err(); err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "error iterating over facility rows")
	}

	return facilities, nil
}
