package facility

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/geo/location"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/domain/repository"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// FacilityWithDistance extends HealthcareFacility with distance information
type FacilityWithDistance struct {
	*model.HealthcareFacility
	DistanceKm     float64 `json:"distance_km"`
	DistanceFormatted string  `json:"distance_formatted"`
	TravelTimeMinutes int     `json:"travel_time_minutes,omitempty"`
	IsOpen          bool     `json:"is_open"`
}

// FacilityFilter defines options for filtering facilities
type FacilityFilter struct {
	Types          []model.FacilityType `json:"types,omitempty"`
	Services       []string             `json:"services,omitempty"`
	MaxDistance    float64              `json:"max_distance,omitempty"` // km
	OpenNow        bool                 `json:"open_now,omitempty"`
	MinCapacity    int                  `json:"min_capacity,omitempty"`
	District       string               `json:"district,omitempty"`
}

// Service provides facility search functionality
type Service struct {
	log            logger.Logger
	facilityRepo   repository.FacilityRepository
	locationService *location.Service
}

// NewService creates a new facility service
func NewService(
	log logger.Logger,
	facilityRepo repository.FacilityRepository,
	locationService *location.Service,
) *Service {
	return &Service{
		log:            log,
		facilityRepo:   facilityRepo,
		locationService: locationService,
	}
}

// FindNearbyFacilities finds healthcare facilities near a location
func (s *Service) FindNearbyFacilities(
	ctx context.Context,
	lat, lng float64,
	radiusKm float64,
	filter *FacilityFilter,
) ([]FacilityWithDistance, error) {
	// Validate coordinates
	_, err := s.locationService.ValidateCoordinates(lat, lng)
	if err != nil {
		return nil, err
	}

	// If no filter provided, create an empty one
	if filter == nil {
		filter = &FacilityFilter{}
	}

	// Set default radius if not provided
	if radiusKm <= 0 {
		radiusKm = 10.0 // Default 10km radius
	}

	// Use PostGIS to find facilities within radius
	facilities, err := s.facilityRepo.FindWithinRadius(ctx, lat, lng, radiusKm)
	if err != nil {
		s.log.Error("Failed to find facilities within radius", logger.Fields{
			"error":     err.Error(),
			"latitude":  lat,
			"longitude": lng,
			"radius_km": radiusKm,
		})
		return nil, errorx.New(errorx.Internal, "Failed to search for nearby facilities")
	}

	// Calculate distances and apply filters
	result := s.calculateDistancesAndFilter(ctx, facilities, lat, lng, filter)

	// Sort by distance
	sort.Slice(result, func(i, j int) bool {
		return result[i].DistanceKm < result[j].DistanceKm
	})

	return result, nil
}

// FindFacilityByID finds a facility by ID
func (s *Service) FindFacilityByID(ctx context.Context, id uuid.UUID) (*model.HealthcareFacility, error) {
	facility, err := s.facilityRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("Failed to find facility by ID", logger.Fields{
			"error": err.Error(),
			"id":    id.String(),
		})
		return nil, errorx.New(errorx.NotFound, "Facility not found")
	}

	return facility, nil
}

// calculateDistancesAndFilter calculates distances for facilities and applies filters
func (s *Service) calculateDistancesAndFilter(
	ctx context.Context,
	facilities []*model.HealthcareFacility,
	lat, lng float64,
	filter *FacilityFilter,
) []FacilityWithDistance {
	result := make([]FacilityWithDistance, 0)

	for _, facility := range facilities {
		// Skip if we're filtering by district and it doesn't match
		if filter.District != "" && facility.District != filter.District {
			continue
		}

		// Skip if we're filtering by facility type and it doesn't match
		if len(filter.Types) > 0 {
			typeMatches := false
			for _, facilityType := range filter.Types {
				if facility.FacilityType == facilityType {
					typeMatches = true
					break
				}
			}
			if !typeMatches {
				continue
			}
		}

		// Skip if we're filtering by minimum capacity and it doesn't meet requirements
		if filter.MinCapacity > 0 && facility.Capacity < filter.MinCapacity {
			continue
		}

		// Skip if we're filtering by services and it doesn't offer all required services
		if len(filter.Services) > 0 {
			hasAllServices := true
			facilityServicesMap := make(map[string]bool)
			
			for _, service := range facility.ServicesOffered {
				facilityServicesMap[service] = true
			}
			
			for _, requiredService := range filter.Services {
				if !facilityServicesMap[requiredService] {
					hasAllServices = false
					break
				}
			}
			
			if !hasAllServices {
				continue
			}
		}

		// Calculate distance
		distanceKm := s.locationService.CalculateDistance(
			lat, lng,
			facility.Location.Latitude, facility.Location.Longitude,
		)

		// Skip if beyond max distance
		if filter.MaxDistance > 0 && distanceKm > filter.MaxDistance {
			continue
		}

		// Check if facility is open now
		isOpen := true
		if filter.OpenNow {
			isOpen = facility.IsOpen(ctx.Value("current_time").(time.Time))
			if !isOpen {
				continue
			}
		}

		// Approximate travel time (assumes 30 km/h average speed)
		travelTimeMinutes := int(distanceKm * 2) // 2 minutes per km

		// Add to results
		result = append(result, FacilityWithDistance{
			HealthcareFacility: facility,
			DistanceKm:         distanceKm,
			DistanceFormatted:  s.locationService.FormatDistance(distanceKm),
			TravelTimeMinutes:  travelTimeMinutes,
			IsOpen:             isOpen,
		})
	}

	return result
}

// SearchFacilities searches for facilities by name or attributes
func (s *Service) SearchFacilities(
	ctx context.Context,
	query string,
	filter *FacilityFilter,
) ([]*model.HealthcareFacility, error) {
	facilities, err := s.facilityRepo.Search(ctx, query)
	if err != nil {
		s.log.Error("Failed to search facilities", logger.Fields{
			"error": err.Error(),
			"query": query,
		})
		return nil, errorx.New(errorx.Internal, "Failed to search for facilities")
	}

	// Apply filters
	if filter == nil {
		return facilities, nil
	}

	filtered := make([]*model.HealthcareFacility, 0)
	for _, facility := range facilities {
		// Apply district filter
		if filter.District != "" && facility.District != filter.District {
			continue
		}

		// Apply facility type filter
		if len(filter.Types) > 0 {
			typeMatches := false
			for _, facilityType := range filter.Types {
				if facility.FacilityType == facilityType {
					typeMatches = true
					break
				}
			}
			if !typeMatches {
				continue
			}
		}

		// Apply minimum capacity filter
		if filter.MinCapacity > 0 && facility.Capacity < filter.MinCapacity {
			continue
		}

		// Apply services filter
		if len(filter.Services) > 0 {
			hasAllServices := true
			facilityServicesMap := make(map[string]bool)
			
			for _, service := range facility.ServicesOffered {
				facilityServicesMap[service] = true
			}
			
			for _, requiredService := range filter.Services {
				if !facilityServicesMap[requiredService] {
					hasAllServices = false
					break
				}
			}
			
			if !hasAllServices {
				continue
			}
		}

		// Apply open now filter
		if filter.OpenNow {
			isOpen := facility.IsOpen(ctx.Value("current_time").(time.Time))
			if !isOpen {
				continue
			}
		}

		filtered = append(filtered, facility)
	}

	return filtered, nil
}
