package tracking

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

// Service provides real-time emergency tracking functionality
type Service struct {
	sosRepo         repository.SOSRepository
	ambulanceRepo   repository.AmbulanceRepository
	facilityRepo    repository.FacilityRepository
	motherRepo      repository.MotherRepository
	routingEngine   RoutingEngine
	trackingNotifier TrackingNotifier
	logger          logger.Logger
}

// RoutingEngine defines the interface for calculating routes and positions
type RoutingEngine interface {
	// CalculateRoute calculates a route between two points
	CalculateRoute(ctx context.Context, fromLat, fromLng, toLat, toLng float64) (*model.Route, error)
	
	// EstimateTimeOfArrival estimates time of arrival between two points
	EstimateTimeOfArrival(ctx context.Context, fromLat, fromLng, toLat, toLng float64) (time.Duration, error)
}

// TrackingNotifier defines the interface for sending tracking notifications
type TrackingNotifier interface {
	// NotifyStatusUpdate notifies about a tracking status update
	NotifyStatusUpdate(ctx context.Context, sosEvent *model.SOSEvent, status string, eta *time.Time) error
	
	// NotifyArrival notifies about an ambulance arrival
	NotifyArrival(ctx context.Context, sosEvent *model.SOSEvent, ambulanceID uuid.UUID) error
	
	// NotifyDelay notifies about a delay in estimated arrival time
	NotifyDelay(ctx context.Context, sosEvent *model.SOSEvent, newETA time.Time, delayMinutes int) error
}

// TrackingUpdate represents a tracking update for an emergency
type TrackingUpdate struct {
	ID          uuid.UUID      `json:"id"`
	SOSEventID  uuid.UUID      `json:"sos_event_id"`
	UpdateType  string         `json:"update_type"`
	Description string         `json:"description"`
	Timestamp   time.Time      `json:"timestamp"`
	Location    *model.Location `json:"location,omitempty"`
	ETA         *time.Time     `json:"eta,omitempty"`
}

// NewService creates a new tracking service
func NewService(
	sosRepo repository.SOSRepository,
	ambulanceRepo repository.AmbulanceRepository,
	facilityRepo repository.FacilityRepository,
	motherRepo repository.MotherRepository,
	routingEngine RoutingEngine,
	trackingNotifier TrackingNotifier,
	logger logger.Logger,
) *Service {
	return &Service{
		sosRepo:         sosRepo,
		ambulanceRepo:   ambulanceRepo,
		facilityRepo:    facilityRepo,
		motherRepo:      motherRepo,
		routingEngine:   routingEngine,
		trackingNotifier: trackingNotifier,
		logger:          logger,
	}
}

// GetEmergencyStatus gets the current status of an emergency SOS event
func (s *Service) GetEmergencyStatus(ctx context.Context, sosID uuid.UUID) (*model.SOSEvent, error) {
	// Get SOS event
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event status", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}
	
	// If an ambulance is assigned, refresh ETA based on current location
	if sosEvent.Status == model.SOSEventStatusDispatched && sosEvent.AmbulanceID != nil {
		s.refreshETA(ctx, sosEvent)
	}
	
	return sosEvent, nil
}

// GetAmbulanceLocation gets the current location of an ambulance
func (s *Service) GetAmbulanceLocation(ctx context.Context, ambulanceID uuid.UUID) (*model.Location, error) {
	// Get ambulance
	ambulance, err := s.ambulanceRepo.GetByID(ctx, ambulanceID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get ambulance for location", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": ambulanceID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "Ambulance not found", err)
	}
	
	if ambulance.Location == nil {
		return nil, errorx.New(errorx.NotFound, "Ambulance location not available")
	}
	
	return ambulance.Location, nil
}

// TrackRoute calculates the current route for an emergency
func (s *Service) TrackRoute(ctx context.Context, sosID uuid.UUID) (*model.Route, error) {
	// Get SOS event
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for route tracking", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}
	
	// Check if ambulance is assigned
	if sosEvent.AmbulanceID == nil {
		return nil, errorx.New(errorx.Validation, "No ambulance assigned to this SOS event")
	}
	
	// Get ambulance
	ambulance, err := s.ambulanceRepo.GetByID(ctx, *sosEvent.AmbulanceID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get ambulance for route tracking", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": sosEvent.AmbulanceID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "Ambulance not found", err)
	}
	
	// Check if ambulance location is available
	if ambulance.Location == nil {
		return nil, errorx.New(errorx.NotFound, "Ambulance location not available")
	}
	
	// Calculate route from ambulance to emergency location
	route, err := s.routingEngine.CalculateRoute(
		ctx,
		ambulance.Location.Latitude,
		ambulance.Location.Longitude,
		sosEvent.Location.Latitude,
		sosEvent.Location.Longitude,
	)
	if err != nil {
		s.logger.Error(ctx, "Failed to calculate route for tracking", logger.FieldsMap{
			"error":        err.Error(),
			"sos_id":      sosID.String(),
			"ambulance_id": ambulance.ID.String(),
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to calculate route", err)
	}
	
	return route, nil
}

// UpdateETABasedOnTraffic updates the ETA based on current traffic conditions
func (s *Service) UpdateETABasedOnTraffic(ctx context.Context, sosID uuid.UUID, trafficDelayMinutes int) (*time.Time, error) {
	// Get SOS event
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for ETA update", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}
	
	// Check if ETA exists
	if sosEvent.ETA == nil {
		return nil, errorx.New(errorx.Validation, "No existing ETA to update")
	}
	
	// Calculate new ETA
	newETA := sosEvent.ETA.Add(time.Duration(trafficDelayMinutes) * time.Minute)
	
	// Update SOS event with new ETA
	sosEvent.UpdateETA(newETA)
	if err := s.sosRepo.Update(ctx, sosEvent); err != nil {
		s.logger.Error(ctx, "Failed to update SOS event with new ETA", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to update ETA", err)
	}
	
	// Calculate significant delay (more than 5 minutes)
	if trafficDelayMinutes > 5 {
		// Notify about the delay
		if err := s.trackingNotifier.NotifyDelay(ctx, sosEvent, newETA, trafficDelayMinutes); err != nil {
			s.logger.Error(ctx, "Failed to send delay notification", logger.FieldsMap{
				"error":   err.Error(),
				"sos_id": sosID.String(),
				"delay":   trafficDelayMinutes,
			})
			// Continue despite notification error
		}
	}
	
	s.logger.Info(ctx, "Updated ETA based on traffic", logger.FieldsMap{
		"sos_id": sosID.String(),
		"delay":   trafficDelayMinutes,
		"new_eta": newETA.Format(time.RFC3339),
	})
	
	return &newETA, nil
}

// RecordAmbulanceArrival records that an ambulance has arrived at the emergency location
func (s *Service) RecordAmbulanceArrival(ctx context.Context, sosID uuid.UUID) error {
	// Get SOS event
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for arrival recording", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}
	
	// Check if ambulance is assigned
	if sosEvent.AmbulanceID == nil {
		return errorx.New(errorx.Validation, "No ambulance assigned to this SOS event")
	}
	
	// Get ambulance
	ambulance, err := s.ambulanceRepo.GetByID(ctx, *sosEvent.AmbulanceID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get ambulance for arrival recording", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": sosEvent.AmbulanceID.String(),
		})
		return errorx.NewWithCause(errorx.NotFound, "Ambulance not found", err)
	}
	
	// Update ambulance status to arrived
	ambulance.MarkArrived()
	if err := s.ambulanceRepo.Update(ctx, ambulance); err != nil {
		s.logger.Error(ctx, "Failed to update ambulance status for arrival", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": ambulance.ID.String(),
		})
		return errorx.NewWithCause(errorx.Internal, "Failed to update ambulance status", err)
	}
	
	// Send arrival notification
	if err := s.trackingNotifier.NotifyArrival(ctx, sosEvent, *sosEvent.AmbulanceID); err != nil {
		s.logger.Error(ctx, "Failed to send arrival notification", logger.FieldsMap{
			"error":        err.Error(),
			"sos_id":      sosID.String(),
			"ambulance_id": ambulance.ID.String(),
		})
		// Continue despite notification error
	}
	
	s.logger.Info(ctx, "Recorded ambulance arrival", logger.FieldsMap{
		"sos_id":      sosID.String(),
		"ambulance_id": ambulance.ID.String(),
	})
	
	return nil
}

// RecordEmergencyStatusUpdate records a status update for an emergency
func (s *Service) RecordEmergencyStatusUpdate(ctx context.Context, sosID uuid.UUID, updateType, description string) (*TrackingUpdate, error) {
	// Get SOS event
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for status update", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}
	
	// Create tracking update
	update := &TrackingUpdate{
		ID:          uuid.New(),
		SOSEventID:  sosID,
		UpdateType:  updateType,
		Description: description,
		Timestamp:   time.Now(),
		ETA:         sosEvent.ETA,
	}
	
	// If ambulance is assigned and has location, include it
	if sosEvent.AmbulanceID != nil {
		ambulance, err := s.ambulanceRepo.GetByID(ctx, *sosEvent.AmbulanceID)
		if err == nil && ambulance.Location != nil {
			update.Location = ambulance.Location
		}
	}
	
	// Send status update notification
	if err := s.trackingNotifier.NotifyStatusUpdate(ctx, sosEvent, updateType, sosEvent.ETA); err != nil {
		s.logger.Error(ctx, "Failed to send status update notification", logger.FieldsMap{
			"error":       err.Error(),
			"sos_id":     sosID.String(),
			"update_type": updateType,
		})
		// Continue despite notification error
	}
	
	s.logger.Info(ctx, "Recorded emergency status update", logger.FieldsMap{
		"sos_id":     sosID.String(),
		"update_type": updateType,
		"description": description,
	})
	
	return update, nil
}

// GetEstimatedTimeOfArrival gets the ETA for an emergency
func (s *Service) GetEstimatedTimeOfArrival(ctx context.Context, sosID uuid.UUID) (*time.Time, error) {
	// Get SOS event
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for ETA", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}
	
	// Check if ambulance is assigned
	if sosEvent.AmbulanceID == nil {
		return nil, errorx.New(errorx.Validation, "No ambulance assigned to this SOS event")
	}
	
	// Refresh ETA
	s.refreshETA(ctx, sosEvent)
	
	// Return current ETA
	if sosEvent.ETA == nil {
		return nil, errorx.New(errorx.NotFound, "ETA not available")
	}
	
	return sosEvent.ETA, nil
}

// GetETAMinutes gets the ETA in minutes for an emergency
func (s *Service) GetETAMinutes(ctx context.Context, sosID uuid.UUID) (int, error) {
	// Get SOS event
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for ETA minutes", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return -1, errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}
	
	// Calculate minutes until ETA
	minutes := sosEvent.CalculateETAMinutes(time.Now())
	
	return minutes, nil
}

// Private methods

// refreshETA refreshes the ETA for an SOS event based on current ambulance location
func (s *Service) refreshETA(ctx context.Context, sosEvent *model.SOSEvent) {
	// Return early if no ambulance is assigned
	if sosEvent.AmbulanceID == nil {
		return
	}
	
	// Get ambulance
	ambulance, err := s.ambulanceRepo.GetByID(ctx, *sosEvent.AmbulanceID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get ambulance for ETA refresh", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": sosEvent.AmbulanceID.String(),
		})
		return
	}
	
	// Return early if ambulance has no location
	if ambulance.Location == nil {
		return
	}
	
	// Calculate new ETA
	duration, err := s.routingEngine.EstimateTimeOfArrival(
		ctx,
		ambulance.Location.Latitude,
		ambulance.Location.Longitude,
		sosEvent.Location.Latitude,
		sosEvent.Location.Longitude,
	)
	if err != nil {
		s.logger.Error(ctx, "Failed to estimate time of arrival for refresh", logger.FieldsMap{
			"error":        err.Error(),
			"sos_id":      sosEvent.ID.String(),
			"ambulance_id": ambulance.ID.String(),
		})
		return
	}
	
	// Calculate new ETA
	newETA := time.Now().Add(duration)
	
	// Check if significant change in ETA
	var significantChange bool
	if sosEvent.ETA != nil {
		diff := newETA.Sub(*sosEvent.ETA)
		if diff < 0 {
			diff = -diff
		}
		// Consider a change of more than 5 minutes significant
		significantChange = diff > 5*time.Minute
	} else {
		significantChange = true
	}
	
	// Update SOS event with new ETA
	oldETA := sosEvent.ETA
	sosEvent.UpdateETA(newETA)
	if err := s.sosRepo.Update(ctx, sosEvent); err != nil {
		s.logger.Error(ctx, "Failed to update SOS event with refreshed ETA", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosEvent.ID.String(),
		})
		return
	}
	
	// Log the refreshed ETA
	if significantChange {
		var oldETAStr string
		if oldETA != nil {
			oldETAStr = oldETA.Format(time.RFC3339)
		} else {
			oldETAStr = "N/A"
		}
		
		s.logger.Info(ctx, "Significant ETA change detected", logger.FieldsMap{
			"sos_id":  sosEvent.ID.String(),
			"old_eta": oldETAStr,
			"new_eta": newETA.Format(time.RFC3339),
		})
		
		// If there's a significant delay, notify
		if oldETA != nil && newETA.After(*oldETA) {
			delayMinutes := int(newETA.Sub(*oldETA).Minutes())
			if delayMinutes > 5 {
				if err := s.trackingNotifier.NotifyDelay(ctx, sosEvent, newETA, delayMinutes); err != nil {
					s.logger.Error(ctx, "Failed to send delay notification on refresh", logger.FieldsMap{
						"error":   err.Error(),
						"sos_id": sosEvent.ID.String(),
						"delay":   fmt.Sprintf("%d", delayMinutes),
					})
				}
			}
		}
	}
}
