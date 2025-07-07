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

// MotherRepository implements repository.MotherRepository interface
type MotherRepository struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

// NewMotherRepository creates a new mother repository
func NewMotherRepository(pool *pgxpool.Pool, logger logger.Logger) repository.MotherRepository {
	return &MotherRepository{
		pool:   pool,
		logger: logger,
	}
}

// scanMother scans a mother from a row
func scanMother(row pgx.Row) (*model.Mother, error) {
	var mother model.Mother
	var pregnancyHistoryJSON []byte
	var healthConditions []string

	err := row.Scan(
		&mother.ID,
		&mother.UserID,
		&mother.ExpectedDeliveryDate,
		&mother.BloodType,
		&healthConditions,
		&pregnancyHistoryJSON,
		&mother.RiskLevel,
		&mother.CreatedAt,
		&mother.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errorx.New(errorx.NotFound, "mother not found")
		}
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan mother")
	}

	// Set health conditions
	mother.HealthConditions = healthConditions

	// Unmarshal pregnancy history JSON
	if pregnancyHistoryJSON != nil {
		if err := json.Unmarshal(pregnancyHistoryJSON, &mother.PregnancyHistory); err != nil {
			return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to unmarshal pregnancy history")
		}
	}

	return &mother, nil
}

// FindByID retrieves a mother by ID
func (r *MotherRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Mother, error) {
	query := `
		SELECT 
			m.id, 
			m.user_id, 
			m.expected_delivery_date, 
			m.blood_type, 
			m.health_conditions, 
			m.pregnancy_history, 
			m.risk_level, 
			m.created_at, 
			m.updated_at
		FROM mothers m
		WHERE m.id = $1
	`

	row := database.GetQuerier(ctx, r.pool).QueryRow(ctx, query, id)
	return scanMother(row)
}

// FindByUserID retrieves a mother by user ID
func (r *MotherRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*model.Mother, error) {
	query := `
		SELECT 
			m.id, 
			m.user_id, 
			m.expected_delivery_date, 
			m.blood_type, 
			m.health_conditions, 
			m.pregnancy_history, 
			m.risk_level, 
			m.created_at, 
			m.updated_at
		FROM mothers m
		WHERE m.user_id = $1
	`

	row := database.GetQuerier(ctx, r.pool).QueryRow(ctx, query, userID)
	return scanMother(row)
}

// FindAll retrieves all mothers
func (r *MotherRepository) FindAll(ctx context.Context) ([]*model.Mother, error) {
	query := `
		SELECT 
			m.id, 
			m.user_id, 
			m.expected_delivery_date, 
			m.blood_type, 
			m.health_conditions, 
			m.pregnancy_history, 
			m.risk_level, 
			m.created_at, 
			m.updated_at
		FROM mothers m
		ORDER BY m.expected_delivery_date
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query mothers")
	}
	defer rows.Close()

	return scanMothers(rows)
}

// FindByRiskLevel retrieves mothers by risk level
func (r *MotherRepository) FindByRiskLevel(ctx context.Context, riskLevel model.RiskLevel) ([]*model.Mother, error) {
	query := `
		SELECT 
			m.id, 
			m.user_id, 
			m.expected_delivery_date, 
			m.blood_type, 
			m.health_conditions, 
			m.pregnancy_history, 
			m.risk_level, 
			m.created_at, 
			m.updated_at
		FROM mothers m
		WHERE m.risk_level = $1
		ORDER BY m.expected_delivery_date
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, riskLevel)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query mothers by risk level")
	}
	defer rows.Close()

	return scanMothers(rows)
}

// FindByDueDate retrieves mothers with EDDs in a date range
func (r *MotherRepository) FindByDueDate(ctx context.Context, startDate, endDate time.Time) ([]*model.Mother, error) {
	query := `
		SELECT 
			m.id, 
			m.user_id, 
			m.expected_delivery_date, 
			m.blood_type, 
			m.health_conditions, 
			m.pregnancy_history, 
			m.risk_level, 
			m.created_at, 
			m.updated_at
		FROM mothers m
		WHERE m.expected_delivery_date BETWEEN $1 AND $2
		ORDER BY m.expected_delivery_date
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query mothers by delivery date range")
	}
	defer rows.Close()

	return scanMothers(rows)
}

// FindByDistrict retrieves mothers in a district
func (r *MotherRepository) FindByDistrict(ctx context.Context, district string) ([]*model.Mother, error) {
	query := `
		SELECT 
			m.id, 
			m.user_id, 
			m.expected_delivery_date, 
			m.blood_type, 
			m.health_conditions, 
			m.pregnancy_history, 
			m.risk_level, 
			m.created_at, 
			m.updated_at
		FROM mothers m
		JOIN users u ON m.user_id = u.id
		WHERE u.district = $1
		ORDER BY m.expected_delivery_date
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, district)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query mothers by district")
	}
	defer rows.Close()

	return scanMothers(rows)
}

// FindByHealthcareProvider retrieves mothers assigned to a healthcare provider
func (r *MotherRepository) FindByHealthcareProvider(ctx context.Context, providerID uuid.UUID) ([]*model.Mother, error) {
	query := `
		SELECT DISTINCT
			m.id, 
			m.user_id, 
			m.expected_delivery_date, 
			m.blood_type, 
			m.health_conditions, 
			m.pregnancy_history, 
			m.risk_level, 
			m.created_at, 
			m.updated_at
		FROM mothers m
		JOIN visits v ON m.id = v.mother_id
		WHERE v.chw_id = $1 OR v.clinician_id = $1
		ORDER BY m.expected_delivery_date
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, providerID)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query mothers by healthcare provider")
	}
	defer rows.Close()

	return scanMothers(rows)
}

// Save creates or updates a mother
func (r *MotherRepository) Save(ctx context.Context, mother *model.Mother) error {
	// Update timestamp for modifications
	mother.UpdatedAt = time.Now()

	// Marshal pregnancy history to JSON
	pregnancyHistoryJSON, err := json.Marshal(mother.PregnancyHistory)
	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to marshal pregnancy history")
	}

	// Use upsert to handle both insert and update
	query := `
		INSERT INTO mothers (
			id, user_id, expected_delivery_date, blood_type, health_conditions,
			pregnancy_history, risk_level, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			expected_delivery_date = EXCLUDED.expected_delivery_date,
			blood_type = EXCLUDED.blood_type,
			health_conditions = EXCLUDED.health_conditions,
			pregnancy_history = EXCLUDED.pregnancy_history,
			risk_level = EXCLUDED.risk_level,
			updated_at = EXCLUDED.updated_at
	`

	_, err = database.GetQuerier(ctx, r.pool).Exec(ctx, query,
		mother.ID,
		mother.UserID,
		mother.ExpectedDeliveryDate,
		mother.BloodType,
		mother.HealthConditions,
		pregnancyHistoryJSON,
		mother.RiskLevel,
		mother.CreatedAt,
		mother.UpdatedAt,
	)

	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to save mother")
	}

	return nil
}

// Delete removes a mother
func (r *MotherRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM mothers WHERE id = $1`

	cmdTag, err := database.GetQuerier(ctx, r.pool).Exec(ctx, query, id)
	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to delete mother")
	}

	if cmdTag.RowsAffected() == 0 {
		return errorx.New(errorx.NotFound, "mother not found")
	}

	return nil
}

// CountByDistrict counts mothers by district
func (r *MotherRepository) CountByDistrict(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT u.district, COUNT(m.id) as count
		FROM mothers m
		JOIN users u ON m.user_id = u.id
		GROUP BY u.district
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to count mothers by district")
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var district string
		var count int
		if err := rows.Scan(&district, &count); err != nil {
			return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan district count")
		}
		counts[district] = count
	}

	if err := rows.Err(); err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "error iterating over district count rows")
	}

	return counts, nil
}

// CountByRiskLevel counts mothers by risk level
func (r *MotherRepository) CountByRiskLevel(ctx context.Context) (map[model.RiskLevel]int, error) {
	query := `
		SELECT risk_level, COUNT(id) as count
		FROM mothers
		GROUP BY risk_level
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to count mothers by risk level")
	}
	defer rows.Close()

	counts := make(map[model.RiskLevel]int)
	for rows.Next() {
		var riskLevel model.RiskLevel
		var count int
		if err := rows.Scan(&riskLevel, &count); err != nil {
			return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan risk level count")
		}
		counts[riskLevel] = count
	}

	if err := rows.Err(); err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "error iterating over risk level count rows")
	}

	return counts, nil
}

// scanMothers scans multiple mothers from rows
func scanMothers(rows pgx.Rows) ([]*model.Mother, error) {
	var mothers []*model.Mother

	for rows.Next() {
		var mother model.Mother
		var pregnancyHistoryJSON []byte
		var healthConditions []string

		err := rows.Scan(
			&mother.ID,
			&mother.UserID,
			&mother.ExpectedDeliveryDate,
			&mother.BloodType,
			&healthConditions,
			&pregnancyHistoryJSON,
			&mother.RiskLevel,
			&mother.CreatedAt,
			&mother.UpdatedAt,
		)

		if err != nil {
			return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan mother")
		}

		// Set health conditions
		mother.HealthConditions = healthConditions

		// Unmarshal pregnancy history JSON
		if pregnancyHistoryJSON != nil {
			if err := json.Unmarshal(pregnancyHistoryJSON, &mother.PregnancyHistory); err != nil {
				return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to unmarshal pregnancy history")
			}
		}

		mothers = append(mothers, &mother)
	}

	if err := rows.Err(); err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "error iterating over mother rows")
	}

	return mothers, nil
}
