package database

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	SQL         string
}

// MigrationManager handles database migrations
type MigrationManager struct {
	pool        *pgxpool.Pool
	logger      logger.Logger
	migrations  []Migration
	initialized bool
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(pool *pgxpool.Pool, logger logger.Logger) *MigrationManager {
	return &MigrationManager{
		pool:       pool,
		logger:     logger,
		migrations: []Migration{},
	}
}

// AddMigration adds a migration to the manager
func (mm *MigrationManager) AddMigration(version int, description, sql string) {
	mm.migrations = append(mm.migrations, Migration{
		Version:     version,
		Description: description,
		SQL:         sql,
	})
}

// Initialize sets up the migrations table if it doesn't exist
func (mm *MigrationManager) Initialize(ctx context.Context) error {
	if mm.initialized {
		return nil
	}

	// Create migrations table if it doesn't exist
	_, err := mm.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INT PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to create migrations table")
	}

	mm.initialized = true
	return nil
}

// GetAppliedMigrations gets all applied migrations
func (mm *MigrationManager) GetAppliedMigrations(ctx context.Context) (map[int]time.Time, error) {
	if err := mm.Initialize(ctx); err != nil {
		return nil, err
	}

	rows, err := mm.pool.Query(ctx, "SELECT version, applied_at FROM schema_migrations")
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query migrations")
	}
	defer rows.Close()

	appliedMigrations := make(map[int]time.Time)
	for rows.Next() {
		var version int
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan migration row")
		}
		appliedMigrations[version] = appliedAt
	}

	return appliedMigrations, nil
}

// Migrate applies all pending migrations
func (mm *MigrationManager) Migrate(ctx context.Context) error {
	appliedMigrations, err := mm.GetAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	// Sort migrations by version
	sort.Slice(mm.migrations, func(i, j int) bool {
		return mm.migrations[i].Version < mm.migrations[j].Version
	})

	// Apply pending migrations
	for _, migration := range mm.migrations {
		if _, ok := appliedMigrations[migration.Version]; !ok {
			mm.logger.Info("Applying migration",
				logger.Field{Key: "version", Value: migration.Version},
				logger.Field{Key: "description", Value: migration.Description})

			// Start transaction for this migration
			tx, err := mm.pool.Begin(ctx)
			if err != nil {
				return errorx.Wrap(err, errorx.InternalServerError, "failed to begin transaction for migration")
			}

			// Execute migration SQL
			_, err = tx.Exec(ctx, migration.SQL)
			if err != nil {
				tx.Rollback(ctx)
				return errorx.Wrap(err, errorx.InternalServerError, fmt.Sprintf("failed to apply migration %d", migration.Version))
			}

			// Record successful migration
			_, err = tx.Exec(ctx, `
				INSERT INTO schema_migrations (version, description, applied_at) 
				VALUES ($1, $2, NOW())
			`, migration.Version, migration.Description)
			if err != nil {
				tx.Rollback(ctx)
				return errorx.Wrap(err, errorx.InternalServerError, "failed to record migration")
			}

			// Commit transaction
			if err := tx.Commit(ctx); err != nil {
				return errorx.Wrap(err, errorx.InternalServerError, "failed to commit migration transaction")
			}

			mm.logger.Info("Migration applied successfully",
				logger.Field{Key: "version", Value: migration.Version})
		}
	}

	return nil
}

// GetDatabaseVersion gets the current database schema version
func (mm *MigrationManager) GetDatabaseVersion(ctx context.Context) (int, error) {
	if err := mm.Initialize(ctx); err != nil {
		return 0, err
	}

	var version int
	err := mm.pool.QueryRow(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, errorx.Wrap(err, errorx.InternalServerError, "failed to get database version")
	}

	return version, nil
}

// CreateInitialMigration creates the initial migration with basic schema
func CreateInitialMigration() Migration {
	return Migration{
		Version:     1,
		Description: "Initial schema",
		SQL: `
-- Enable PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    phone_number TEXT,
    role TEXT NOT NULL,
    district TEXT,
    facility_id UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Facilities table with geospatial data
CREATE TABLE IF NOT EXISTS facilities (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    district TEXT NOT NULL,
    location GEOMETRY(Point, 4326) NOT NULL,
    facility_type TEXT NOT NULL,
    capacity INTEGER,
    operating_hours JSONB,
    services_offered TEXT[],
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create spatial index on facilities
CREATE INDEX idx_facilities_location ON facilities USING GIST(location);

-- Mothers table
CREATE TABLE IF NOT EXISTS mothers (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    expected_delivery_date DATE NOT NULL,
    blood_type TEXT NOT NULL,
    health_conditions TEXT[],
    pregnancy_history JSONB NOT NULL,
    risk_level TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Visits/appointments table
CREATE TABLE IF NOT EXISTS visits (
    id UUID PRIMARY KEY,
    mother_id UUID NOT NULL REFERENCES mothers(id),
    facility_id UUID NOT NULL REFERENCES facilities(id),
    chw_id UUID REFERENCES users(id),
    clinician_id UUID REFERENCES users(id),
    scheduled_time TIMESTAMP WITH TIME ZONE NOT NULL,
    check_in_time TIMESTAMP WITH TIME ZONE,
    check_out_time TIMESTAMP WITH TIME ZONE,
    visit_type TEXT NOT NULL,
    visit_notes TEXT,
    status TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Health metrics table
CREATE TABLE IF NOT EXISTS health_metrics (
    id UUID PRIMARY KEY,
    mother_id UUID NOT NULL REFERENCES mothers(id),
    visit_id UUID REFERENCES visits(id),
    recorded_by UUID NOT NULL REFERENCES users(id),
    metric_type TEXT NOT NULL,
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL,
    numeric_value DOUBLE PRECISION,
    blood_pressure JSONB,
    contractions JSONB,
    notes TEXT,
    is_abnormal BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- SOS events table with geospatial data
CREATE TABLE IF NOT EXISTS sos_events (
    id UUID PRIMARY KEY,
    mother_id UUID NOT NULL REFERENCES mothers(id),
    reported_by UUID NOT NULL REFERENCES users(id),
    location GEOMETRY(Point, 4326) NOT NULL,
    nature TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL,
    ambulance_id UUID,
    facility_id UUID REFERENCES facilities(id),
    priority INTEGER NOT NULL,
    eta TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create spatial index on SOS events
CREATE INDEX idx_sos_events_location ON sos_events USING GIST(location);

-- Create indexes for common queries
CREATE INDEX idx_visits_mother_id ON visits(mother_id);
CREATE INDEX idx_visits_facility_id ON visits(facility_id);
CREATE INDEX idx_visits_scheduled_time ON visits(scheduled_time);
CREATE INDEX idx_visits_status ON visits(status);
CREATE INDEX idx_health_metrics_mother_id ON health_metrics(mother_id);
CREATE INDEX idx_health_metrics_metric_type ON health_metrics(metric_type);
CREATE INDEX idx_sos_events_mother_id ON sos_events(mother_id);
CREATE INDEX idx_sos_events_status ON sos_events(status);
		`,
	}
}
