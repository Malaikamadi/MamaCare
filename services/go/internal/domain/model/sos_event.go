package model

import (
	"time"

	"github.com/google/uuid"
)

// SOSEventNature represents the nature of an SOS event
type SOSEventNature string

const (
	// SOSEventNatureLabor represents a labor emergency
	SOSEventNatureLabor SOSEventNature = "labor"
	// SOSEventNatureBleeding represents a bleeding emergency
	SOSEventNatureBleeding SOSEventNature = "bleeding"
	// SOSEventNatureAccident represents an accident emergency
	SOSEventNatureAccident SOSEventNature = "accident"
	// SOSEventNatureOther represents other types of emergencies
	SOSEventNatureOther SOSEventNature = "other"
)

// SOSEventStatus represents the status of an SOS event
type SOSEventStatus string

const (
	// SOSEventStatusReported represents a reported SOS event
	SOSEventStatusReported SOSEventStatus = "reported"
	// SOSEventStatusDispatched represents a dispatched SOS event
	SOSEventStatusDispatched SOSEventStatus = "dispatched"
	// SOSEventStatusResolved represents a resolved SOS event
	SOSEventStatusResolved SOSEventStatus = "resolved"
	// SOSEventStatusCancelled represents a cancelled SOS event
	SOSEventStatusCancelled SOSEventStatus = "cancelled"
)

// SOSEvent represents an emergency SOS event
type SOSEvent struct {
	ID          uuid.UUID      `json:"id"`
	MotherID    uuid.UUID      `json:"mother_id"`
	ReportedBy  uuid.UUID      `json:"reported_by"`
	Location    Location       `json:"location"`
	Nature      SOSEventNature `json:"nature"`
	Description string         `json:"description"`
	Status      SOSEventStatus `json:"status"`
	AmbulanceID *uuid.UUID     `json:"ambulance_id,omitempty"`
	FacilityID  *uuid.UUID     `json:"facility_id,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Priority    int            `json:"priority"`
	ETA         *time.Time     `json:"eta,omitempty"`
}

// NewSOSEvent creates a new SOS event
func NewSOSEvent(id, motherID, reportedBy uuid.UUID, lat, lng float64, nature SOSEventNature) *SOSEvent {
	now := time.Now()
	return &SOSEvent{
		ID:         id,
		MotherID:   motherID,
		ReportedBy: reportedBy,
		Location: Location{
			Latitude:  lat,
			Longitude: lng,
		},
		Nature:     nature,
		Status:     SOSEventStatusReported,
		CreatedAt:  now,
		UpdatedAt:  now,
		Priority:   calculatePriority(nature),
	}
}

// WithDescription adds a description to the SOS event
func (s *SOSEvent) WithDescription(description string) *SOSEvent {
	s.Description = description
	return s
}

// WithFacility assigns a facility to the SOS event
func (s *SOSEvent) WithFacility(facilityID uuid.UUID) *SOSEvent {
	s.FacilityID = &facilityID
	return s
}

// Dispatch assigns an ambulance to the SOS event
func (s *SOSEvent) Dispatch(ambulanceID uuid.UUID, eta time.Time) {
	s.AmbulanceID = &ambulanceID
	s.Status = SOSEventStatusDispatched
	s.ETA = &eta
	s.UpdatedAt = time.Now()
}

// Resolve marks the SOS event as resolved
func (s *SOSEvent) Resolve(facilityID uuid.UUID) {
	s.Status = SOSEventStatusResolved
	s.FacilityID = &facilityID
	s.UpdatedAt = time.Now()
}

// Cancel marks the SOS event as cancelled
func (s *SOSEvent) Cancel() {
	s.Status = SOSEventStatusCancelled
	s.UpdatedAt = time.Now()
}

// UpdatePriority updates the priority of the SOS event
func (s *SOSEvent) UpdatePriority(priority int) {
	s.Priority = priority
	s.UpdatedAt = time.Now()
}

// UpdateETA updates the estimated time of arrival
func (s *SOSEvent) UpdateETA(eta time.Time) {
	s.ETA = &eta
	s.UpdatedAt = time.Now()
}

// CalculateETAMinutes calculates the minutes until the estimated arrival
func (s *SOSEvent) CalculateETAMinutes(referenceTime time.Time) int {
	if s.ETA == nil {
		return -1
	}
	
	duration := s.ETA.Sub(referenceTime)
	minutes := int(duration.Minutes())
	
	if minutes < 0 {
		return 0
	}
	
	return minutes
}

// IsActive checks if the SOS event is active (reported or dispatched)
func (s *SOSEvent) IsActive() bool {
	return s.Status == SOSEventStatusReported || s.Status == SOSEventStatusDispatched
}

// Private function to calculate priority based on nature
func calculatePriority(nature SOSEventNature) int {
	// Higher number means higher priority
	switch nature {
	case SOSEventNatureLabor:
		return 3
	case SOSEventNatureBleeding:
		return 4
	case SOSEventNatureAccident:
		return 5
	default:
		return 2
	}
}
