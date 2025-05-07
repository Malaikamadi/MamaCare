package alert

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

// Service provides emergency alert functionality
type Service struct {
	sosRepo         repository.SOSRepository
	facilityRepo    repository.FacilityRepository
	contactRepo     repository.ContactRepository
	notifier        AlertNotifier
	logger          logger.Logger
}

// AlertNotifier defines the interface for sending emergency alerts
type AlertNotifier interface {
	// SendFacilityAlert sends an alert to a facility
	SendFacilityAlert(ctx context.Context, sosEvent *model.SOSEvent, facility *model.Facility, alertLevel string) error
	
	// SendRegionalAlert sends an alert to all facilities in a region
	SendRegionalAlert(ctx context.Context, sosEvent *model.SOSEvent, facilities []*model.Facility, alertLevel string) error
	
	// SendContactAlert sends an alert to specific contacts
	SendContactAlert(ctx context.Context, sosEvent *model.SOSEvent, contacts []*model.Contact, alertLevel string) error
}

// AlertLevel represents the severity level of an alert
type AlertLevel string

const (
	// AlertLevelInfo represents an informational alert
	AlertLevelInfo AlertLevel = "info"
	// AlertLevelWarning represents a warning alert
	AlertLevelWarning AlertLevel = "warning"
	// AlertLevelEmergency represents an emergency alert
	AlertLevelEmergency AlertLevel = "emergency"
	// AlertLevelCritical represents a critical alert
	AlertLevelCritical AlertLevel = "critical"
)

// NewService creates a new alert service
func NewService(
	sosRepo repository.SOSRepository,
	facilityRepo repository.FacilityRepository,
	contactRepo repository.ContactRepository,
	notifier AlertNotifier,
	logger logger.Logger,
) *Service {
	return &Service{
		sosRepo:      sosRepo,
		facilityRepo: facilityRepo,
		contactRepo:  contactRepo,
		notifier:     notifier,
		logger:       logger,
	}
}

// AlertNearbyFacilities sends alerts to facilities near an SOS event
func (s *Service) AlertNearbyFacilities(ctx context.Context, sosID uuid.UUID, radiusKm float64) (int, error) {
	// Get SOS event
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for facility alerts", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return 0, errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}
	
	// Find nearby facilities
	facilities, err := s.facilityRepo.FindNearest(
		ctx,
		sosEvent.Location.Latitude,
		sosEvent.Location.Longitude,
		int(radiusKm), // This interface expects an int, but we should modify it to use float64
		nil,           // No filter criteria
	)
	if err != nil {
		s.logger.Error(ctx, "Failed to find nearby facilities for alerts", logger.FieldsMap{
			"error":    err.Error(),
			"sos_id":  sosID.String(),
			"radius_km": fmt.Sprintf("%f", radiusKm),
		})
		return 0, errorx.NewWithCause(errorx.Internal, "Failed to find nearby facilities", err)
	}
	
	if len(facilities) == 0 {
		s.logger.Info(ctx, "No nearby facilities found for alerts", logger.FieldsMap{
			"sos_id":  sosID.String(),
			"radius_km": fmt.Sprintf("%f", radiusKm),
		})
		return 0, nil
	}
	
	// Determine alert level based on SOS event nature
	var alertLevel AlertLevel
	switch sosEvent.Nature {
	case model.SOSEventNatureBleeding:
		alertLevel = AlertLevelCritical
	case model.SOSEventNatureLabor:
		alertLevel = AlertLevelEmergency
	case model.SOSEventNatureAccident:
		alertLevel = AlertLevelEmergency
	default:
		alertLevel = AlertLevelWarning
	}
	
	// Send regional alert to all nearby facilities
	if err := s.notifier.SendRegionalAlert(ctx, sosEvent, facilities, string(alertLevel)); err != nil {
		s.logger.Error(ctx, "Failed to send regional alert to facilities", logger.FieldsMap{
			"error":          err.Error(),
			"sos_id":        sosID.String(),
			"facility_count": len(facilities),
			"alert_level":    string(alertLevel),
		})
		return 0, errorx.NewWithCause(errorx.Internal, "Failed to send alerts to facilities", err)
	}
	
	s.logger.Info(ctx, "Sent alerts to nearby facilities", logger.FieldsMap{
		"sos_id":        sosID.String(),
		"facility_count": len(facilities),
		"radius_km":      fmt.Sprintf("%f", radiusKm),
		"alert_level":    string(alertLevel),
	})
	
	return len(facilities), nil
}

// AlertEmergencyContacts sends alerts to emergency contacts
func (s *Service) AlertEmergencyContacts(ctx context.Context, sosID uuid.UUID) (int, error) {
	// Get SOS event
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for contact alerts", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return 0, errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}
	
	// Get emergency contacts
	contacts, err := s.contactRepo.GetEmergencyContacts(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to get emergency contacts for alerts", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return 0, errorx.NewWithCause(errorx.Internal, "Failed to get emergency contacts", err)
	}
	
	if len(contacts) == 0 {
		s.logger.Info(ctx, "No emergency contacts found for alerts", logger.FieldsMap{
			"sos_id": sosID.String(),
		})
		return 0, nil
	}
	
	// Determine alert level based on SOS event nature
	var alertLevel AlertLevel
	switch sosEvent.Nature {
	case model.SOSEventNatureBleeding:
		alertLevel = AlertLevelCritical
	case model.SOSEventNatureLabor:
		alertLevel = AlertLevelEmergency
	case model.SOSEventNatureAccident:
		alertLevel = AlertLevelEmergency
	default:
		alertLevel = AlertLevelWarning
	}
	
	// Send alert to emergency contacts
	if err := s.notifier.SendContactAlert(ctx, sosEvent, contacts, string(alertLevel)); err != nil {
		s.logger.Error(ctx, "Failed to send alert to emergency contacts", logger.FieldsMap{
			"error":          err.Error(),
			"sos_id":        sosID.String(),
			"contact_count":  len(contacts),
			"alert_level":    string(alertLevel),
		})
		return 0, errorx.NewWithCause(errorx.Internal, "Failed to send alerts to emergency contacts", err)
	}
	
	s.logger.Info(ctx, "Sent alerts to emergency contacts", logger.FieldsMap{
		"sos_id":        sosID.String(),
		"contact_count":  len(contacts),
		"alert_level":    string(alertLevel),
	})
	
	return len(contacts), nil
}

// AlertFacility sends an alert to a specific facility
func (s *Service) AlertFacility(ctx context.Context, sosID uuid.UUID, facilityID uuid.UUID) error {
	// Get SOS event
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for facility alert", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}
	
	// Get facility
	facility, err := s.facilityRepo.GetByID(ctx, facilityID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get facility for alert", logger.FieldsMap{
			"error":       err.Error(),
			"facility_id": facilityID.String(),
		})
		return errorx.NewWithCause(errorx.NotFound, "Facility not found", err)
	}
	
	// Determine alert level based on SOS event nature
	var alertLevel AlertLevel
	switch sosEvent.Nature {
	case model.SOSEventNatureBleeding:
		alertLevel = AlertLevelCritical
	case model.SOSEventNatureLabor:
		alertLevel = AlertLevelEmergency
	case model.SOSEventNatureAccident:
		alertLevel = AlertLevelEmergency
	default:
		alertLevel = AlertLevelWarning
	}
	
	// Send alert to facility
	if err := s.notifier.SendFacilityAlert(ctx, sosEvent, facility, string(alertLevel)); err != nil {
		s.logger.Error(ctx, "Failed to send alert to facility", logger.FieldsMap{
			"error":        err.Error(),
			"sos_id":      sosID.String(),
			"facility_id":  facilityID.String(),
			"facility_name": facility.Name,
			"alert_level":  string(alertLevel),
		})
		return errorx.NewWithCause(errorx.Internal, "Failed to send alert to facility", err)
	}
	
	s.logger.Info(ctx, "Sent alert to facility", logger.FieldsMap{
		"sos_id":      sosID.String(),
		"facility_id":  facilityID.String(),
		"facility_name": facility.Name,
		"alert_level":  string(alertLevel),
	})
	
	return nil
}

// SendStatusAlert sends an alert about a status change in an SOS event
func (s *Service) SendStatusAlert(ctx context.Context, sosID uuid.UUID) error {
	// Get SOS event
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for status alert", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}
	
	// For status changes, use lower alert level
	alertLevel := AlertLevelInfo
	
	// If the SOS event has a facility, alert that facility
	if sosEvent.FacilityID != nil {
		facility, err := s.facilityRepo.GetByID(ctx, *sosEvent.FacilityID)
		if err != nil {
			s.logger.Error(ctx, "Failed to get facility for status alert", logger.FieldsMap{
				"error":       err.Error(),
				"facility_id": sosEvent.FacilityID.String(),
			})
		} else {
			if err := s.notifier.SendFacilityAlert(ctx, sosEvent, facility, string(alertLevel)); err != nil {
				s.logger.Error(ctx, "Failed to send status alert to facility", logger.FieldsMap{
					"error":        err.Error(),
					"sos_id":      sosID.String(),
					"facility_id":  sosEvent.FacilityID.String(),
					"facility_name": facility.Name,
				})
				// Continue despite error
			} else {
				s.logger.Info(ctx, "Sent status alert to facility", logger.FieldsMap{
					"sos_id":      sosID.String(),
					"facility_id":  sosEvent.FacilityID.String(),
					"facility_name": facility.Name,
					"status":       string(sosEvent.Status),
				})
			}
		}
	}
	
	// Get nearby facilities (within smaller radius)
	const radiusKm = 5.0
	facilities, err := s.facilityRepo.FindNearest(
		ctx,
		sosEvent.Location.Latitude,
		sosEvent.Location.Longitude,
		int(radiusKm),
		nil,
	)
	if err != nil {
		s.logger.Error(ctx, "Failed to find nearby facilities for status alert", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		// Continue despite error
	} else if len(facilities) > 0 {
		if err := s.notifier.SendRegionalAlert(ctx, sosEvent, facilities, string(alertLevel)); err != nil {
			s.logger.Error(ctx, "Failed to send status alert to nearby facilities", logger.FieldsMap{
				"error":          err.Error(),
				"sos_id":        sosID.String(),
				"facility_count": len(facilities),
			})
			// Continue despite error
		} else {
			s.logger.Info(ctx, "Sent status alert to nearby facilities", logger.FieldsMap{
				"sos_id":        sosID.String(),
				"facility_count": len(facilities),
				"status":         string(sosEvent.Status),
			})
		}
	}
	
	return nil
}

// SendCustomAlert sends a custom alert with a specific message
func (s *Service) SendCustomAlert(
	ctx context.Context,
	sosID uuid.UUID,
	message string,
	alertLevel AlertLevel,
	radiusKm float64,
) (int, error) {
	// Get SOS event
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for custom alert", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return 0, errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}
	
	// Create a copy of the SOS event with custom message
	customSOS := *sosEvent
	customSOS.Description = message
	
	// Find nearby facilities
	facilities, err := s.facilityRepo.FindNearest(
		ctx,
		sosEvent.Location.Latitude,
		sosEvent.Location.Longitude,
		int(radiusKm),
		nil,
	)
	if err != nil {
		s.logger.Error(ctx, "Failed to find nearby facilities for custom alert", logger.FieldsMap{
			"error":    err.Error(),
			"sos_id":  sosID.String(),
			"radius_km": fmt.Sprintf("%f", radiusKm),
		})
		return 0, errorx.NewWithCause(errorx.Internal, "Failed to find nearby facilities", err)
	}
	
	if len(facilities) == 0 {
		s.logger.Info(ctx, "No nearby facilities found for custom alert", logger.FieldsMap{
			"sos_id":  sosID.String(),
			"radius_km": fmt.Sprintf("%f", radiusKm),
		})
		return 0, nil
	}
	
	// Send regional alert to all nearby facilities
	if err := s.notifier.SendRegionalAlert(ctx, &customSOS, facilities, string(alertLevel)); err != nil {
		s.logger.Error(ctx, "Failed to send custom alert to facilities", logger.FieldsMap{
			"error":          err.Error(),
			"sos_id":        sosID.String(),
			"facility_count": len(facilities),
			"alert_level":    string(alertLevel),
		})
		return 0, errorx.NewWithCause(errorx.Internal, "Failed to send custom alert to facilities", err)
	}
	
	s.logger.Info(ctx, "Sent custom alert to facilities", logger.FieldsMap{
		"sos_id":        sosID.String(),
		"facility_count": len(facilities),
		"radius_km":      fmt.Sprintf("%f", radiusKm),
		"alert_level":    string(alertLevel),
		"message":        message,
	})
	
	return len(facilities), nil
}
