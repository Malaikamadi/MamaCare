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

// VisitRepository implements repository.VisitRepository interface
type VisitRepository struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

// NewVisitRepository creates a new visit repository
func NewVisitRepository(pool *pgxpool.Pool, logger logger.Logger) repository.VisitRepository {
	return &VisitRepository{
		pool:   pool,
		logger: logger,
	}
}

// scanVisit scans a visit from a row
func scanVisit(row pgx.Row) (*model.Visit, error) {
	var visit model.Visit
	var checkInTime, checkOutTime *time.Time
	var chwID, clinicianID *uuid.UUID

	err := row.Scan(
		&visit.ID,
		&visit.MotherID,
		&visit.FacilityID,
		&chwID,
		&clinicianID,
		&visit.ScheduledTime,
		&checkInTime,
		&checkOutTime,
		&visit.VisitType,
		&visit.VisitNotes,
		&visit.Status,
		&visit.CreatedAt,
		&visit.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errorx.New(errorx.NotFound, "visit not found")
		}
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan visit")
	}

	visit.CHWID = chwID
	visit.ClinicianID = clinicianID
	visit.CheckInTime = checkInTime
	visit.CheckOutTime = checkOutTime

	return &visit, nil
}

// FindByID retrieves a visit by ID
func (r *VisitRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Visit, error) {
	query := `
		SELECT 
			v.id, 
			v.mother_id, 
			v.facility_id, 
			v.chw_id, 
			v.clinician_id, 
			v.scheduled_time, 
			v.check_in_time, 
			v.check_out_time, 
			v.visit_type, 
			v.visit_notes, 
			v.status, 
			v.created_at, 
			v.updated_at
		FROM visits v
		WHERE v.id = $1
	`

	row := database.GetQuerier(ctx, r.pool).QueryRow(ctx, query, id)
	return scanVisit(row)
}

// FindByMother retrieves visits for a mother
func (r *VisitRepository) FindByMother(ctx context.Context, motherID uuid.UUID) ([]*model.Visit, error) {
	query := `
		SELECT 
			v.id, 
			v.mother_id, 
			v.facility_id, 
			v.chw_id, 
			v.clinician_id, 
			v.scheduled_time, 
			v.check_in_time, 
			v.check_out_time, 
			v.visit_type, 
			v.visit_notes, 
			v.status, 
			v.created_at, 
			v.updated_at
		FROM visits v
		WHERE v.mother_id = $1
		ORDER BY v.scheduled_time DESC
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, motherID)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query visits by mother")
	}
	defer rows.Close()

	return scanVisits(rows)
}

// FindByFacility retrieves visits at a facility
func (r *VisitRepository) FindByFacility(ctx context.Context, facilityID uuid.UUID) ([]*model.Visit, error) {
	query := `
		SELECT 
			v.id, 
			v.mother_id, 
			v.facility_id, 
			v.chw_id, 
			v.clinician_id, 
			v.scheduled_time, 
			v.check_in_time, 
			v.check_out_time, 
			v.visit_type, 
			v.visit_notes, 
			v.status, 
			v.created_at, 
			v.updated_at
		FROM visits v
		WHERE v.facility_id = $1
		ORDER BY v.scheduled_time
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, facilityID)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query visits by facility")
	}
	defer rows.Close()

	return scanVisits(rows)
}

// FindByCHW retrieves visits assigned to a CHW
func (r *VisitRepository) FindByCHW(ctx context.Context, chwID uuid.UUID) ([]*model.Visit, error) {
	query := `
		SELECT 
			v.id, 
			v.mother_id, 
			v.facility_id, 
			v.chw_id, 
			v.clinician_id, 
			v.scheduled_time, 
			v.check_in_time, 
			v.check_out_time, 
			v.visit_type, 
			v.visit_notes, 
			v.status, 
			v.created_at, 
			v.updated_at
		FROM visits v
		WHERE v.chw_id = $1
		ORDER BY v.scheduled_time
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, chwID)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query visits by CHW")
	}
	defer rows.Close()

	return scanVisits(rows)
}

// FindByClinician retrieves visits assigned to a clinician
func (r *VisitRepository) FindByClinician(ctx context.Context, clinicianID uuid.UUID) ([]*model.Visit, error) {
	query := `
		SELECT 
			v.id, 
			v.mother_id, 
			v.facility_id, 
			v.chw_id, 
			v.clinician_id, 
			v.scheduled_time, 
			v.check_in_time, 
			v.check_out_time, 
			v.visit_type, 
			v.visit_notes, 
			v.status, 
			v.created_at, 
			v.updated_at
		FROM visits v
		WHERE v.clinician_id = $1
		ORDER BY v.scheduled_time
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, clinicianID)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query visits by clinician")
	}
	defer rows.Close()

	return scanVisits(rows)
}

// FindByDateRange retrieves visits scheduled in a date range
func (r *VisitRepository) FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*model.Visit, error) {
	query := `
		SELECT 
			v.id, 
			v.mother_id, 
			v.facility_id, 
			v.chw_id, 
			v.clinician_id, 
			v.scheduled_time, 
			v.check_in_time, 
			v.check_out_time, 
			v.visit_type, 
			v.visit_notes, 
			v.status, 
			v.created_at, 
			v.updated_at
		FROM visits v
		WHERE v.scheduled_time BETWEEN $1 AND $2
		ORDER BY v.scheduled_time
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query visits by date range")
	}
	defer rows.Close()

	return scanVisits(rows)
}

// FindByStatus retrieves visits by status
func (r *VisitRepository) FindByStatus(ctx context.Context, status model.VisitStatus) ([]*model.Visit, error) {
	query := `
		SELECT 
			v.id, 
			v.mother_id, 
			v.facility_id, 
			v.chw_id, 
			v.clinician_id, 
			v.scheduled_time, 
			v.check_in_time, 
			v.check_out_time, 
			v.visit_type, 
			v.visit_notes, 
			v.status, 
			v.created_at, 
			v.updated_at
		FROM visits v
		WHERE v.status = $1
		ORDER BY v.scheduled_time
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, status)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query visits by status")
	}
	defer rows.Close()

	return scanVisits(rows)
}

// FindUpcoming retrieves upcoming visits for a mother
func (r *VisitRepository) FindUpcoming(ctx context.Context, motherID uuid.UUID) ([]*model.Visit, error) {
	query := `
		SELECT 
			v.id, 
			v.mother_id, 
			v.facility_id, 
			v.chw_id, 
			v.clinician_id, 
			v.scheduled_time, 
			v.check_in_time, 
			v.check_out_time, 
			v.visit_type, 
			v.visit_notes, 
			v.status, 
			v.created_at, 
			v.updated_at
		FROM visits v
		WHERE v.mother_id = $1 
		AND v.scheduled_time > NOW() 
		AND v.status = 'scheduled'
		ORDER BY v.scheduled_time
		LIMIT 5
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, motherID)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query upcoming visits")
	}
	defer rows.Close()

	return scanVisits(rows)
}

// FindUpcomingByFacility retrieves upcoming visits at a facility
func (r *VisitRepository) FindUpcomingByFacility(ctx context.Context, facilityID uuid.UUID) ([]*model.Visit, error) {
	query := `
		SELECT 
			v.id, 
			v.mother_id, 
			v.facility_id, 
			v.chw_id, 
			v.clinician_id, 
			v.scheduled_time, 
			v.check_in_time, 
			v.check_out_time, 
			v.visit_type, 
			v.visit_notes, 
			v.status, 
			v.created_at, 
			v.updated_at
		FROM visits v
		WHERE v.facility_id = $1 
		AND v.scheduled_time > NOW() 
		AND v.status = 'scheduled'
		ORDER BY v.scheduled_time
		LIMIT 20
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, facilityID)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query upcoming visits by facility")
	}
	defer rows.Close()

	return scanVisits(rows)
}

// Save creates or updates a visit
func (r *VisitRepository) Save(ctx context.Context, visit *model.Visit) error {
	// Update timestamp for modifications
	visit.UpdatedAt = time.Now()

	// Use upsert to handle both insert and update
	query := `
		INSERT INTO visits (
			id, mother_id, facility_id, chw_id, clinician_id, scheduled_time,
			check_in_time, check_out_time, visit_type, visit_notes, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		) ON CONFLICT (id) DO UPDATE SET
			mother_id = EXCLUDED.mother_id,
			facility_id = EXCLUDED.facility_id,
			chw_id = EXCLUDED.chw_id,
			clinician_id = EXCLUDED.clinician_id,
			scheduled_time = EXCLUDED.scheduled_time,
			check_in_time = EXCLUDED.check_in_time,
			check_out_time = EXCLUDED.check_out_time,
			visit_type = EXCLUDED.visit_type,
			visit_notes = EXCLUDED.visit_notes,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`

	_, err := database.GetQuerier(ctx, r.pool).Exec(ctx, query,
		visit.ID,
		visit.MotherID,
		visit.FacilityID,
		visit.CHWID,
		visit.ClinicianID,
		visit.ScheduledTime,
		visit.CheckInTime,
		visit.CheckOutTime,
		visit.VisitType,
		visit.VisitNotes,
		visit.Status,
		visit.CreatedAt,
		visit.UpdatedAt,
	)

	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to save visit")
	}

	return nil
}

// Delete removes a visit
func (r *VisitRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM visits WHERE id = $1`

	cmdTag, err := database.GetQuerier(ctx, r.pool).Exec(ctx, query, id)
	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to delete visit")
	}

	if cmdTag.RowsAffected() == 0 {
		return errorx.New(errorx.NotFound, "visit not found")
	}

	return nil
}

// CountByStatusAndDate counts visits by status and date
func (r *VisitRepository) CountByStatusAndDate(ctx context.Context, date time.Time) (map[model.VisitStatus]int, error) {
	// Get the start and end of the specified date
	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endDate := startDate.Add(24 * time.Hour)

	query := `
		SELECT status, COUNT(id) as count
		FROM visits
		WHERE scheduled_time BETWEEN $1 AND $2
		GROUP BY status
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to count visits by status and date")
	}
	defer rows.Close()

	counts := make(map[model.VisitStatus]int)
	for rows.Next() {
		var status model.VisitStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan visit status count")
		}
		counts[status] = count
	}

	if err := rows.Err(); err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "error iterating over visit status count rows")
	}

	return counts, nil
}

// scanVisits scans multiple visits from rows
func scanVisits(rows pgx.Rows) ([]*model.Visit, error) {
	var visits []*model.Visit

	for rows.Next() {
		var visit model.Visit
		var checkInTime, checkOutTime *time.Time
		var chwID, clinicianID *uuid.UUID

		err := rows.Scan(
			&visit.ID,
			&visit.MotherID,
			&visit.FacilityID,
			&chwID,
			&clinicianID,
			&visit.ScheduledTime,
			&checkInTime,
			&checkOutTime,
			&visit.VisitType,
			&visit.VisitNotes,
			&visit.Status,
			&visit.CreatedAt,
			&visit.UpdatedAt,
		)

		if err != nil {
			return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan visit")
		}

		visit.CHWID = chwID
		visit.ClinicianID = clinicianID
		visit.CheckInTime = checkInTime
		visit.CheckOutTime = checkOutTime

		visits = append(visits, &visit)
	}

	if err := rows.Err(); err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "error iterating over visit rows")
	}

	return visits, nil
}
