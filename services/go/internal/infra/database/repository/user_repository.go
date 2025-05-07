package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/domain/repository"
	"github.com/mamacare/services/internal/infra/database"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// UserRepository implements repository.UserRepository interface
type UserRepository struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(pool *pgxpool.Pool, logger logger.Logger) repository.UserRepository {
	return &UserRepository{
		pool:   pool,
		logger: logger,
	}
}

// scanUser scans a user from a row
func scanUser(row pgx.Row) (*model.User, error) {
	var user model.User
	var facilityID *uuid.UUID

	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.Role,
		&user.District,
		&facilityID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errorx.New(errorx.NotFound, "user not found")
		}
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan user")
	}

	user.FacilityID = facilityID
	return &user, nil
}

// FindByID retrieves a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, name, email, phone_number, role, district, facility_id, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	row := database.GetQuerier(ctx, r.pool).QueryRow(ctx, query, id)
	return scanUser(row)
}

// FindByEmail retrieves a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, name, email, phone_number, role, district, facility_id, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	row := database.GetQuerier(ctx, r.pool).QueryRow(ctx, query, email)
	return scanUser(row)
}

// FindByPhoneNumber retrieves a user by phone number
func (r *UserRepository) FindByPhoneNumber(ctx context.Context, phoneNumber string) (*model.User, error) {
	query := `
		SELECT id, name, email, phone_number, role, district, facility_id, created_at, updated_at
		FROM users
		WHERE phone_number = $1
	`

	row := database.GetQuerier(ctx, r.pool).QueryRow(ctx, query, phoneNumber)
	return scanUser(row)
}

// FindByFirebaseUID retrieves a user by Firebase UID (using custom query)
func (r *UserRepository) FindByFirebaseUID(ctx context.Context, firebaseUID string) (*model.User, error) {
	// Assuming we have a column for firebase_uid or similar mechanism
	// This might need to be adjusted based on how Firebase integration is handled
	query := `
		SELECT id, name, email, phone_number, role, district, facility_id, created_at, updated_at
		FROM users
		WHERE firebase_uid = $1
	`

	row := database.GetQuerier(ctx, r.pool).QueryRow(ctx, query, firebaseUID)
	return scanUser(row)
}

// Save creates or updates a user
func (r *UserRepository) Save(ctx context.Context, user *model.User) error {
	// Update timestamp for modifications
	user.UpdatedAt = time.Now()

	// Use upsert to handle both insert and update
	query := `
		INSERT INTO users (
			id, name, email, phone_number, role, district, facility_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			email = EXCLUDED.email,
			phone_number = EXCLUDED.phone_number,
			role = EXCLUDED.role,
			district = EXCLUDED.district,
			facility_id = EXCLUDED.facility_id,
			updated_at = EXCLUDED.updated_at
	`

	_, err := database.GetQuerier(ctx, r.pool).Exec(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.Phone,
		user.Role,
		user.District,
		user.FacilityID,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" { // Unique violation
				return errorx.New(errorx.BadRequest, "user with this email already exists")
			}
		}
		return errorx.Wrap(err, errorx.InternalServerError, "failed to save user")
	}

	return nil
}

// Delete removes a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	cmdTag, err := database.GetQuerier(ctx, r.pool).Exec(ctx, query, id)
	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to delete user")
	}

	if cmdTag.RowsAffected() == 0 {
		return errorx.New(errorx.NotFound, "user not found")
	}

	return nil
}

// FindHealthcareProvidersByDistrict retrieves healthcare providers by district
func (r *UserRepository) FindHealthcareProvidersByDistrict(ctx context.Context, district string) ([]*model.User, error) {
	query := `
		SELECT id, name, email, phone_number, role, district, facility_id, created_at, updated_at
		FROM users
		WHERE district = $1 AND role IN ('chw', 'clinician')
		ORDER BY name
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, district)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query healthcare providers")
	}
	defer rows.Close()

	return scanUsers(rows)
}

// FindByRole retrieves users by role
func (r *UserRepository) FindByRole(ctx context.Context, role model.UserRole) ([]*model.User, error) {
	query := `
		SELECT id, name, email, phone_number, role, district, facility_id, created_at, updated_at
		FROM users
		WHERE role = $1
		ORDER BY name
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, role)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, fmt.Sprintf("failed to query users by role: %s", role))
	}
	defer rows.Close()

	return scanUsers(rows)
}

// FindByFacility retrieves users associated with a facility
func (r *UserRepository) FindByFacility(ctx context.Context, facilityID uuid.UUID) ([]*model.User, error) {
	query := `
		SELECT 
			u.id, 
			u.name, 
			u.email, 
			u.phone_number, 
			u.role, 
			u.district, 
			u.facility_id, 
			u.created_at, 
			u.updated_at
		FROM users u
		WHERE u.facility_id = $1
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, facilityID)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query users by facility")
	}
	defer rows.Close()

	return scanUsers(rows)
}

// FindByDistrict retrieves users by district
func (r *UserRepository) FindByDistrict(ctx context.Context, district string) ([]*model.User, error) {
	query := `
		SELECT 
			u.id, 
			u.name, 
			u.email, 
			u.phone_number, 
			u.role, 
			u.district, 
			u.facility_id, 
			u.created_at, 
			u.updated_at
		FROM users u
		WHERE u.district = $1
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, district)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query users by district")
	}
	defer rows.Close()

	return scanUsers(rows)
}

// FindAll retrieves all users
func (r *UserRepository) FindAll(ctx context.Context) ([]*model.User, error) {
	query := `
		SELECT 
			u.id, 
			u.name, 
			u.email, 
			u.phone_number, 
			u.role, 
			u.district, 
			u.facility_id, 
			u.created_at, 
			u.updated_at
		FROM users u
		ORDER BY u.name
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query all users")
	}
	defer rows.Close()

	return scanUsers(rows)
}

// CountByRole counts users by role
func (r *UserRepository) CountByRole(ctx context.Context) (map[model.UserRole]int, error) {
	query := `
		SELECT 
			u.role, 
			COUNT(*) as count
		FROM users u
		GROUP BY u.role
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to count users by role")
	}
	defer rows.Close()

	result := make(map[model.UserRole]int)
	for rows.Next() {
		var role model.UserRole
		var count int

		if err := rows.Scan(&role, &count); err != nil {
			return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan user count")
		}

		result[role] = count
	}

	if err := rows.Err(); err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "error iterating user count rows")
	}

	return result, nil
}

// scanUsers scans multiple users from rows
func scanUsers(rows pgx.Rows) ([]*model.User, error) {
	var users []*model.User

	for rows.Next() {
		var user model.User
		var facilityID *uuid.UUID

		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Phone,
			&user.Role,
			&user.District,
			&facilityID,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan user")
		}

		user.FacilityID = facilityID
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "error iterating over user rows")
	}

	return users, nil
}
