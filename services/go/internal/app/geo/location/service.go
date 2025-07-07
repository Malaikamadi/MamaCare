package location

import (
	"fmt"
	"math"

	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// Service provides location-related functionality
type Service struct {
	log logger.Logger
}

// NewService creates a new location service
func NewService(log logger.Logger) *Service {
	return &Service{
		log: log,
	}
}

// ValidCoordinates represents a validated set of coordinates
type ValidCoordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	IsValid   bool    `json:"is_valid"`
	Country   string  `json:"country,omitempty"`
	Region    string  `json:"region,omitempty"`
}

// SierraLeoneRegion represents regions in Sierra Leone
type SierraLeoneRegion string

const (
	// RegionWestern represents Western Area (Freetown and surroundings)
	RegionWestern SierraLeoneRegion = "western"
	// RegionNorthern represents Northern Province
	RegionNorthern SierraLeoneRegion = "northern"
	// RegionEastern represents Eastern Province
	RegionEastern SierraLeoneRegion = "eastern"
	// RegionSouthern represents Southern Province
	RegionSouthern SierraLeoneRegion = "southern"
	// RegionNorthWest represents North West Province
	RegionNorthWest SierraLeoneRegion = "north_west"
)

// ValidateCoordinates checks if coordinates are valid
func (s *Service) ValidateCoordinates(lat, lng float64) (*ValidCoordinates, error) {
	if lat < -90 || lat > 90 {
		return nil, errorx.New(errorx.BadRequest, "latitude must be between -90 and 90")
	}

	if lng < -180 || lng > 180 {
		return nil, errorx.New(errorx.BadRequest, "longitude must be between -180 and 180")
	}

	// Create valid coordinates
	coords := &ValidCoordinates{
		Latitude:  lat,
		Longitude: lng,
		IsValid:   true,
		Country:   "Sierra Leone", // Assuming all coordinates in Sierra Leone for MVP
	}

	// Determine region in Sierra Leone based on coordinates
	coords.Region = s.determineRegion(lat, lng)

	return coords, nil
}

// IsSierraLeone checks if coordinates are within Sierra Leone
// Sierra Leone approximate bounding box:
// North: 10.0
// South: 6.9
// West: -13.3
// East: -10.3
func (s *Service) IsSierraLeone(lat, lng float64) bool {
	return lat >= 6.9 && lat <= 10.0 && lng >= -13.3 && lng <= -10.3
}

// determineRegion determines which region of Sierra Leone the coordinates are in
// This is a simple approximation for MVP purposes
func (s *Service) determineRegion(lat, lng float64) string {
	// Simplified regional boundaries for Sierra Leone
	// Western (Freetown): Around 8.4, -13.2
	// Northern: Between 9.0 and 10.0, -12.0 and -11.0
	// Eastern: Between 7.5 and 9.0, -11.5 and -10.3
	// Southern: Between 6.9 and 8.0, -12.5 and -11.0
	// North West: Between 9.0 and 10.0, -13.0 and -12.0

	if lat >= 8.0 && lat <= 8.6 && lng >= -13.3 && lng <= -13.0 {
		return string(RegionWestern) // Western Area (Freetown)
	} else if lat >= 9.0 && lat <= 10.0 && lng >= -12.0 && lng <= -11.0 {
		return string(RegionNorthern) // Northern Province
	} else if lat >= 7.5 && lat <= 9.0 && lng >= -11.5 && lng <= -10.3 {
		return string(RegionEastern) // Eastern Province
	} else if lat >= 6.9 && lat <= 8.0 && lng >= -12.5 && lng <= -11.0 {
		return string(RegionSouthern) // Southern Province
	} else if lat >= 9.0 && lat <= 10.0 && lng >= -13.0 && lng <= -12.0 {
		return string(RegionNorthWest) // North West Province
	}

	return "unknown"
}

// NormalizeCoordinates ensures coordinates have appropriate precision
func (s *Service) NormalizeCoordinates(lat, lng float64) (float64, float64) {
	// Round to 6 decimal places (approximately 0.1 meter precision)
	normalizedLat := math.Round(lat*1000000) / 1000000
	normalizedLng := math.Round(lng*1000000) / 1000000

	return normalizedLat, normalizedLng
}

// CoordinatesToLocation converts coordinates to a model.Location
func (s *Service) CoordinatesToLocation(lat, lng float64) (*model.Location, error) {
	coords, err := s.ValidateCoordinates(lat, lng)
	if err != nil {
		return nil, err
	}

	normalizedLat, normalizedLng := s.NormalizeCoordinates(coords.Latitude, coords.Longitude)

	return &model.Location{
		Latitude:  normalizedLat,
		Longitude: normalizedLng,
	}, nil
}

// CalculateDistance calculates the distance between two coordinates in kilometers
// using the Haversine formula
func (s *Service) CalculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371.0 // Earth radius in kilometers

	// Convert latitude and longitude from degrees to radians
	lat1Rad := toRadians(lat1)
	lng1Rad := toRadians(lng1)
	lat2Rad := toRadians(lat2)
	lng2Rad := toRadians(lng2)

	// Differences
	latDiff := lat2Rad - lat1Rad
	lngDiff := lng2Rad - lng1Rad

	// Haversine formula
	a := math.Sin(latDiff/2)*math.Sin(latDiff/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(lngDiff/2)*math.Sin(lngDiff/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := earthRadius * c

	return distance
}

// FormatDistance formats a distance nicely for display
func (s *Service) FormatDistance(distance float64) string {
	if distance < 1.0 {
		// If less than 1 km, show in meters
		meters := int(distance * 1000)
		return fmt.Sprintf("%d m", meters)
	}

	// Otherwise show in kilometers with one decimal place
	return fmt.Sprintf("%.1f km", distance)
}

// toRadians converts degrees to radians
func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}
