package territory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/geo/location"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/domain/repository"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// Territory represents a geographic area assigned to a CHW
type Territory struct {
	ID          uuid.UUID       `json:"id"`
	CHWID       uuid.UUID       `json:"chw_id"`
	Name        string          `json:"name"`
	District    string          `json:"district"`
	Description string          `json:"description,omitempty"`
	Boundaries  []model.Location `json:"boundaries"`
	CenterPoint model.Location   `json:"center_point"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// AssignmentResult represents the result of a territory assignment
type AssignmentResult struct {
	Territory      *Territory   `json:"territory"`
	CHW            *model.User  `json:"chw"`
	MotherCount    int          `json:"mother_count"`
	AssignedAt     time.Time    `json:"assigned_at"`
}

// MotherInTerritory represents a mother with territory information
type MotherInTerritory struct {
	MotherID        uuid.UUID     `json:"mother_id"`
	UserID          uuid.UUID     `json:"user_id"`
	Location        model.Location `json:"location"`
	TerritoryID     *uuid.UUID    `json:"territory_id,omitempty"`
	TerritoryName   string        `json:"territory_name,omitempty"`
	CHWID           *uuid.UUID    `json:"chw_id,omitempty"`
	CHWName         string        `json:"chw_name,omitempty"`
	DistanceFromCHW float64       `json:"distance_from_chw,omitempty"`
}

// Service provides territory management functionality
type Service struct {
	log            logger.Logger
	territoryRepo  repository.TerritoryRepository
	userRepo       repository.UserRepository
	motherRepo     repository.MotherRepository
	locationService *location.Service
}

// NewService creates a new territory service
func NewService(
	log logger.Logger,
	territoryRepo repository.TerritoryRepository,
	userRepo repository.UserRepository,
	motherRepo repository.MotherRepository,
	locationService *location.Service,
) *Service {
	return &Service{
		log:             log,
		territoryRepo:   territoryRepo,
		userRepo:        userRepo,
		motherRepo:      motherRepo,
		locationService: locationService,
	}
}

// CreateTerritory creates a new territory for a CHW
func (s *Service) CreateTerritory(
	ctx context.Context,
	chwID uuid.UUID,
	name, district, description string,
	boundaries []model.Location,
) (*Territory, error) {
	// Verify CHW exists and has the correct role
	chw, err := s.userRepo.GetByID(ctx, chwID)
	if err != nil {
		return nil, errorx.New(errorx.NotFound, "CHW not found")
	}

	// Ensure the user has the CHW role
	if chw.Role != "chw" {
		return nil, errorx.New(errorx.BadRequest, "Only CHWs can be assigned territories")
	}

	// Calculate center point
	centerPoint, err := s.calculateCenterPoint(boundaries)
	if err != nil {
		return nil, err
	}

	// Create the territory
	territory := &Territory{
		ID:          uuid.New(),
		CHWID:       chwID,
		Name:        name,
		District:    district,
		Description: description,
		Boundaries:  boundaries,
		CenterPoint: centerPoint,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save territory
	err = s.territoryRepo.Create(ctx, territory)
	if err != nil {
		s.log.Error("Failed to create territory", logger.Fields{
			"error":    err.Error(),
			"chw_id":   chwID.String(),
			"territory_name": name,
		})
		return nil, errorx.New(errorx.Internal, "Failed to create territory")
	}

	return territory, nil
}

// GetTerritory gets a territory by ID
func (s *Service) GetTerritory(
	ctx context.Context,
	territoryID uuid.UUID,
) (*Territory, error) {
	territory, err := s.territoryRepo.GetByID(ctx, territoryID)
	if err != nil {
		return nil, errorx.New(errorx.NotFound, "Territory not found")
	}

	return territory, nil
}

// GetTerritoryForCHW gets a territory for a CHW
func (s *Service) GetTerritoryForCHW(
	ctx context.Context,
	chwID uuid.UUID,
) (*Territory, error) {
	territory, err := s.territoryRepo.GetByCHWID(ctx, chwID)
	if err != nil {
		return nil, errorx.New(errorx.NotFound, "No territory found for this CHW")
	}

	return territory, nil
}

// UpdateTerritory updates a territory
func (s *Service) UpdateTerritory(
	ctx context.Context,
	territoryID uuid.UUID,
	name, district, description string,
	boundaries []model.Location,
) (*Territory, error) {
	// Get existing territory
	territory, err := s.territoryRepo.GetByID(ctx, territoryID)
	if err != nil {
		return nil, errorx.New(errorx.NotFound, "Territory not found")
	}

	// Calculate new center point if boundaries changed
	if len(boundaries) > 0 {
		centerPoint, err := s.calculateCenterPoint(boundaries)
		if err != nil {
			return nil, err
		}
		territory.Boundaries = boundaries
		territory.CenterPoint = centerPoint
	}

	// Update fields
	if name != "" {
		territory.Name = name
	}
	if district != "" {
		territory.District = district
	}
	territory.Description = description
	territory.UpdatedAt = time.Now()

	// Save updated territory
	err = s.territoryRepo.Update(ctx, territory)
	if err != nil {
		s.log.Error("Failed to update territory", logger.Fields{
			"error":        err.Error(),
			"territory_id": territoryID.String(),
		})
		return nil, errorx.New(errorx.Internal, "Failed to update territory")
	}

	return territory, nil
}

// AssignCHWToTerritory assigns a CHW to a territory
func (s *Service) AssignCHWToTerritory(
	ctx context.Context,
	chwID, territoryID uuid.UUID,
) (*AssignmentResult, error) {
	// Verify CHW exists and has the correct role
	chw, err := s.userRepo.GetByID(ctx, chwID)
	if err != nil {
		return nil, errorx.New(errorx.NotFound, "CHW not found")
	}

	// Ensure the user has the CHW role
	if chw.Role != "chw" {
		return nil, errorx.New(errorx.BadRequest, "Only CHWs can be assigned territories")
	}

	// Get territory
	territory, err := s.territoryRepo.GetByID(ctx, territoryID)
	if err != nil {
		return nil, errorx.New(errorx.NotFound, "Territory not found")
	}

	// Update territory with new CHW
	territory.CHWID = chwID
	territory.UpdatedAt = time.Now()

	// Save updated territory
	err = s.territoryRepo.Update(ctx, territory)
	if err != nil {
		s.log.Error("Failed to assign CHW to territory", logger.Fields{
			"error":        err.Error(),
			"territory_id": territoryID.String(),
			"chw_id":       chwID.String(),
		})
		return nil, errorx.New(errorx.Internal, "Failed to assign CHW to territory")
	}

	// Count mothers in territory
	mothers, err := s.motherRepo.FindInTerritory(ctx, territoryID)
	if err != nil {
		s.log.Warn("Failed to count mothers in territory", logger.Fields{
			"error":        err.Error(),
			"territory_id": territoryID.String(),
		})
	}

	// Create assignment result
	result := &AssignmentResult{
		Territory:   territory,
		CHW:         chw,
		MotherCount: len(mothers),
		AssignedAt:  time.Now(),
	}

	return result, nil
}

// FindTerritoryForLocation finds the territory that contains a location
func (s *Service) FindTerritoryForLocation(
	ctx context.Context,
	lat, lng float64,
) (*Territory, error) {
	// Validate coordinates
	_, err := s.locationService.ValidateCoordinates(lat, lng)
	if err != nil {
		return nil, err
	}

	// Find territory containing point
	territory, err := s.territoryRepo.FindContainingPoint(ctx, lat, lng)
	if err != nil {
		return nil, errorx.New(errorx.NotFound, "No territory found for this location")
	}

	return territory, nil
}

// FindNearestCHW finds the nearest CHW to a location
func (s *Service) FindNearestCHW(
	ctx context.Context,
	lat, lng float64,
	maxDistanceKm float64,
) (*model.User, float64, error) {
	// Validate coordinates
	_, err := s.locationService.ValidateCoordinates(lat, lng)
	if err != nil {
		return nil, 0, err
	}

	// If no max distance specified, use a default
	if maxDistanceKm <= 0 {
		maxDistanceKm = 50.0 // 50km default
	}

	// Get all CHWs with their territories
	territories, err := s.territoryRepo.GetAllWithCHWs(ctx)
	if err != nil {
		s.log.Error("Failed to get territories with CHWs", logger.Fields{
			"error": err.Error(),
		})
		return nil, 0, errorx.New(errorx.Internal, "Failed to find nearest CHW")
	}

	// Find the nearest CHW
	var closestCHW *model.User
	var closestDistance float64 = -1.0

	for _, t := range territories {
		// Calculate distance from location to territory center
		distance := s.locationService.CalculateDistance(
			lat, lng,
			t.CenterPoint.Latitude, t.CenterPoint.Longitude,
		)

		// Update closest CHW if this one is closer
		if closestDistance < 0 || distance < closestDistance {
			// Get the CHW for this territory
			chw, err := s.userRepo.GetByID(ctx, t.CHWID)
			if err != nil {
				// Skip if CHW not found
				continue
			}

			closestCHW = chw
			closestDistance = distance
		}
	}

	// Check if we found a CHW within the max distance
	if closestCHW == nil || (maxDistanceKm > 0 && closestDistance > maxDistanceKm) {
		return nil, 0, errorx.New(errorx.NotFound, "No CHW found within the specified distance")
	}

	return closestCHW, closestDistance, nil
}

// GetMothersInTerritory gets all mothers in a territory
func (s *Service) GetMothersInTerritory(
	ctx context.Context,
	territoryID uuid.UUID,
) ([]MotherInTerritory, error) {
	// Get territory
	territory, err := s.territoryRepo.GetByID(ctx, territoryID)
	if err != nil {
		return nil, errorx.New(errorx.NotFound, "Territory not found")
	}

	// Find mothers in territory
	mothers, err := s.motherRepo.FindInTerritory(ctx, territoryID)
	if err != nil {
		s.log.Error("Failed to find mothers in territory", logger.Fields{
			"error":        err.Error(),
			"territory_id": territoryID.String(),
		})
		return nil, errorx.New(errorx.Internal, "Failed to find mothers in territory")
	}

	// Get CHW information
	var chw *model.User
	if territory.CHWID != uuid.Nil {
		chw, err = s.userRepo.GetByID(ctx, territory.CHWID)
		if err != nil {
			s.log.Warn("CHW not found for territory", logger.Fields{
				"error":        err.Error(),
				"territory_id": territoryID.String(),
				"chw_id":       territory.CHWID.String(),
			})
		}
	}

	// Create result
	result := make([]MotherInTerritory, 0, len(mothers))
	for _, mother := range mothers {
		// Get user details for the mother
		user, err := s.userRepo.GetByID(ctx, mother.UserID)
		if err != nil {
			// Skip if user not found
			continue
		}

		// Calculate distance from CHW if CHW exists
		var distanceFromCHW float64 = -1
		
		// For simplicity, we're assuming mother location is stored in user profile or other location
		// In a real implementation, you would have a better way to get mother's location
		var motherLocation model.Location
		
		// If we have CHW and mother location, calculate distance
		if chw != nil && motherLocation.Latitude != 0 && motherLocation.Longitude != 0 {
			// Again, assuming CHW location is stored or derived from territory
			distanceFromCHW = s.locationService.CalculateDistance(
				motherLocation.Latitude, motherLocation.Longitude,
				territory.CenterPoint.Latitude, territory.CenterPoint.Longitude,
			)
		}

		// Create mother in territory
		mit := MotherInTerritory{
			MotherID:      mother.ID,
			UserID:        mother.UserID,
			Location:      motherLocation,
			TerritoryID:   &territoryID,
			TerritoryName: territory.Name,
		}

		// Add CHW info if available
		if chw != nil {
			mit.CHWID = &chw.ID
			mit.CHWName = chw.Name
			mit.DistanceFromCHW = distanceFromCHW
		}

		result = append(result, mit)
	}

	return result, nil
}

// calculateCenterPoint calculates the center point of a polygon
func (s *Service) calculateCenterPoint(boundaries []model.Location) (model.Location, error) {
	if len(boundaries) < 3 {
		return model.Location{}, errorx.New(errorx.BadRequest, "Territory must have at least 3 boundary points")
	}

	// Calculate center as average of all points
	var sumLat, sumLng float64
	for _, point := range boundaries {
		sumLat += point.Latitude
		sumLng += point.Longitude
	}

	centerLat := sumLat / float64(len(boundaries))
	centerLng := sumLng / float64(len(boundaries))

	// Normalize coordinates
	centerLat, centerLng = s.locationService.NormalizeCoordinates(centerLat, centerLng)

	return model.Location{
		Latitude:  centerLat,
		Longitude: centerLng,
	}, nil
}

// IsPointInTerritory checks if a point is inside a territory
func (s *Service) IsPointInTerritory(point model.Location, territory *Territory) bool {
	// Ray casting algorithm to determine if point is inside polygon
	boundaries := territory.Boundaries
	inside := false
	
	// Need at least 3 points to form a polygon
	if len(boundaries) < 3 {
		return false
	}
	
	j := len(boundaries) - 1
	for i := 0; i < len(boundaries); i++ {
		// Check if point is on boundary edge
		if (boundaries[i].Latitude == point.Latitude && boundaries[i].Longitude == point.Longitude) ||
		   (boundaries[j].Latitude == point.Latitude && boundaries[j].Longitude == point.Longitude) {
			return true
		}
		
		// Check if point is inside boundary
		if ((boundaries[i].Latitude < point.Latitude && boundaries[j].Latitude >= point.Latitude) ||
			(boundaries[j].Latitude < point.Latitude && boundaries[i].Latitude >= point.Latitude)) &&
			(boundaries[i].Longitude + (point.Latitude-boundaries[i].Latitude)/
				(boundaries[j].Latitude-boundaries[i].Latitude)*
				(boundaries[j].Longitude-boundaries[i].Longitude) < point.Longitude) {
			inside = !inside
		}
		j = i
	}
	
	return inside
}
