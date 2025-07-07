package dispatch

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/incognito25/mamacare/services/go/internal/domain/model"
	"github.com/incognito25/mamacare/services/go/internal/domain/repository"
	"github.com/incognito25/mamacare/services/go/internal/errorx"
	"github.com/incognito25/mamacare/services/go/internal/logger"
)

// Service provides ambulance dispatch functionality
type Service struct {
	ambulanceRepo repository.AmbulanceRepository
	sosRepo       repository.SOSRepository
	facilityRepo  repository.FacilityRepository
	routingEngine RoutingEngine
	notifier      DispatchNotifier
	logger        logger.Logger
}

// RoutingEngine defines the interface for calculating routes
type RoutingEngine interface {
	// CalculateRoute calculates a route between two points
	CalculateRoute(ctx context.Context, fromLat, fromLng, toLat, toLng float64) (*model.Route, error)
	
	// EstimateTimeOfArrival estimates time of arrival between two points
	EstimateTimeOfArrival(ctx context.Context, fromLat, fromLng, toLat, toLng float64) (time.Duration, error)
}

// DispatchNotifier defines the interface for dispatch notifications
type DispatchNotifier interface {
	// NotifyDispatch notifies about an ambulance dispatch
	NotifyDispatch(ctx context.Context, sosEvent *model.SOSEvent, ambulance *model.Ambulance, eta time.Time) error
	
	// NotifyStatusUpdate notifies about dispatch status updates
	NotifyStatusUpdate(ctx context.Context, sosEvent *model.SOSEvent, ambulance *model.Ambulance, status model.AmbulanceStatus) error
}

// NewService creates a new dispatch service
func NewService(
	ambulanceRepo repository.AmbulanceRepository,
	sosRepo repository.SOSRepository,
	facilityRepo repository.FacilityRepository,
	routingEngine RoutingEngine,
	notifier DispatchNotifier,
	logger logger.Logger,
) *Service {
	return &Service{
		ambulanceRepo: ambulanceRepo,
		sosRepo:       sosRepo,
		facilityRepo:  facilityRepo,
		routingEngine: routingEngine,
		notifier:      notifier,
		logger:        logger,
	}
}

// DispatchAmbulance dispatches an ambulance to an SOS event
func (s *Service) DispatchAmbulance(ctx context.Context, sosID, ambulanceID uuid.UUID) (*model.SOSEvent, error) {
	// Get SOS event
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for dispatch", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}

	// Check if SOS event is in a dispatchable state
	if sosEvent.Status != model.SOSEventStatusReported {
		return nil, errorx.New(errorx.Validation, fmt.Sprintf("SOS event is not in a dispatchable state: %s", sosEvent.Status))
	}

	// Get ambulance
	ambulance, err := s.ambulanceRepo.GetByID(ctx, ambulanceID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get ambulance for dispatch", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": ambulanceID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "Ambulance not found", err)
	}

	// Check if ambulance is available
	if !ambulance.IsAvailable() {
		return nil, errorx.New(errorx.Validation, fmt.Sprintf("Ambulance is not available: %s", ambulance.Status))
	}

	// Calculate ETA
	var eta time.Time
	if ambulance.Location != nil {
		duration, err := s.routingEngine.EstimateTimeOfArrival(
			ctx,
			ambulance.Location.Latitude,
			ambulance.Location.Longitude,
			sosEvent.Location.Latitude,
			sosEvent.Location.Longitude,
		)
		if err != nil {
			s.logger.Error(ctx, "Failed to estimate time of arrival", logger.FieldsMap{
				"error":        err.Error(),
				"ambulance_id": ambulanceID.String(),
				"sos_id":      sosID.String(),
			})
			// Continue with dispatch even if ETA calculation fails
			eta = time.Now().Add(time.Hour) // Default 1 hour ETA
		} else {
			eta = time.Now().Add(duration)
		}
	} else {
		// If ambulance location is unknown, use default ETA
		eta = time.Now().Add(time.Hour) // Default 1 hour ETA
	}

	// Update ambulance status
	ambulance.Dispatch(sosID)
	if err := s.ambulanceRepo.Update(ctx, ambulance); err != nil {
		s.logger.Error(ctx, "Failed to update ambulance status for dispatch", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": ambulanceID.String(),
			"sos_id":      sosID.String(),
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to update ambulance status", err)
	}

	// Update SOS event with ambulance and dispatch status
	sosEvent.Dispatch(ambulanceID, eta)
	if err := s.sosRepo.Update(ctx, sosEvent); err != nil {
		s.logger.Error(ctx, "Failed to update SOS event for dispatch", logger.FieldsMap{
			"error":        err.Error(),
			"sos_id":      sosID.String(),
			"ambulance_id": ambulanceID.String(),
		})
		// Try to revert ambulance status
		ambulance.MarkAvailable()
		_ = s.ambulanceRepo.Update(ctx, ambulance)
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to update SOS event with dispatch info", err)
	}

	// Send dispatch notifications
	if err := s.notifier.NotifyDispatch(ctx, sosEvent, ambulance, eta); err != nil {
		s.logger.Error(ctx, "Failed to send dispatch notification", logger.FieldsMap{
			"error":        err.Error(),
			"sos_id":      sosID.String(),
			"ambulance_id": ambulanceID.String(),
		})
		// Continue despite notification error
	}

	s.logger.Info(ctx, "Ambulance dispatched to SOS event", logger.FieldsMap{
		"sos_id":      sosID.String(),
		"ambulance_id": ambulanceID.String(),
		"eta":          eta.Format(time.RFC3339),
	})

	return sosEvent, nil
}

// FindSuitableAmbulances finds suitable ambulances for an SOS event
func (s *Service) FindSuitableAmbulances(ctx context.Context, sosID uuid.UUID, maxResults int) ([]*model.Ambulance, error) {
	// Get SOS event
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for finding ambulances", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}

	// Get available ambulances within a reasonable radius
	const radiusKm = 20.0
	ambulances, err := s.ambulanceRepo.GetAvailableInRadius(
		ctx,
		sosEvent.Location.Latitude,
		sosEvent.Location.Longitude,
		radiusKm,
	)
	if err != nil {
		s.logger.Error(ctx, "Failed to get available ambulances in radius", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
			"radius":  radiusKm,
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to find ambulances", err)
	}

	if len(ambulances) == 0 {
		s.logger.Info(ctx, "No available ambulances found in radius", logger.FieldsMap{
			"sos_id": sosID.String(),
			"radius":  radiusKm,
		})
		// Try to get any available ambulance regardless of distance
		ambulances, err = s.ambulanceRepo.GetByStatus(ctx, model.AmbulanceStatusAvailable)
		if err != nil {
			s.logger.Error(ctx, "Failed to get any available ambulances", logger.FieldsMap{
				"error":   err.Error(),
				"sos_id": sosID.String(),
			})
			return nil, errorx.NewWithCause(errorx.Internal, "Failed to find any ambulances", err)
		}
	}

	// Score and rank ambulances
	scoredAmbulances := s.scoreAmbulances(ctx, sosEvent, ambulances)

	// Sort by score (highest first)
	sort.Slice(scoredAmbulances, func(i, j int) bool {
		return scoredAmbulances[i].score > scoredAmbulances[j].score
	})

	// Limit results
	if maxResults > 0 && len(scoredAmbulances) > maxResults {
		scoredAmbulances = scoredAmbulances[:maxResults]
	}

	// Extract just the ambulances
	result := make([]*model.Ambulance, 0, len(scoredAmbulances))
	for _, sa := range scoredAmbulances {
		result = append(result, sa.ambulance)
	}

	return result, nil
}

// UpdateAmbulanceStatus updates the status of an ambulance during an emergency
func (s *Service) UpdateAmbulanceStatus(ctx context.Context, ambulanceID uuid.UUID, status model.AmbulanceStatus) (*model.Ambulance, error) {
	// Get ambulance
	ambulance, err := s.ambulanceRepo.GetByID(ctx, ambulanceID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get ambulance for status update", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": ambulanceID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "Ambulance not found", err)
	}

	// Check if ambulance has an active SOS event
	if status != model.AmbulanceStatusAvailable && status != model.AmbulanceStatusMaintenance {
		if !ambulance.IsActive() || ambulance.CurrentSOSID == nil {
			return nil, errorx.New(errorx.Validation, "Ambulance is not currently on an active SOS event")
		}
	}

	// Update ambulance status based on the requested status
	switch status {
	case model.AmbulanceStatusEnRoute:
		ambulance.MarkEnRoute()
	case model.AmbulanceStatusArrived:
		ambulance.MarkArrived()
	case model.AmbulanceStatusReturning:
		ambulance.MarkReturning()
	case model.AmbulanceStatusAvailable:
		ambulance.MarkAvailable()
	case model.AmbulanceStatusMaintenance:
		ambulance.MarkMaintenance()
	default:
		return nil, errorx.New(errorx.Validation, fmt.Sprintf("Invalid ambulance status: %s", status))
	}

	// Update ambulance in repository
	if err := s.ambulanceRepo.Update(ctx, ambulance); err != nil {
		s.logger.Error(ctx, "Failed to update ambulance status", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": ambulanceID.String(),
			"status":       string(status),
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to update ambulance status", err)
	}

	// If ambulance has an active SOS event, send status notifications
	if ambulance.CurrentSOSID != nil && ambulance.IsActive() {
		sosEvent, err := s.sosRepo.GetByID(ctx, *ambulance.CurrentSOSID)
		if err == nil {
			if err := s.notifier.NotifyStatusUpdate(ctx, sosEvent, ambulance, status); err != nil {
				s.logger.Error(ctx, "Failed to send ambulance status update notification", logger.FieldsMap{
					"error":        err.Error(),
					"ambulance_id": ambulanceID.String(),
					"sos_id":      ambulance.CurrentSOSID.String(),
					"status":       string(status),
				})
				// Continue despite notification error
			}
		}
	}

	s.logger.Info(ctx, "Updated ambulance status", logger.FieldsMap{
		"ambulance_id": ambulanceID.String(),
		"status":       string(status),
	})

	return ambulance, nil
}

// UpdateAmbulanceLocation updates the location of an ambulance
func (s *Service) UpdateAmbulanceLocation(ctx context.Context, ambulanceID uuid.UUID, lat, lng float64) (*model.Ambulance, error) {
	// Update location in repository
	if err := s.ambulanceRepo.UpdateLocation(ctx, ambulanceID, lat, lng); err != nil {
		s.logger.Error(ctx, "Failed to update ambulance location", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": ambulanceID.String(),
			"lat":          fmt.Sprintf("%f", lat),
			"lng":          fmt.Sprintf("%f", lng),
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to update ambulance location", err)
	}

	// Get updated ambulance
	ambulance, err := s.ambulanceRepo.GetByID(ctx, ambulanceID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get ambulance after location update", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": ambulanceID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "Ambulance not found", err)
	}

	// If ambulance is active on an SOS event, update ETA
	if ambulance.IsActive() && ambulance.CurrentSOSID != nil {
		sosEvent, err := s.sosRepo.GetByID(ctx, *ambulance.CurrentSOSID)
		if err == nil && (sosEvent.Status == model.SOSEventStatusDispatched) {
			s.updateETA(ctx, sosEvent, ambulance)
		}
	}

	return ambulance, nil
}

// UpdateETA updates the estimated time of arrival for an active dispatch
func (s *Service) updateETA(ctx context.Context, sosEvent *model.SOSEvent, ambulance *model.Ambulance) {
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
		s.logger.Error(ctx, "Failed to recalculate ETA", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": ambulance.ID.String(),
			"sos_id":      sosEvent.ID.String(),
		})
		return
	}

	// Update ETA in SOS event
	eta := time.Now().Add(duration)
	sosEvent.UpdateETA(eta)
	if err := s.sosRepo.Update(ctx, sosEvent); err != nil {
		s.logger.Error(ctx, "Failed to update SOS event with new ETA", logger.FieldsMap{
			"error":        err.Error(),
			"sos_id":      sosEvent.ID.String(),
			"ambulance_id": ambulance.ID.String(),
		})
	}
}

// scoredAmbulance represents an ambulance with a score for ranking
type scoredAmbulance struct {
	ambulance *model.Ambulance
	score     float64
	distance  float64 // in kilometers
	etaMinutes int    // estimated time of arrival in minutes
}

// scoreAmbulances scores and ranks ambulances based on various factors
func (s *Service) scoreAmbulances(ctx context.Context, sosEvent *model.SOSEvent, ambulances []*model.Ambulance) []scoredAmbulance {
	scoredAmbulances := make([]scoredAmbulance, 0, len(ambulances))

	for _, ambulance := range ambulances {
		sa := scoredAmbulance{
			ambulance: ambulance,
			score:     0,
			distance:  -1,
			etaMinutes: -1,
		}

		// Base score on ambulance type
		switch ambulance.AmbulanceType {
		case model.AmbulanceTypeOB:
			// Prioritize obstetric ambulances for maternal emergencies
			sa.score += 50
		case model.AmbulanceTypeAdvanced:
			sa.score += 30
		case model.AmbulanceTypeBasic:
			sa.score += 10
		}

		// Score based on location if available
		if ambulance.Location != nil {
			// Calculate distance
			distance := calculateDistance(
				ambulance.Location.Latitude,
				ambulance.Location.Longitude,
				sosEvent.Location.Latitude,
				sosEvent.Location.Longitude,
			)
			sa.distance = distance

			// Lower score for greater distance (inverse relationship)
			distanceScore := 100 / (distance + 1) // +1 to avoid division by zero
			sa.score += distanceScore

			// Calculate ETA if possible
			duration, err := s.routingEngine.EstimateTimeOfArrival(
				ctx,
				ambulance.Location.Latitude,
				ambulance.Location.Longitude,
				sosEvent.Location.Latitude,
				sosEvent.Location.Longitude,
			)
			if err == nil {
				etaMinutes := int(duration.Minutes())
				sa.etaMinutes = etaMinutes

				// Lower score for longer ETA (inverse relationship)
				etaScore := 200 / (float64(etaMinutes) + 1) // +1 to avoid division by zero
				sa.score += etaScore
			}
		}

		// Score based on capacity
		sa.score += float64(ambulance.Capacity) * 5

		// Add the scored ambulance to the result
		scoredAmbulances = append(scoredAmbulances, sa)
	}

	return scoredAmbulances
}

// calculateDistance calculates the distance between two coordinates (simplified version)
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Use a very simplified distance calculation for demonstration
	// In a real app, we would use the Haversine formula or similar
	latDiff := lat2 - lat1
	lonDiff := lon2 - lon1
	return float64(latDiff*latDiff + lonDiff*lonDiff)
}
