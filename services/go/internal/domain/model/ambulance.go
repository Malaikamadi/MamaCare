package model

import (
	"time"

	"github.com/google/uuid"
)

// AmbulanceStatus represents the current status of an ambulance
type AmbulanceStatus string

const (
	// AmbulanceStatusAvailable indicates the ambulance is available for dispatch
	AmbulanceStatusAvailable AmbulanceStatus = "available"
	// AmbulanceStatusDispatched indicates the ambulance has been dispatched
	AmbulanceStatusDispatched AmbulanceStatus = "dispatched"
	// AmbulanceStatusEnRoute indicates the ambulance is en route to the emergency
	AmbulanceStatusEnRoute AmbulanceStatus = "en_route"
	// AmbulanceStatusArrived indicates the ambulance has arrived at the emergency
	AmbulanceStatusArrived AmbulanceStatus = "arrived"
	// AmbulanceStatusReturning indicates the ambulance is returning with patient
	AmbulanceStatusReturning AmbulanceStatus = "returning"
	// AmbulanceStatusMaintenance indicates the ambulance is under maintenance
	AmbulanceStatusMaintenance AmbulanceStatus = "maintenance"
)

// AmbulanceType represents the type/capability of an ambulance
type AmbulanceType string

const (
	// AmbulanceTypeBasic represents a basic ambulance with minimal equipment
	AmbulanceTypeBasic AmbulanceType = "basic"
	// AmbulanceTypeAdvanced represents an advanced ambulance with more equipment
	AmbulanceTypeAdvanced AmbulanceType = "advanced"
	// AmbulanceTypeOB represents an ambulance equipped for obstetric emergencies
	AmbulanceTypeOB AmbulanceType = "obstetric"
)

// Ambulance represents an ambulance in the system
type Ambulance struct {
	ID            uuid.UUID        `json:"id"`
	CallSign      string           `json:"call_sign"`
	VehicleID     string           `json:"vehicle_id"`
	AmbulanceType AmbulanceType    `json:"ambulance_type"`
	Capacity      int              `json:"capacity"`
	Status        AmbulanceStatus  `json:"status"`
	CurrentSOSID  *uuid.UUID       `json:"current_sos_id,omitempty"`
	FacilityID    uuid.UUID        `json:"facility_id"` // Home facility
	Location      *Location        `json:"location,omitempty"`
	Crew          []uuid.UUID      `json:"crew"`
	LastUpdated   time.Time        `json:"last_updated"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// NewAmbulance creates a new ambulance
func NewAmbulance(id uuid.UUID, callSign, vehicleID string, ambulanceType AmbulanceType, facilityID uuid.UUID) *Ambulance {
	now := time.Now()
	return &Ambulance{
		ID:            id,
		CallSign:      callSign,
		VehicleID:     vehicleID,
		AmbulanceType: ambulanceType,
		Capacity:      calculateCapacity(ambulanceType),
		Status:        AmbulanceStatusAvailable,
		FacilityID:    facilityID,
		Crew:          make([]uuid.UUID, 0),
		LastUpdated:   now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// WithLocation adds a location to the ambulance
func (a *Ambulance) WithLocation(lat, lng float64) *Ambulance {
	a.Location = &Location{
		Latitude:  lat,
		Longitude: lng,
	}
	a.LastUpdated = time.Now()
	a.UpdatedAt = time.Now()
	return a
}

// AddCrewMember adds a crew member to the ambulance
func (a *Ambulance) AddCrewMember(staffID uuid.UUID) *Ambulance {
	a.Crew = append(a.Crew, staffID)
	a.UpdatedAt = time.Now()
	return a
}

// RemoveCrewMember removes a crew member from the ambulance
func (a *Ambulance) RemoveCrewMember(staffID uuid.UUID) *Ambulance {
	for i, id := range a.Crew {
		if id == staffID {
			a.Crew = append(a.Crew[:i], a.Crew[i+1:]...)
			break
		}
	}
	a.UpdatedAt = time.Now()
	return a
}

// Dispatch marks the ambulance as dispatched to an SOS event
func (a *Ambulance) Dispatch(sosID uuid.UUID) {
	a.Status = AmbulanceStatusDispatched
	a.CurrentSOSID = &sosID
	a.UpdatedAt = time.Now()
	a.LastUpdated = time.Now()
}

// MarkEnRoute marks the ambulance as en route to the emergency
func (a *Ambulance) MarkEnRoute() {
	a.Status = AmbulanceStatusEnRoute
	a.UpdatedAt = time.Now()
	a.LastUpdated = time.Now()
}

// MarkArrived marks the ambulance as arrived at the emergency
func (a *Ambulance) MarkArrived() {
	a.Status = AmbulanceStatusArrived
	a.UpdatedAt = time.Now()
	a.LastUpdated = time.Now()
}

// MarkReturning marks the ambulance as returning with a patient
func (a *Ambulance) MarkReturning() {
	a.Status = AmbulanceStatusReturning
	a.UpdatedAt = time.Now()
	a.LastUpdated = time.Now()
}

// MarkAvailable marks the ambulance as available
func (a *Ambulance) MarkAvailable() {
	a.Status = AmbulanceStatusAvailable
	a.CurrentSOSID = nil
	a.UpdatedAt = time.Now()
	a.LastUpdated = time.Now()
}

// MarkMaintenance marks the ambulance as under maintenance
func (a *Ambulance) MarkMaintenance() {
	a.Status = AmbulanceStatusMaintenance
	a.CurrentSOSID = nil
	a.UpdatedAt = time.Now()
	a.LastUpdated = time.Now()
}

// UpdateLocation updates the ambulance's location
func (a *Ambulance) UpdateLocation(lat, lng float64) {
	if a.Location == nil {
		a.Location = &Location{}
	}
	a.Location.Latitude = lat
	a.Location.Longitude = lng
	a.LastUpdated = time.Now()
	a.UpdatedAt = time.Now()
}

// IsAvailable checks if the ambulance is available for dispatch
func (a *Ambulance) IsAvailable() bool {
	return a.Status == AmbulanceStatusAvailable
}

// IsActive checks if the ambulance is currently on a call
func (a *Ambulance) IsActive() bool {
	return a.Status == AmbulanceStatusDispatched ||
		a.Status == AmbulanceStatusEnRoute ||
		a.Status == AmbulanceStatusArrived ||
		a.Status == AmbulanceStatusReturning
}

// calculateCapacity determines the capacity based on ambulance type
func calculateCapacity(ambulanceType AmbulanceType) int {
	switch ambulanceType {
	case AmbulanceTypeBasic:
		return 1
	case AmbulanceTypeAdvanced:
		return 2
	case AmbulanceTypeOB:
		return 2
	default:
		return 1
	}
}
