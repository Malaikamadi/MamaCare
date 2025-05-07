package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/domain/model"
)

// UserRepository defines methods for user data access
type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByPhoneNumber(ctx context.Context, phoneNumber string) (*model.User, error)
	FindByRole(ctx context.Context, role model.UserRole) ([]*model.User, error)
	FindByDistrict(ctx context.Context, district string) ([]*model.User, error)
	FindByFacility(ctx context.Context, facilityID uuid.UUID) ([]*model.User, error)
	FindAll(ctx context.Context) ([]*model.User, error)
	Save(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByRole(ctx context.Context) (map[model.UserRole]int, error)
}

// MotherRepository defines methods for mother data access
type MotherRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Mother, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*model.Mother, error)
	FindByDueDate(ctx context.Context, startDate, endDate time.Time) ([]*model.Mother, error)
	FindByDistrict(ctx context.Context, district string) ([]*model.Mother, error)
	FindByRiskLevel(ctx context.Context, riskLevel model.RiskLevel) ([]*model.Mother, error)
	FindAll(ctx context.Context) ([]*model.Mother, error)
	Save(ctx context.Context, mother *model.Mother) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByRiskLevel(ctx context.Context) (map[model.RiskLevel]int, error)
}

// FacilityRepository defines methods for healthcare facility data access
type FacilityRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.HealthcareFacility, error)
	FindByName(ctx context.Context, name string) (*model.HealthcareFacility, error)
	FindByDistrict(ctx context.Context, district string) ([]*model.HealthcareFacility, error)
	FindByType(ctx context.Context, facilityType model.FacilityType) ([]*model.HealthcareFacility, error)
	FindNearby(ctx context.Context, lat, lng float64, radiusKm float64) ([]*model.HealthcareFacility, error)
	FindAll(ctx context.Context) ([]*model.HealthcareFacility, error)
	Save(ctx context.Context, facility *model.HealthcareFacility) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByType(ctx context.Context) (map[model.FacilityType]int, error)
}

// VisitRepository defines methods for visit data access
type VisitRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Visit, error)
	FindByMother(ctx context.Context, motherID uuid.UUID) ([]*model.Visit, error)
	FindByFacility(ctx context.Context, facilityID uuid.UUID) ([]*model.Visit, error)
	FindByCHW(ctx context.Context, chwID uuid.UUID) ([]*model.Visit, error)
	FindByClinician(ctx context.Context, clinicianID uuid.UUID) ([]*model.Visit, error)
	FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*model.Visit, error)
	FindByStatus(ctx context.Context, status model.VisitStatus) ([]*model.Visit, error)
	FindUpcoming(ctx context.Context, motherID uuid.UUID) ([]*model.Visit, error)
	FindUpcomingByFacility(ctx context.Context, facilityID uuid.UUID) ([]*model.Visit, error)
	Save(ctx context.Context, visit *model.Visit) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByStatusAndDate(ctx context.Context, date time.Time) (map[model.VisitStatus]int, error)
}

// SOSRepository defines methods for SOS event data access
type SOSRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.SOSEvent, error)
	FindByMother(ctx context.Context, motherID uuid.UUID) ([]*model.SOSEvent, error)
	FindActive(ctx context.Context) ([]*model.SOSEvent, error)
	FindByStatus(ctx context.Context, status model.SOSEventStatus) ([]*model.SOSEvent, error)
	FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*model.SOSEvent, error)
	FindByDistrict(ctx context.Context, district string) ([]*model.SOSEvent, error)
	FindByNature(ctx context.Context, nature model.SOSEventNature) ([]*model.SOSEvent, error)
	FindNearby(ctx context.Context, lat, lng float64, radiusKm float64) ([]*model.SOSEvent, error)
	Save(ctx context.Context, sosEvent *model.SOSEvent) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByStatusAndDate(ctx context.Context, date time.Time) (map[model.SOSEventStatus]int, error)
}

// HealthMetricRepository defines methods for health metric data access
type HealthMetricRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.HealthMetric, error)
	FindByMother(ctx context.Context, motherID uuid.UUID) ([]*model.HealthMetric, error)
	FindByVisit(ctx context.Context, visitID uuid.UUID) ([]*model.HealthMetric, error)
	FindByDateRange(ctx context.Context, motherID uuid.UUID, startDate, endDate time.Time) ([]*model.HealthMetric, error)
	FindLatest(ctx context.Context, motherID uuid.UUID) (*model.HealthMetric, error)
	FindAbnormalBloodPressure(ctx context.Context, threshold float64) ([]*model.HealthMetric, error)
	GetWeightTrend(ctx context.Context, motherID uuid.UUID, months int) ([]model.WeightRecord, error)
	GetVitalSignsAverage(ctx context.Context, motherID uuid.UUID, days int) (*model.VitalSigns, error)
	Save(ctx context.Context, metric *model.HealthMetric) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// NotificationRepository defines methods for notification data access
type NotificationRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Notification, error)
	FindByUser(ctx context.Context, userID uuid.UUID) ([]*model.Notification, error)
	FindUnreadByUser(ctx context.Context, userID uuid.UUID) ([]*model.Notification, error)
	FindByType(ctx context.Context, notificationType model.NotificationType) ([]*model.Notification, error)
	FindByPriority(ctx context.Context, priority model.NotificationPriority) ([]*model.Notification, error)
	FindByDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*model.Notification, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	Save(ctx context.Context, notification *model.Notification) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteAllByUser(ctx context.Context, userID uuid.UUID) error
}

// EducationContentRepository defines methods for educational content data access
type EducationContentRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.EducationContent, error)
	FindByCategory(ctx context.Context, category model.ContentCategory) ([]*model.EducationContent, error)
	FindByTags(ctx context.Context, tags []string) ([]*model.EducationContent, error)
	FindByTrimester(ctx context.Context, trimester int) ([]*model.EducationContent, error)
	FindRecommended(ctx context.Context, motherID uuid.UUID) ([]*model.EducationContent, error)
	FindMostViewed(ctx context.Context, limit int) ([]*model.EducationContent, error)
	FindRecent(ctx context.Context, limit int) ([]*model.EducationContent, error)
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
	Save(ctx context.Context, content *model.EducationContent) error
	Delete(ctx context.Context, id uuid.UUID) error
}

