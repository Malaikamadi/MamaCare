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

// EducationContentRepository implements repository.EducationContentRepository interface
type EducationContentRepository struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

// NewEducationContentRepository creates a new education content repository
func NewEducationContentRepository(pool *pgxpool.Pool, logger logger.Logger) repository.EducationContentRepository {
	return &EducationContentRepository{
		pool:   pool,
		logger: logger,
	}
}

// scanEducationContent scans an education content item from a row
func scanEducationContent(row pgx.Row) (*model.EducationContent, error) {
	var content model.EducationContent
	var tags []string

	err := row.Scan(
		&content.ID,
		&content.Title,
		&content.Description,
		&content.Content,
		&content.Category,
		&tags,
		&content.Trimester,
		&content.ContentType,
		&content.MediaURL,
		&content.ThumbnailURL,
		&content.Author,
		&content.ViewCount,
		&content.CreatedAt,
		&content.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errorx.New(errorx.NotFound, "education content not found")
		}
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan education content")
	}

	content.Tags = tags

	return &content, nil
}

// FindByID retrieves education content by ID
func (r *EducationContentRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.EducationContent, error) {
	query := `
		SELECT 
			e.id, 
			e.title, 
			e.description, 
			e.content, 
			e.category, 
			e.tags, 
			e.trimester, 
			e.content_type, 
			e.media_url, 
			e.thumbnail_url, 
			e.author, 
			e.view_count, 
			e.created_at, 
			e.updated_at
		FROM education_content e
		WHERE e.id = $1
	`

	row := database.GetQuerier(ctx, r.pool).QueryRow(ctx, query, id)
	return scanEducationContent(row)
}

// FindByCategory retrieves education content by category
func (r *EducationContentRepository) FindByCategory(ctx context.Context, category model.ContentCategory) ([]*model.EducationContent, error) {
	query := `
		SELECT 
			e.id, 
			e.title, 
			e.description, 
			e.content, 
			e.category, 
			e.tags, 
			e.trimester, 
			e.content_type, 
			e.media_url, 
			e.thumbnail_url, 
			e.author, 
			e.view_count, 
			e.created_at, 
			e.updated_at
		FROM education_content e
		WHERE e.category = $1
		ORDER BY e.created_at DESC
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, category)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query education content by category")
	}
	defer rows.Close()

	return scanEducationContents(rows)
}

// FindByTags retrieves education content by tags
func (r *EducationContentRepository) FindByTags(ctx context.Context, tags []string) ([]*model.EducationContent, error) {
	query := `
		SELECT 
			e.id, 
			e.title, 
			e.description, 
			e.content, 
			e.category, 
			e.tags, 
			e.trimester, 
			e.content_type, 
			e.media_url, 
			e.thumbnail_url, 
			e.author, 
			e.view_count, 
			e.created_at, 
			e.updated_at
		FROM education_content e
		WHERE e.tags && $1
		ORDER BY e.created_at DESC
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, tags)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query education content by tags")
	}
	defer rows.Close()

	return scanEducationContents(rows)
}

// FindByTrimester retrieves education content by trimester
func (r *EducationContentRepository) FindByTrimester(ctx context.Context, trimester int) ([]*model.EducationContent, error) {
	query := `
		SELECT 
			e.id, 
			e.title, 
			e.description, 
			e.content, 
			e.category, 
			e.tags, 
			e.trimester, 
			e.content_type, 
			e.media_url, 
			e.thumbnail_url, 
			e.author, 
			e.view_count, 
			e.created_at, 
			e.updated_at
		FROM education_content e
		WHERE e.trimester = $1
		ORDER BY e.created_at DESC
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, trimester)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query education content by trimester")
	}
	defer rows.Close()

	return scanEducationContents(rows)
}

// FindRecommended retrieves recommended education content for a mother
func (r *EducationContentRepository) FindRecommended(ctx context.Context, motherID uuid.UUID) ([]*model.EducationContent, error) {
	// This query joins with the mothers table to determine the appropriate trimester
	// and selects content that matches the mother's trimester and health conditions
	query := `
		WITH mother_data AS (
			SELECT 
				m.expected_delivery_date,
				m.health_conditions
			FROM mothers m
			WHERE m.id = $1
		),
		trimester_calc AS (
			SELECT 
				CASE
					WHEN (CURRENT_DATE - (expected_delivery_date - INTERVAL '9 months')) < INTERVAL '3 months' THEN 1
					WHEN (CURRENT_DATE - (expected_delivery_date - INTERVAL '9 months')) < INTERVAL '6 months' THEN 2
					ELSE 3
				END AS current_trimester,
				health_conditions
			FROM mother_data
		)
		SELECT 
			e.id, 
			e.title, 
			e.description, 
			e.content, 
			e.category, 
			e.tags, 
			e.trimester, 
			e.content_type, 
			e.media_url, 
			e.thumbnail_url, 
			e.author, 
			e.view_count, 
			e.created_at, 
			e.updated_at
		FROM education_content e, trimester_calc t
		WHERE (e.trimester = t.current_trimester OR e.trimester IS NULL)
		AND (e.tags && t.health_conditions OR e.category = 'general')
		ORDER BY 
			CASE WHEN e.tags && t.health_conditions THEN 0 ELSE 1 END,
			e.view_count DESC
		LIMIT 10
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, motherID)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query recommended education content")
	}
	defer rows.Close()

	return scanEducationContents(rows)
}

// FindMostViewed retrieves the most viewed education content
func (r *EducationContentRepository) FindMostViewed(ctx context.Context, limit int) ([]*model.EducationContent, error) {
	query := `
		SELECT 
			e.id, 
			e.title, 
			e.description, 
			e.content, 
			e.category, 
			e.tags, 
			e.trimester, 
			e.content_type, 
			e.media_url, 
			e.thumbnail_url, 
			e.author, 
			e.view_count, 
			e.created_at, 
			e.updated_at
		FROM education_content e
		ORDER BY e.view_count DESC
		LIMIT $1
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, limit)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query most viewed education content")
	}
	defer rows.Close()

	return scanEducationContents(rows)
}

// FindRecent retrieves the most recent education content
func (r *EducationContentRepository) FindRecent(ctx context.Context, limit int) ([]*model.EducationContent, error) {
	query := `
		SELECT 
			e.id, 
			e.title, 
			e.description, 
			e.content, 
			e.category, 
			e.tags, 
			e.trimester, 
			e.content_type, 
			e.media_url, 
			e.thumbnail_url, 
			e.author, 
			e.view_count, 
			e.created_at, 
			e.updated_at
		FROM education_content e
		ORDER BY e.created_at DESC
		LIMIT $1
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, limit)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query recent education content")
	}
	defer rows.Close()

	return scanEducationContents(rows)
}

// IncrementViewCount increments the view count for education content
func (r *EducationContentRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE education_content
		SET view_count = view_count + 1, updated_at = $1
		WHERE id = $2
	`

	now := time.Now()
	cmdTag, err := database.GetQuerier(ctx, r.pool).Exec(ctx, query, now, id)
	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to increment view count")
	}

	if cmdTag.RowsAffected() == 0 {
		return errorx.New(errorx.NotFound, "education content not found")
	}

	return nil
}

// Save creates or updates education content
func (r *EducationContentRepository) Save(ctx context.Context, content *model.EducationContent) error {
	// Update timestamp for modifications
	content.UpdatedAt = time.Now()

	// Use upsert to handle both insert and update
	query := `
		INSERT INTO education_content (
			id, title, description, content, category, tags, trimester, content_type,
			media_url, thumbnail_url, author, view_count, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			content = EXCLUDED.content,
			category = EXCLUDED.category,
			tags = EXCLUDED.tags,
			trimester = EXCLUDED.trimester,
			content_type = EXCLUDED.content_type,
			media_url = EXCLUDED.media_url,
			thumbnail_url = EXCLUDED.thumbnail_url,
			author = EXCLUDED.author,
			view_count = EXCLUDED.view_count,
			updated_at = EXCLUDED.updated_at
	`

	_, err := database.GetQuerier(ctx, r.pool).Exec(ctx, query,
		content.ID,
		content.Title,
		content.Description,
		content.Content,
		content.Category,
		content.Tags,
		content.Trimester,
		content.ContentType,
		content.MediaURL,
		content.ThumbnailURL,
		content.Author,
		content.ViewCount,
		content.CreatedAt,
		content.UpdatedAt,
	)

	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to save education content")
	}

	return nil
}

// Delete removes education content
func (r *EducationContentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM education_content WHERE id = $1`

	cmdTag, err := database.GetQuerier(ctx, r.pool).Exec(ctx, query, id)
	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to delete education content")
	}

	if cmdTag.RowsAffected() == 0 {
		return errorx.New(errorx.NotFound, "education content not found")
	}

	return nil
}

// scanEducationContents scans multiple education content items from rows
func scanEducationContents(rows pgx.Rows) ([]*model.EducationContent, error) {
	var contents []*model.EducationContent

	for rows.Next() {
		var content model.EducationContent
		var tags []string

		err := rows.Scan(
			&content.ID,
			&content.Title,
			&content.Description,
			&content.Content,
			&content.Category,
			&tags,
			&content.Trimester,
			&content.ContentType,
			&content.MediaURL,
			&content.ThumbnailURL,
			&content.Author,
			&content.ViewCount,
			&content.CreatedAt,
			&content.UpdatedAt,
		)

		if err != nil {
			return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan education content")
		}

		content.Tags = tags
		contents = append(contents, &content)
	}

	if err := rows.Err(); err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "error iterating over education content rows")
	}

	return contents, nil
}
