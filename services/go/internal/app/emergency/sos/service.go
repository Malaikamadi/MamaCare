package sos

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/incognito25/mamacare/services/go/internal/domain/model"
	"github.com/incognito25/mamacare/services/go/internal/domain/repository"
	"github.com/incognito25/mamacare/services/go/internal/errorx"
	"github.com/incognito25/mamacare/services/go/internal/logger"
)

// Service provides SOS event coordination functionality
type Service struct {
	sosRepo         repository.SOSRepository
	motherRepo      repository.MotherRepository
	facilityRepo    repository.FacilityRepository
	notificationSvc NotificationService
	chw             CHWLocator
	logger          logger.Logger
}

// NotificationService defines the interface for sending emergency notifications
type NotificationService interface {
	// SendSOSNotification sends a notification for an SOS event
	SendSOSNotification(ctx context.Context, sosEvent *model.SOSEvent, recipientIDs []uuid.UUID) error
	
	// SendSOSUpdateNotification sends a notification for an SOS event update
	SendSOSUpdateNotification(ctx context.Context, sosEvent *model.SOSEvent, recipientIDs []uuid.UUID, update string) error
}

// CHWLocator defines the interface for locating nearby CHWs
type CHWLocator interface {
	// FindNearbyCHWs finds CHWs near a location
	FindNearbyCHWs(ctx context.Context, lat, lng float64, radiusKm float64) ([]uuid.UUID, error)
}

// NewService creates a new SOS service
func NewService(
	sosRepo repository.SOSRepository,
	motherRepo repository.MotherRepository,
	facilityRepo repository.FacilityRepository,
	notificationSvc NotificationService,
	chw CHWLocator,
	logger logger.Logger,
) *Service {
	return &Service{
		sosRepo:         sosRepo,
		motherRepo:      motherRepo,
		facilityRepo:    facilityRepo,
		notificationSvc: notificationSvc,
		chw:             chw,
		logger:          logger,
	}
}

// ReportSOSEvent reports a new SOS event
func (s *Service) ReportSOSEvent(
	ctx context.Context,
	motherID uuid.UUID,
	reportedByID uuid.UUID,
	lat, lng float64,
	nature model.SOSEventNature,
	description string,
) (*model.SOSEvent, error) {
	// Check if mother exists
	mother, err := s.motherRepo.GetByID(ctx, motherID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get mother", logger.FieldsMap{
			"error":     err.Error(),
			"mother_id": motherID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "Mother not found", err)
	}

	// Create SOS event
	sosID := uuid.New()
	sosEvent := model.NewSOSEvent(sosID, motherID, reportedByID, lat, lng, nature)
	sosEvent.WithDescription(description)

	// Set facility if the mother has a primary facility
	if mother.PrimaryFacilityID != nil {
		sosEvent.WithFacility(*mother.PrimaryFacilityID)
	} else {
		// Find nearest facility
		facilities, err := s.facilityRepo.FindNearest(ctx, lat, lng, 1, nil)
		if err != nil {
			s.logger.Error(ctx, "Failed to find nearest facility", logger.FieldsMap{
				"error": err.Error(),
				"lat":   fmt.Sprintf("%f", lat),
				"lng":   fmt.Sprintf("%f", lng),
			})
		} else if len(facilities) > 0 {
			sosEvent.WithFacility(facilities[0].ID)
		}
	}

	// Save SOS event
	if err := s.sosRepo.Create(ctx, sosEvent); err != nil {
		s.logger.Error(ctx, "Failed to create SOS event", logger.FieldsMap{
			"error":     err.Error(),
			"mother_id": motherID.String(),
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to create SOS event", err)
	}

	// Find nearby CHWs to notify
	s.notifyNearbyCHWs(ctx, sosEvent)

	// Notify the facility if assigned
	if sosEvent.FacilityID != nil {
		s.notifyFacility(ctx, sosEvent)
	}

	return sosEvent, nil
}

// GetSOSEvent gets an SOS event by ID
func (s *Service) GetSOSEvent(ctx context.Context, sosID uuid.UUID) (*model.SOSEvent, error) {
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}
	return sosEvent, nil
}

// GetActiveSOSEvents gets all active SOS events
func (s *Service) GetActiveSOSEvents(ctx context.Context) ([]*model.SOSEvent, error) {
	sosEvents, err := s.sosRepo.GetActive(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to get active SOS events", logger.FieldsMap{
			"error": err.Error(),
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to get active SOS events", err)
	}
	return sosEvents, nil
}

// GetSOSEventsByMotherID gets all SOS events for a mother
func (s *Service) GetSOSEventsByMotherID(ctx context.Context, motherID uuid.UUID) ([]*model.SOSEvent, error) {
	sosEvents, err := s.sosRepo.GetByMotherID(ctx, motherID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS events for mother", logger.FieldsMap{
			"error":     err.Error(),
			"mother_id": motherID.String(),
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to get SOS events for mother", err)
	}
	return sosEvents, nil
}

// UpdateSOSEventStatus updates the status of an SOS event
func (s *Service) UpdateSOSEventStatus(ctx context.Context, sosID uuid.UUID, status model.SOSEventStatus) (*model.SOSEvent, error) {
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for status update", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}

	// Update status based on requested status
	switch status {
	case model.SOSEventStatusCancelled:
		sosEvent.Cancel()
	case model.SOSEventStatusResolved:
		if sosEvent.FacilityID == nil {
			return nil, errorx.New(errorx.Validation, "Cannot mark as resolved without a facility")
		}
		sosEvent.Resolve(*sosEvent.FacilityID)
	default:
		// For other statuses, just update the status field
		sosEvent.Status = status
		sosEvent.UpdatedAt = time.Now()
	}

	if err := s.sosRepo.Update(ctx, sosEvent); err != nil {
		s.logger.Error(ctx, "Failed to update SOS event status", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
			"status":  string(status),
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to update SOS event status", err)
	}

	// Notify relevant parties about the status update
	s.notifyStatusUpdate(ctx, sosEvent)

	return sosEvent, nil
}

// AssignFacilityToSOSEvent assigns a facility to an SOS event
func (s *Service) AssignFacilityToSOSEvent(ctx context.Context, sosID, facilityID uuid.UUID) (*model.SOSEvent, error) {
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for facility assignment", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}

	// Verify facility exists
	facility, err := s.facilityRepo.GetByID(ctx, facilityID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get facility for SOS event assignment", logger.FieldsMap{
			"error":       err.Error(),
			"facility_id": facilityID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "Facility not found", err)
	}

	// Assign facility
	sosEvent.WithFacility(facilityID)
	
	if err := s.sosRepo.Update(ctx, sosEvent); err != nil {
		s.logger.Error(ctx, "Failed to assign facility to SOS event", logger.FieldsMap{
			"error":       err.Error(),
			"sos_id":     sosID.String(),
			"facility_id": facilityID.String(),
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to assign facility to SOS event", err)
	}

	// Notify the facility
	s.notifyFacility(ctx, sosEvent)

	s.logger.Info(ctx, "Assigned facility to SOS event", logger.FieldsMap{
		"sos_id":       sosID.String(),
		"facility_id":   facilityID.String(),
		"facility_name": facility.Name,
	})

	return sosEvent, nil
}

// UpdateSOSEventPriority updates the priority of an SOS event
func (s *Service) UpdateSOSEventPriority(ctx context.Context, sosID uuid.UUID, priority int) (*model.SOSEvent, error) {
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for priority update", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}

	// Update priority
	sosEvent.UpdatePriority(priority)
	
	if err := s.sosRepo.Update(ctx, sosEvent); err != nil {
		s.logger.Error(ctx, "Failed to update SOS event priority", logger.FieldsMap{
			"error":    err.Error(),
			"sos_id":  sosID.String(),
			"priority": priority,
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to update SOS event priority", err)
	}

	return sosEvent, nil
}

// GetSOSEventsByFacility gets all SOS events assigned to a facility
func (s *Service) GetSOSEventsByFacility(ctx context.Context, facilityID uuid.UUID) ([]*model.SOSEvent, error) {
	sosEvents, err := s.sosRepo.GetByFacility(ctx, facilityID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS events for facility", logger.FieldsMap{
			"error":       err.Error(),
			"facility_id": facilityID.String(),
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to get SOS events for facility", err)
	}
	return sosEvents, nil
}

// GetSOSEventsInTimeRange gets SOS events within a time range
func (s *Service) GetSOSEventsInTimeRange(ctx context.Context, start, end time.Time) ([]*model.SOSEvent, error) {
	sosEvents, err := s.sosRepo.GetByTimeRange(ctx, start, end)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS events in time range", logger.FieldsMap{
			"error": err.Error(),
			"start": start.Format(time.RFC3339),
			"end":   end.Format(time.RFC3339),
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to get SOS events in time range", err)
	}
	return sosEvents, nil
}

// GetSOSEventsInRadius gets active SOS events within a radius of coordinates
func (s *Service) GetSOSEventsInRadius(ctx context.Context, lat, lng, radiusKm float64) ([]*model.SOSEvent, error) {
	sosEvents, err := s.sosRepo.GetInRadius(ctx, lat, lng, radiusKm)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS events in radius", logger.FieldsMap{
			"error":    err.Error(),
			"lat":      fmt.Sprintf("%f", lat),
			"lng":      fmt.Sprintf("%f", lng),
			"radiusKm": fmt.Sprintf("%f", radiusKm),
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to get SOS events in radius", err)
	}
	return sosEvents, nil
}

// Private methods

// notifyNearbyCHWs notifies nearby CHWs about an SOS event
func (s *Service) notifyNearbyCHWs(ctx context.Context, sosEvent *model.SOSEvent) {
	const radiusKm = 10.0 // Notification radius in kilometers
	
	chwIDs, err := s.chw.FindNearbyCHWs(ctx, sosEvent.Location.Latitude, sosEvent.Location.Longitude, radiusKm)
	if err != nil {
		s.logger.Error(ctx, "Failed to find nearby CHWs for SOS notification", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosEvent.ID.String(),
		})
		return
	}

	if len(chwIDs) == 0 {
		s.logger.Info(ctx, "No nearby CHWs found for SOS notification", logger.FieldsMap{
			"sos_id": sosEvent.ID.String(),
		})
		return
	}

	// Send notification to found CHWs
	if err := s.notificationSvc.SendSOSNotification(ctx, sosEvent, chwIDs); err != nil {
		s.logger.Error(ctx, "Failed to send SOS notification to CHWs", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosEvent.ID.String(),
			"chw_count": len(chwIDs),
		})
	} else {
		s.logger.Info(ctx, "SOS notification sent to CHWs", logger.FieldsMap{
			"sos_id": sosEvent.ID.String(),
			"chw_count": len(chwIDs),
		})
	}
}

// notifyFacility notifies the assigned facility about an SOS event
func (s *Service) notifyFacility(ctx context.Context, sosEvent *model.SOSEvent) {
	if sosEvent.FacilityID == nil {
		return
	}

	facility, err := s.facilityRepo.GetByID(ctx, *sosEvent.FacilityID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get facility for SOS notification", logger.FieldsMap{
			"error":       err.Error(),
			"facility_id": sosEvent.FacilityID.String(),
		})
		return
	}

	// Typically we would get staff/contacts from the facility here
	// For simplicity, just using facility ID as recipient
	recipientIDs := []uuid.UUID{*sosEvent.FacilityID}

	if err := s.notificationSvc.SendSOSNotification(ctx, sosEvent, recipientIDs); err != nil {
		s.logger.Error(ctx, "Failed to send SOS notification to facility", logger.FieldsMap{
			"error":         err.Error(),
			"sos_id":       sosEvent.ID.String(),
			"facility_id":   sosEvent.FacilityID.String(),
			"facility_name": facility.Name,
		})
	} else {
		s.logger.Info(ctx, "SOS notification sent to facility", logger.FieldsMap{
			"sos_id":       sosEvent.ID.String(),
			"facility_id":   sosEvent.FacilityID.String(),
			"facility_name": facility.Name,
		})
	}
}

// notifyStatusUpdate notifies relevant parties about an SOS event status update
func (s *Service) notifyStatusUpdate(ctx context.Context, sosEvent *model.SOSEvent) {
	// Determine who should be notified based on the current status
	recipientIDs := make([]uuid.UUID, 0)

	// Get mother to notify
	if mother, err := s.motherRepo.GetByID(ctx, sosEvent.MotherID); err == nil && mother.UserID != nil {
		recipientIDs = append(recipientIDs, *mother.UserID)
	}

	// Get facility staff to notify
	if sosEvent.FacilityID != nil {
		// In a real implementation, we'd query facility staff here
		// For now, just using facility ID as a placeholder
		recipientIDs = append(recipientIDs, *sosEvent.FacilityID)
	}

	// Get reporter to notify
	recipientIDs = append(recipientIDs, sosEvent.ReportedBy)

	// Prepare status update message
	update := fmt.Sprintf("SOS event status changed to %s", sosEvent.Status)

	// Send notifications
	if err := s.notificationSvc.SendSOSUpdateNotification(ctx, sosEvent, recipientIDs, update); err != nil {
		s.logger.Error(ctx, "Failed to send SOS update notification", logger.FieldsMap{
			"error":         err.Error(),
			"sos_id":       sosEvent.ID.String(),
			"status":        string(sosEvent.Status),
			"recipient_count": len(recipientIDs),
		})
	} else {
		s.logger.Info(ctx, "SOS update notification sent", logger.FieldsMap{
			"sos_id":        sosEvent.ID.String(),
			"status":         string(sosEvent.Status),
			"recipient_count": len(recipientIDs),
		})
	}
}
