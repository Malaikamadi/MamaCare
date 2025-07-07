package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/pkg/errorx"
)

// Constants for PostGIS operations
const (
	// SRID 4326 is the WGS84 GPS coordinate system used worldwide
	SRID = 4326
)

// SpatialQueries contains PostGIS related query utilities
type SpatialQueries struct {
	querier Querier
}

// NewSpatialQueries creates a new SpatialQueries instance
func NewSpatialQueries(querier Querier) *SpatialQueries {
	return &SpatialQueries{
		querier: querier,
	}
}

// MakePoint generates a PostGIS point from latitude and longitude
func MakePoint(lat, lng float64) string {
	return fmt.Sprintf("ST_SetSRID(ST_MakePoint(%f, %f), %d)", lng, lat, SRID)
}

// PointFromLocation creates a PostGIS point from a Location
func PointFromLocation(location model.Location) string {
	return MakePoint(location.Latitude, location.Longitude)
}

// FindWithinRadius finds items within a radius of a point
// table: The table to query
// geomColumn: The PostGIS geometry column name
// lat, lng: The center point coordinates
// radiusMeters: The radius in meters
// selectColumns: Columns to select
// additionalConditions: Any additional WHERE conditions
// args: Arguments for the additional conditions
func (sq *SpatialQueries) FindWithinRadius(
	ctx context.Context,
	table, geomColumn string,
	lat, lng float64,
	radiusMeters float64,
	selectColumns []string,
	additionalConditions string,
	args ...interface{},
) (pgx.Rows, error) {
	// Create point string
	pointStr := MakePoint(lat, lng)
	
	// Build query
	query := fmt.Sprintf(`
		SELECT %s, ST_Distance(
			%s,
			%s::geography
		) AS distance
		FROM %s
		WHERE ST_DWithin(
			%s::geography,
			%s::geography,
			$1
		)`,
		selectColumnsString(selectColumns),
		geomColumn,
		pointStr,
		table,
		geomColumn,
		pointStr,
	)
	
	// Add additional conditions if provided
	if additionalConditions != "" {
		query += " AND " + additionalConditions
	}
	
	// Add order by distance
	query += " ORDER BY distance ASC"
	
	// Add radius as first parameter
	allArgs := make([]interface{}, 0, len(args)+1)
	allArgs = append(allArgs, radiusMeters)
	allArgs = append(allArgs, args...)
	
	// Execute query
	rows, err := sq.querier.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to execute spatial query")
	}
	
	return rows, nil
}

// CalculateDistance calculates the distance between two points
func (sq *SpatialQueries) CalculateDistance(
	ctx context.Context,
	lat1, lng1, lat2, lng2 float64,
) (float64, error) {
	point1 := MakePoint(lat1, lng1)
	point2 := MakePoint(lat2, lng2)
	
	query := fmt.Sprintf(`
		SELECT ST_Distance(
			%s::geography,
			%s::geography
		)`,
		point1, point2,
	)
	
	var distance float64
	err := sq.querier.QueryRow(ctx, query).Scan(&distance)
	if err != nil {
		return 0, errorx.Wrap(err, errorx.InternalServerError, "failed to calculate distance")
	}
	
	return distance, nil
}

// CalculateDistanceFromLocation calculates the distance between two locations
func (sq *SpatialQueries) CalculateDistanceFromLocation(
	ctx context.Context,
	loc1, loc2 model.Location,
) (float64, error) {
	return sq.CalculateDistance(ctx, loc1.Latitude, loc1.Longitude, loc2.Latitude, loc2.Longitude)
}

// FindNearest finds the nearest items to a point
func (sq *SpatialQueries) FindNearest(
	ctx context.Context,
	table, geomColumn string,
	lat, lng float64,
	limit int,
	selectColumns []string,
	additionalConditions string,
	args ...interface{},
) (pgx.Rows, error) {
	// Create point string
	pointStr := MakePoint(lat, lng)
	
	// Build query
	query := fmt.Sprintf(`
		SELECT %s, ST_Distance(
			%s,
			%s::geography
		) AS distance
		FROM %s`,
		selectColumnsString(selectColumns),
		geomColumn,
		pointStr,
		table,
	)
	
	// Add additional conditions if provided
	if additionalConditions != "" {
		query += " WHERE " + additionalConditions
	}
	
	// Add order by distance and limit
	query += fmt.Sprintf(" ORDER BY %s <-> %s::geography LIMIT $1", geomColumn, pointStr)
	
	// Add limit as first parameter
	allArgs := make([]interface{}, 0, len(args)+1)
	allArgs = append(allArgs, limit)
	allArgs = append(allArgs, args...)
	
	// Execute query
	rows, err := sq.querier.Query(ctx, query, allArgs...)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to execute nearest spatial query")
	}
	
	return rows, nil
}

// Helper function to create a comma-separated list of columns
func selectColumnsString(columns []string) string {
	if len(columns) == 0 {
		return "*"
	}
	
	result := ""
	for i, col := range columns {
		if i > 0 {
			result += ", "
		}
		result += col
	}
	
	return result
}
