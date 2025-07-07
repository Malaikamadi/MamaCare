package model

import (
	"time"

	"github.com/google/uuid"
)

// FacilityType represents the type of a healthcare facility
type FacilityType string

const (
	// FacilityTypeHospital represents a hospital
	FacilityTypeHospital FacilityType = "hospital"
	// FacilityTypeClinic represents a clinic
	FacilityTypeClinic FacilityType = "clinic"
	// FacilityTypeHealthCenter represents a health center
	FacilityTypeHealthCenter FacilityType = "health_center"
	// FacilityTypeHealthPost represents a health post
	FacilityTypeHealthPost FacilityType = "health_post"
)

// OperatingHours represents the operating hours of a facility
type OperatingHours struct {
	Monday    DayHours `json:"monday"`
	Tuesday   DayHours `json:"tuesday"`
	Wednesday DayHours `json:"wednesday"`
	Thursday  DayHours `json:"thursday"`
	Friday    DayHours `json:"friday"`
	Saturday  DayHours `json:"saturday"`
	Sunday    DayHours `json:"sunday"`
}

// DayHours represents the operating hours for a day
type DayHours struct {
	Open  string `json:"open"`  // Time format: "08:00"
	Close string `json:"close"` // Time format: "17:00"
	IsClosed bool   `json:"is_closed"`
}

// Location represents a geographical location
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// HealthcareFacility represents a healthcare facility
type HealthcareFacility struct {
	ID            uuid.UUID     `json:"id"`
	Name          string        `json:"name"`
	Address       string        `json:"address"`
	District      string        `json:"district"`
	Location      Location      `json:"location"`
	FacilityType  FacilityType  `json:"facility_type"`
	Capacity      int           `json:"capacity"`
	OperatingHours OperatingHours `json:"operating_hours"`
	ServicesOffered []string      `json:"services_offered"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// NewHealthcareFacility creates a new healthcare facility
func NewHealthcareFacility(id uuid.UUID, name, address, district string, lat, lng float64, facilityType FacilityType) *HealthcareFacility {
	now := time.Now()
	return &HealthcareFacility{
		ID:           id,
		Name:         name,
		Address:      address,
		District:     district,
		Location: Location{
			Latitude:  lat,
			Longitude: lng,
		},
		FacilityType:   facilityType,
		CreatedAt:    now,
		UpdatedAt:    now,
		ServicesOffered: []string{},
	}
}

// WithCapacity sets the capacity of the facility
func (f *HealthcareFacility) WithCapacity(capacity int) *HealthcareFacility {
	f.Capacity = capacity
	return f
}

// WithOperatingHours sets the operating hours of the facility
func (f *HealthcareFacility) WithOperatingHours(hours OperatingHours) *HealthcareFacility {
	f.OperatingHours = hours
	return f
}

// WithServices adds services offered by the facility
func (f *HealthcareFacility) WithServices(services []string) *HealthcareFacility {
	f.ServicesOffered = services
	return f
}

// IsOpen checks if the facility is open at the given time
func (f *HealthcareFacility) IsOpen(t time.Time) bool {
	// Get the day of the week
	day := t.Weekday()
	
	var hours DayHours
	switch day {
	case time.Monday:
		hours = f.OperatingHours.Monday
	case time.Tuesday:
		hours = f.OperatingHours.Tuesday
	case time.Wednesday:
		hours = f.OperatingHours.Wednesday
	case time.Thursday:
		hours = f.OperatingHours.Thursday
	case time.Friday:
		hours = f.OperatingHours.Friday
	case time.Saturday:
		hours = f.OperatingHours.Saturday
	case time.Sunday:
		hours = f.OperatingHours.Sunday
	}
	
	if hours.IsClosed {
		return false
	}
	
	// Parse time strings
	timeFormat := "15:04"
	openTime, _ := time.Parse(timeFormat, hours.Open)
	closeTime, _ := time.Parse(timeFormat, hours.Close)
	
	// Get current hour and minute
	currentTime, _ := time.Parse(timeFormat, t.Format(timeFormat))
	
	// Check if the current time is within operating hours
	return (currentTime.Equal(openTime) || currentTime.After(openTime)) && 
		   (currentTime.Before(closeTime))
}

// OffersService checks if the facility offers a specific service
func (f *HealthcareFacility) OffersService(service string) bool {
	for _, s := range f.ServicesOffered {
		if s == service {
			return true
		}
	}
	return false
}
