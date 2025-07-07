package model

import (
	"time"

	"github.com/google/uuid"
)

// VisitType represents the type of visit
type VisitType string

const (
	// VisitTypeRoutine represents a routine checkup
	VisitTypeRoutine VisitType = "routine"
	// VisitTypeEmergency represents an emergency visit
	VisitTypeEmergency VisitType = "emergency"
	// VisitTypeFollowUp represents a follow-up visit
	VisitTypeFollowUp VisitType = "follow_up"
)

// VisitStatus represents the status of a visit
type VisitStatus string

const (
	// VisitStatusScheduled represents a scheduled visit
	VisitStatusScheduled VisitStatus = "scheduled"
	// VisitStatusCheckedIn represents a visit where the mother has checked in
	VisitStatusCheckedIn VisitStatus = "checked_in"
	// VisitStatusCompleted represents a completed visit
	VisitStatusCompleted VisitStatus = "completed"
	// VisitStatusCancelled represents a cancelled visit
	VisitStatusCancelled VisitStatus = "cancelled"
)

// Visit represents a healthcare visit/appointment
type Visit struct {
	ID           uuid.UUID   `json:"id"`
	MotherID     uuid.UUID   `json:"mother_id"`
	FacilityID   uuid.UUID   `json:"facility_id"`
	CHWID        *uuid.UUID  `json:"chw_id,omitempty"`
	ClinicianID  *uuid.UUID  `json:"clinician_id,omitempty"`
	ScheduledTime time.Time   `json:"scheduled_time"`
	CheckInTime   *time.Time  `json:"check_in_time,omitempty"`
	CheckOutTime  *time.Time  `json:"check_out_time,omitempty"`
	VisitType     VisitType   `json:"visit_type"`
	VisitNotes    string      `json:"visit_notes,omitempty"`
	Status        VisitStatus `json:"status"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// NewVisit creates a new visit
func NewVisit(id, motherID, facilityID uuid.UUID, scheduledTime time.Time, visitType VisitType) *Visit {
	now := time.Now()
	return &Visit{
		ID:            id,
		MotherID:      motherID,
		FacilityID:    facilityID,
		ScheduledTime: scheduledTime,
		VisitType:     visitType,
		Status:        VisitStatusScheduled,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// WithCHW assigns a CHW to the visit
func (v *Visit) WithCHW(chwID uuid.UUID) *Visit {
	v.CHWID = &chwID
	return v
}

// WithClinician assigns a clinician to the visit
func (v *Visit) WithClinician(clinicianID uuid.UUID) *Visit {
	v.ClinicianID = &clinicianID
	return v
}

// WithNotes adds notes to the visit
func (v *Visit) WithNotes(notes string) *Visit {
	v.VisitNotes = notes
	return v
}

// CheckIn marks the visit as checked in
func (v *Visit) CheckIn() {
	now := time.Now()
	v.CheckInTime = &now
	v.Status = VisitStatusCheckedIn
	v.UpdatedAt = now
}

// Complete marks the visit as completed
func (v *Visit) Complete(notes string) {
	now := time.Now()
	v.CheckOutTime = &now
	v.Status = VisitStatusCompleted
	v.VisitNotes = notes
	v.UpdatedAt = now
}

// Cancel marks the visit as cancelled
func (v *Visit) Cancel() {
	now := time.Now()
	v.Status = VisitStatusCancelled
	v.UpdatedAt = now
}

// Reschedule changes the scheduled time of the visit
func (v *Visit) Reschedule(newTime time.Time) {
	v.ScheduledTime = newTime
	v.UpdatedAt = time.Now()
	
	// If the visit was cancelled, mark it as scheduled again
	if v.Status == VisitStatusCancelled {
		v.Status = VisitStatusScheduled
	}
}

// IsUpcoming checks if the visit is upcoming
func (v *Visit) IsUpcoming(referenceTime time.Time) bool {
	return v.Status == VisitStatusScheduled && v.ScheduledTime.After(referenceTime)
}

// IsMissed checks if the visit was missed
func (v *Visit) IsMissed(referenceTime time.Time) bool {
	// A visit is considered missed if it's still scheduled but the scheduled time has passed
	return v.Status == VisitStatusScheduled && v.ScheduledTime.Before(referenceTime)
}

// GetDuration returns the duration of the visit if completed
func (v *Visit) GetDuration() *time.Duration {
	if v.CheckInTime != nil && v.CheckOutTime != nil {
		duration := v.CheckOutTime.Sub(*v.CheckInTime)
		return &duration
	}
	return nil
}
