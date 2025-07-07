package model

import (
	"time"

	"github.com/google/uuid"
)

// ContentCategory represents a category of educational content
type ContentCategory string

const (
	// ContentCategoryGeneral for general pregnancy information
	ContentCategoryGeneral ContentCategory = "general"
	// ContentCategoryNutrition for nutritional guidance
	ContentCategoryNutrition ContentCategory = "nutrition"
	// ContentCategoryExercise for exercise and physical activity
	ContentCategoryExercise ContentCategory = "exercise"
	// ContentCategoryMentalHealth for mental health and wellness
	ContentCategoryMentalHealth ContentCategory = "mental_health"
	// ContentCategoryChildcare for infant care information
	ContentCategoryChildcare ContentCategory = "childcare"
	// ContentCategoryEmergency for emergency and critical situations
	ContentCategoryEmergency ContentCategory = "emergency"
	// ContentCategoryPostpartum for postpartum information
	ContentCategoryPostpartum ContentCategory = "postpartum"
)

// ContentType represents the type/format of educational content
type ContentType string

const (
	// ContentTypeArticle for text articles
	ContentTypeArticle ContentType = "article"
	// ContentTypeVideo for video content
	ContentTypeVideo ContentType = "video"
	// ContentTypeInfographic for visual infographics
	ContentTypeInfographic ContentType = "infographic"
	// ContentTypeAudio for audio content like podcasts
	ContentTypeAudio ContentType = "audio"
	// ContentTypeQuiz for interactive quizzes
	ContentTypeQuiz ContentType = "quiz"
)

// EducationContent represents educational material
type EducationContent struct {
	ID            uuid.UUID      `json:"id"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Content       string         `json:"content"`
	Category      ContentCategory `json:"category"`
	Tags          []string       `json:"tags"`
	Trimester     *int           `json:"trimester,omitempty"` // Which trimester this content applies to (1, 2, 3), null for general content
	ContentType   ContentType    `json:"content_type"`
	MediaURL      string         `json:"media_url,omitempty"`
	ThumbnailURL  string         `json:"thumbnail_url,omitempty"`
	Author        string         `json:"author"`
	ViewCount     int            `json:"view_count"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// NewEducationContent creates new educational content
func NewEducationContent(id uuid.UUID, title, description, content string, category ContentCategory, contentType ContentType) *EducationContent {
	now := time.Now()
	return &EducationContent{
		ID:          id,
		Title:       title,
		Description: description,
		Content:     content,
		Category:    category,
		Tags:        []string{},
		ContentType: contentType,
		ViewCount:   0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// WithTrimester specifies which trimester the content is relevant for
func (e *EducationContent) WithTrimester(trimester int) *EducationContent {
	e.Trimester = &trimester
	return e
}

// WithTags adds tags to the content for better searchability
func (e *EducationContent) WithTags(tags []string) *EducationContent {
	e.Tags = tags
	return e
}

// WithMedia adds media URLs to the content
func (e *EducationContent) WithMedia(mediaURL, thumbnailURL string) *EducationContent {
	e.MediaURL = mediaURL
	e.ThumbnailURL = thumbnailURL
	return e
}

// WithAuthor specifies the author of the content
func (e *EducationContent) WithAuthor(author string) *EducationContent {
	e.Author = author
	return e
}

// IncrementViewCount increases the view count by one
func (e *EducationContent) IncrementViewCount() {
	e.ViewCount++
	e.UpdatedAt = time.Now()
}

// IsRelevantForTrimester checks if content is relevant for a specific trimester
func (e *EducationContent) IsRelevantForTrimester(trimester int) bool {
	return e.Trimester == nil || *e.Trimester == trimester
}

// HasTag checks if the content has a specific tag
func (e *EducationContent) HasTag(tag string) bool {
	for _, t := range e.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// HasAnyTag checks if the content has any of the specified tags
func (e *EducationContent) HasAnyTag(tags []string) bool {
	for _, searchTag := range tags {
		if e.HasTag(searchTag) {
			return true
		}
	}
	return false
}

// IsPopular checks if the content has a significant number of views
func (e *EducationContent) IsPopular() bool {
	return e.ViewCount > 100
}

// TimeSinceCreation returns the duration since the content was created
func (e *EducationContent) TimeSinceCreation() time.Duration {
	return time.Since(e.CreatedAt)
}

// IsRecent checks if the content was created within the specified duration
func (e *EducationContent) IsRecent(duration time.Duration) bool {
	return e.TimeSinceCreation() <= duration
}
