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

// HealthMetricRepository implements repository.HealthMetricRepository interface
type HealthMetricRepository struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

// NewHealthMetricRepository creates a new health metric repository
func NewHealthMetricRepository(pool *pgxpool.Pool, logger logger.Logger) repository.HealthMetricRepository {
	return &HealthMetricRepository{
		pool:   pool,
		logger: logger,
	}
}

// scanHealthMetric scans a health metric from a row
func scanHealthMetric(row pgx.Row) (*model.HealthMetric, error) {
	var metric model.HealthMetric
	var recordedByID *uuid.UUID
	var bloodPressureSystolic, bloodPressureDiastolic, fetalHeartRate, fetalMovement *float64
	var bloodSugar, hemoglobinLevel, ironLevel, weight *float64

	err := row.Scan(
		&metric.ID,
		&metric.MotherID,
		&metric.VisitID,
		&recordedByID,
		&metric.RecordedAt,
		&bloodPressureSystolic,
		&bloodPressureDiastolic,
		&fetalHeartRate,
		&fetalMovement,
		&bloodSugar,
		&hemoglobinLevel,
		&ironLevel,
		&weight,
		&metric.Notes,
		&metric.CreatedAt,
		&metric.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errorx.New(errorx.NotFound, "health metric not found")
		}
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan health metric")
	}

	// Set optional fields
	metric.RecordedByID = recordedByID
	
	// Set vital signs
	metric.VitalSigns = model.VitalSigns{}
	if bloodPressureSystolic != nil && bloodPressureDiastolic != nil {
		metric.VitalSigns.BloodPressure = &model.BloodPressure{
			Systolic:  *bloodPressureSystolic,
			Diastolic: *bloodPressureDiastolic,
		}
	}
	
	metric.VitalSigns.FetalHeartRate = fetalHeartRate
	metric.VitalSigns.FetalMovement = fetalMovement
	metric.VitalSigns.BloodSugar = bloodSugar
	metric.VitalSigns.HemoglobinLevel = hemoglobinLevel
	metric.VitalSigns.IronLevel = ironLevel
	metric.VitalSigns.Weight = weight

	return &metric, nil
}

// FindByID retrieves a health metric by ID
func (r *HealthMetricRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.HealthMetric, error) {
	query := `
		SELECT 
			h.id, 
			h.mother_id, 
			h.visit_id, 
			h.recorded_by, 
			h.recorded_at, 
			h.blood_pressure_systolic, 
			h.blood_pressure_diastolic, 
			h.fetal_heart_rate, 
			h.fetal_movement, 
			h.blood_sugar, 
			h.hemoglobin_level, 
			h.iron_level, 
			h.weight, 
			h.notes, 
			h.created_at, 
			h.updated_at
		FROM health_metrics h
		WHERE h.id = $1
	`

	row := database.GetQuerier(ctx, r.pool).QueryRow(ctx, query, id)
	return scanHealthMetric(row)
}

// FindByMother retrieves health metrics for a mother
func (r *HealthMetricRepository) FindByMother(ctx context.Context, motherID uuid.UUID) ([]*model.HealthMetric, error) {
	query := `
		SELECT 
			h.id, 
			h.mother_id, 
			h.visit_id, 
			h.recorded_by, 
			h.recorded_at, 
			h.blood_pressure_systolic, 
			h.blood_pressure_diastolic, 
			h.fetal_heart_rate, 
			h.fetal_movement, 
			h.blood_sugar, 
			h.hemoglobin_level, 
			h.iron_level, 
			h.weight, 
			h.notes, 
			h.created_at, 
			h.updated_at
		FROM health_metrics h
		WHERE h.mother_id = $1
		ORDER BY h.recorded_at DESC
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, motherID)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query health metrics by mother")
	}
	defer rows.Close()

	return scanHealthMetrics(rows)
}

// FindByVisit retrieves health metrics for a visit
func (r *HealthMetricRepository) FindByVisit(ctx context.Context, visitID uuid.UUID) ([]*model.HealthMetric, error) {
	query := `
		SELECT 
			h.id, 
			h.mother_id, 
			h.visit_id, 
			h.recorded_by, 
			h.recorded_at, 
			h.blood_pressure_systolic, 
			h.blood_pressure_diastolic, 
			h.fetal_heart_rate, 
			h.fetal_movement, 
			h.blood_sugar, 
			h.hemoglobin_level, 
			h.iron_level, 
			h.weight, 
			h.notes, 
			h.created_at, 
			h.updated_at
		FROM health_metrics h
		WHERE h.visit_id = $1
		ORDER BY h.recorded_at
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, visitID)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query health metrics by visit")
	}
	defer rows.Close()

	return scanHealthMetrics(rows)
}

// FindByDateRange retrieves health metrics recorded in a date range
func (r *HealthMetricRepository) FindByDateRange(ctx context.Context, motherID uuid.UUID, startDate, endDate time.Time) ([]*model.HealthMetric, error) {
	query := `
		SELECT 
			h.id, 
			h.mother_id, 
			h.visit_id, 
			h.recorded_by, 
			h.recorded_at, 
			h.blood_pressure_systolic, 
			h.blood_pressure_diastolic, 
			h.fetal_heart_rate, 
			h.fetal_movement, 
			h.blood_sugar, 
			h.hemoglobin_level, 
			h.iron_level, 
			h.weight, 
			h.notes, 
			h.created_at, 
			h.updated_at
		FROM health_metrics h
		WHERE h.mother_id = $1 AND h.recorded_at BETWEEN $2 AND $3
		ORDER BY h.recorded_at
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, motherID, startDate, endDate)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query health metrics by date range")
	}
	defer rows.Close()

	return scanHealthMetrics(rows)
}

// FindLatest retrieves the latest health metric for a mother
func (r *HealthMetricRepository) FindLatest(ctx context.Context, motherID uuid.UUID) (*model.HealthMetric, error) {
	query := `
		SELECT 
			h.id, 
			h.mother_id, 
			h.visit_id, 
			h.recorded_by, 
			h.recorded_at, 
			h.blood_pressure_systolic, 
			h.blood_pressure_diastolic, 
			h.fetal_heart_rate, 
			h.fetal_movement, 
			h.blood_sugar, 
			h.hemoglobin_level, 
			h.iron_level, 
			h.weight, 
			h.notes, 
			h.created_at, 
			h.updated_at
		FROM health_metrics h
		WHERE h.mother_id = $1
		ORDER BY h.recorded_at DESC
		LIMIT 1
	`

	row := database.GetQuerier(ctx, r.pool).QueryRow(ctx, query, motherID)
	return scanHealthMetric(row)
}

// FindAbnormalBloodPressure retrieves recent records with abnormal blood pressure
func (r *HealthMetricRepository) FindAbnormalBloodPressure(ctx context.Context, threshold float64) ([]*model.HealthMetric, error) {
	query := `
		SELECT 
			h.id, 
			h.mother_id, 
			h.visit_id, 
			h.recorded_by, 
			h.recorded_at, 
			h.blood_pressure_systolic, 
			h.blood_pressure_diastolic, 
			h.fetal_heart_rate, 
			h.fetal_movement, 
			h.blood_sugar, 
			h.hemoglobin_level, 
			h.iron_level, 
			h.weight, 
			h.notes, 
			h.created_at, 
			h.updated_at
		FROM health_metrics h
		WHERE (h.blood_pressure_systolic > $1 OR h.blood_pressure_diastolic > $2)
		AND h.recorded_at > NOW() - INTERVAL '30 days'
		ORDER BY h.recorded_at DESC
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, threshold, threshold*0.6) // Diastolic is typically ~60% of systolic
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query abnormal blood pressure")
	}
	defer rows.Close()

	return scanHealthMetrics(rows)
}

// GetWeightTrend retrieves weight trend for a mother
func (r *HealthMetricRepository) GetWeightTrend(ctx context.Context, motherID uuid.UUID, months int) ([]model.WeightRecord, error) {
	query := `
		SELECT 
			h.recorded_at, 
			h.weight
		FROM health_metrics h
		WHERE h.mother_id = $1 
		AND h.weight IS NOT NULL
		AND h.recorded_at > NOW() - INTERVAL '$2 months'
		ORDER BY h.recorded_at
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, motherID, months)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query weight trend")
	}
	defer rows.Close()

	var trend []model.WeightRecord
	for rows.Next() {
		var record model.WeightRecord
		var weight float64
		
		if err := rows.Scan(&record.Date, &weight); err != nil {
			return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan weight record")
		}
		
		record.Weight = weight
		trend = append(trend, record)
	}

	if err := rows.Err(); err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "error iterating over weight records")
	}

	return trend, nil
}

// Save creates or updates a health metric
func (r *HealthMetricRepository) Save(ctx context.Context, metric *model.HealthMetric) error {
	// Update timestamp for modifications
	metric.UpdatedAt = time.Now()

	// Extract blood pressure values
	var bloodPressureSystolic, bloodPressureDiastolic *float64
	if metric.VitalSigns.BloodPressure != nil {
		systolic := metric.VitalSigns.BloodPressure.Systolic
		diastolic := metric.VitalSigns.BloodPressure.Diastolic
		bloodPressureSystolic = &systolic
		bloodPressureDiastolic = &diastolic
	}

	// Use upsert to handle both insert and update
	query := `
		INSERT INTO health_metrics (
			id, mother_id, visit_id, recorded_by, recorded_at, 
			blood_pressure_systolic, blood_pressure_diastolic, fetal_heart_rate, fetal_movement,
			blood_sugar, hemoglobin_level, iron_level, weight, notes, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		) ON CONFLICT (id) DO UPDATE SET
			mother_id = EXCLUDED.mother_id,
			visit_id = EXCLUDED.visit_id,
			recorded_by = EXCLUDED.recorded_by,
			recorded_at = EXCLUDED.recorded_at,
			blood_pressure_systolic = EXCLUDED.blood_pressure_systolic,
			blood_pressure_diastolic = EXCLUDED.blood_pressure_diastolic,
			fetal_heart_rate = EXCLUDED.fetal_heart_rate,
			fetal_movement = EXCLUDED.fetal_movement,
			blood_sugar = EXCLUDED.blood_sugar,
			hemoglobin_level = EXCLUDED.hemoglobin_level,
			iron_level = EXCLUDED.iron_level,
			weight = EXCLUDED.weight,
			notes = EXCLUDED.notes,
			updated_at = EXCLUDED.updated_at
	`

	_, err := database.GetQuerier(ctx, r.pool).Exec(ctx, query,
		metric.ID,
		metric.MotherID,
		metric.VisitID,
		metric.RecordedByID,
		metric.RecordedAt,
		bloodPressureSystolic,
		bloodPressureDiastolic,
		metric.VitalSigns.FetalHeartRate,
		metric.VitalSigns.FetalMovement,
		metric.VitalSigns.BloodSugar,
		metric.VitalSigns.HemoglobinLevel,
		metric.VitalSigns.IronLevel,
		metric.VitalSigns.Weight,
		metric.Notes,
		metric.CreatedAt,
		metric.UpdatedAt,
	)

	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to save health metric")
	}

	return nil
}

// Delete removes a health metric
func (r *HealthMetricRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM health_metrics WHERE id = $1`

	cmdTag, err := database.GetQuerier(ctx, r.pool).Exec(ctx, query, id)
	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to delete health metric")
	}

	if cmdTag.RowsAffected() == 0 {
		return errorx.New(errorx.NotFound, "health metric not found")
	}

	return nil
}

// GetVitalSignsAverage calculates average vital signs for a mother over a period
func (r *HealthMetricRepository) GetVitalSignsAverage(ctx context.Context, motherID uuid.UUID, days int) (*model.VitalSigns, error) {
	query := `
		SELECT 
			AVG(blood_pressure_systolic) as avg_systolic,
			AVG(blood_pressure_diastolic) as avg_diastolic,
			AVG(fetal_heart_rate) as avg_fetal_heart_rate,
			AVG(fetal_movement) as avg_fetal_movement,
			AVG(blood_sugar) as avg_blood_sugar,
			AVG(hemoglobin_level) as avg_hemoglobin,
			AVG(iron_level) as avg_iron,
			AVG(weight) as avg_weight
		FROM health_metrics
		WHERE mother_id = $1
		AND recorded_at > NOW() - INTERVAL '$2 days'
	`

	var avgSystolic, avgDiastolic, avgFetalHeartRate, avgFetalMovement *float64
	var avgBloodSugar, avgHemoglobin, avgIron, avgWeight *float64

	err := database.GetQuerier(ctx, r.pool).QueryRow(ctx, query, motherID, days).Scan(
		&avgSystolic,
		&avgDiastolic,
		&avgFetalHeartRate,
		&avgFetalMovement,
		&avgBloodSugar,
		&avgHemoglobin,
		&avgIron,
		&avgWeight,
	)

	if err != nil && err != pgx.ErrNoRows {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to calculate vital signs average")
	}

	vitalSigns := &model.VitalSigns{
		FetalHeartRate:  avgFetalHeartRate,
		FetalMovement:   avgFetalMovement,
		BloodSugar:      avgBloodSugar,
		HemoglobinLevel: avgHemoglobin,
		IronLevel:       avgIron,
		Weight:          avgWeight,
	}

	if avgSystolic != nil && avgDiastolic != nil {
		vitalSigns.BloodPressure = &model.BloodPressure{
			Systolic:  *avgSystolic,
			Diastolic: *avgDiastolic,
		}
	}

	return vitalSigns, nil
}

// scanHealthMetrics scans multiple health metrics from rows
func scanHealthMetrics(rows pgx.Rows) ([]*model.HealthMetric, error) {
	var metrics []*model.HealthMetric

	for rows.Next() {
		var metric model.HealthMetric
		var recordedByID *uuid.UUID
		var bloodPressureSystolic, bloodPressureDiastolic, fetalHeartRate, fetalMovement *float64
		var bloodSugar, hemoglobinLevel, ironLevel, weight *float64

		err := rows.Scan(
			&metric.ID,
			&metric.MotherID,
			&metric.VisitID,
			&recordedByID,
			&metric.RecordedAt,
			&bloodPressureSystolic,
			&bloodPressureDiastolic,
			&fetalHeartRate,
			&fetalMovement,
			&bloodSugar,
			&hemoglobinLevel,
			&ironLevel,
			&weight,
			&metric.Notes,
			&metric.CreatedAt,
			&metric.UpdatedAt,
		)

		if err != nil {
			return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan health metric")
		}

		// Set optional fields
		metric.RecordedByID = recordedByID
		
		// Set vital signs
		metric.VitalSigns = model.VitalSigns{}
		if bloodPressureSystolic != nil && bloodPressureDiastolic != nil {
			metric.VitalSigns.BloodPressure = &model.BloodPressure{
				Systolic:  *bloodPressureSystolic,
				Diastolic: *bloodPressureDiastolic,
			}
		}
		
		metric.VitalSigns.FetalHeartRate = fetalHeartRate
		metric.VitalSigns.FetalMovement = fetalMovement
		metric.VitalSigns.BloodSugar = bloodSugar
		metric.VitalSigns.HemoglobinLevel = hemoglobinLevel
		metric.VitalSigns.IronLevel = ironLevel
		metric.VitalSigns.Weight = weight

		metrics = append(metrics, &metric)
	}

	if err := rows.Err(); err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "error iterating over health metric rows")
	}

	return metrics, nil
}
