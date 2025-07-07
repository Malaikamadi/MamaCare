package status

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/domain/repository"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// Service provides visit status tracking functionality
type Service struct {
	visitRepo    repository.VisitRepository
	motherRepo   repository.MotherRepository
	userRepo     repository.UserRepository
	facilityRepo repository.FacilityRepository
	log          logger.Logger
}

// NewService creates a new visit status service
func NewService(
	visitRepo repository.VisitRepository,
	motherRepo repository.MotherRepository,
	userRepo repository.UserRepository,
	facilityRepo repository.FacilityRepository,
	log logger.Logger,
) *Service {
	return &Service{
		visitRepo:    visitRepo,
		motherRepo:   motherRepo,
		userRepo:     userRepo,
		facilityRepo: facilityRepo,
		log:          log,
	}
}

// CheckInVisit marks a visit as checked in
func (s *Service) CheckInVisit(
	ctx context.Context,
	visitID uuid.UUID,
) (*model.Visit, error) {
	// Retrieve visit
	visit, err := s.visitRepo.GetByID(ctx, visitID)
	if err != nil {
		s.log.Error("Failed to find visit", logger.Fields{
			"error":    err.Error(),
			"visit_id": visitID.String(),
		})
		return nil, errorx.Wrap(err, "failed to find visit")
	}

	// Check if visit can be checked in
	if visit.Status != model.VisitStatusScheduled {
		return nil, errorx.Newf(errorx.BadRequest, "cannot check in visit with status %s", visit.Status)
	}

	// Check in visit
	visit.CheckIn()

	// Update visit
	if err := s.visitRepo.Update(ctx, visit); err != nil {
		s.log.Error("Failed to update visit", logger.Fields{
			"error":    err.Error(),
			"visit_id": visitID.String(),
		})
		return nil, errorx.Wrap(err, "failed to update visit")
	}

	s.log.Info("Visit checked in successfully", logger.Fields{
		"visit_id":     visitID.String(),
		"mother_id":    visit.MotherID.String(),
		"facility_id":  visit.FacilityID.String(),
		"check_in_time": visit.CheckInTime.Format(time.RFC3339),
		"visit_type":   string(visit.VisitType),
	})

	return visit, nil
}

// CompleteVisit marks a visit as completed
func (s *Service) CompleteVisit(
	ctx context.Context,
	visitID uuid.UUID,
	notes string,
) (*model.Visit, error) {
	// Retrieve visit
	visit, err := s.visitRepo.GetByID(ctx, visitID)
	if err != nil {
		s.log.Error("Failed to find visit", logger.Fields{
			"error":    err.Error(),
			"visit_id": visitID.String(),
		})
		return nil, errorx.Wrap(err, "failed to find visit")
	}

	// Check if visit can be completed
	if visit.Status != model.VisitStatusCheckedIn {
		return nil, errorx.Newf(errorx.BadRequest, "cannot complete visit with status %s, must be checked in first", visit.Status)
	}

	// Complete visit
	visit.Complete(notes)

	// Update visit
	if err := s.visitRepo.Update(ctx, visit); err != nil {
		s.log.Error("Failed to update visit", logger.Fields{
			"error":    err.Error(),
			"visit_id": visitID.String(),
		})
		return nil, errorx.Wrap(err, "failed to update visit")
	}

	s.log.Info("Visit completed successfully", logger.Fields{
		"visit_id":      visitID.String(),
		"mother_id":     visit.MotherID.String(),
		"facility_id":   visit.FacilityID.String(),
		"check_in_time": visit.CheckInTime.Format(time.RFC3339),
		"check_out_time": visit.CheckOutTime.Format(time.RFC3339),
		"visit_type":    string(visit.VisitType),
	})

	return visit, nil
}

// MarkMissedVisits identifies and marks visits that were missed
func (s *Service) MarkMissedVisits(ctx context.Context) (int, error) {
	// Get current time as reference
	now := time.Now()
	
	// Get yesterday's date (to ensure we don't mark today's visits as missed)
	yesterday := now.AddDate(0, 0, -1)
	endOfYesterday := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, yesterday.Location())
	
	// Create a date 7 days ago for limiting the search window
	sevenDaysAgo := now.AddDate(0, 0, -7)
	
	// Find scheduled visits between 7 days ago and yesterday that weren't checked in
	options := repository.NewVisitQueryOptions().
		WithStatus(model.VisitStatusScheduled).
		WithDateRange(sevenDaysAgo, endOfYesterday).
		WithLimit(100) // Process in batches

	visits, err := s.visitRepo.GetByDateRange(ctx, sevenDaysAgo, endOfYesterday, options)
	if err != nil {
		s.log.Error("Failed to get scheduled visits to mark as missed", logger.Fields{
			"error": err.Error(),
			"from":  sevenDaysAgo.Format(time.RFC3339),
			"to":    endOfYesterday.Format(time.RFC3339),
		})
		return 0, errorx.Wrap(err, "failed to get scheduled visits")
	}

	var missedCount int
	var errorCount int

	// Define a custom status for missed visits (we'll use a helper function to update the model)
	missedStatus := "missed"

	// Identify and mark missed visits
	for _, visit := range visits {
		// Skip if already cancelled
		if visit.Status == model.VisitStatusCancelled {
			continue
		}
		
		// Mark as missed
		visit.Status = model.VisitStatusCancelled // For now we'll use cancelled, but add a note
		visit.VisitNotes = "Automatically marked as missed: " + visit.VisitNotes
		visit.UpdatedAt = now
		
		// Update in database
		if err := s.visitRepo.Update(ctx, visit); err != nil {
			s.log.Error("Failed to mark visit as missed", logger.Fields{
				"error":    err.Error(),
				"visit_id": visit.ID.String(),
				"mother_id": visit.MotherID.String(),
			})
			errorCount++
			continue
		}
		
		missedCount++
	}

	s.log.Info("Marked missed visits", logger.Fields{
		"total_visits": len(visits),
		"missed_count": missedCount,
		"error_count":  errorCount,
		"date_range":   sevenDaysAgo.Format("2006-01-02") + " to " + endOfYesterday.Format("2006-01-02"),
	})

	if errorCount > 0 {
		return missedCount, errorx.Newf(errorx.Internal, "encountered %d errors while marking missed visits", errorCount)
	}

	return missedCount, nil
}

// GetVisitStatus retrieves the detailed status of a visit
func (s *Service) GetVisitStatus(
	ctx context.Context,
	visitID uuid.UUID,
) (*VisitStatusDetails, error) {
	// Retrieve visit
	visit, err := s.visitRepo.GetByID(ctx, visitID)
	if err != nil {
		s.log.Error("Failed to find visit", logger.Fields{
			"error":    err.Error(),
			"visit_id": visitID.String(),
		})
		return nil, errorx.Wrap(err, "failed to find visit")
	}

	// Retrieve mother
	mother, err := s.motherRepo.GetByID(ctx, visit.MotherID)
	if err != nil {
		s.log.Error("Failed to find mother", logger.Fields{
			"error":     err.Error(),
			"mother_id": visit.MotherID.String(),
		})
		return nil, errorx.Wrap(err, "failed to find mother")
	}

	// Retrieve facility
	facility, err := s.facilityRepo.GetByID(ctx, visit.FacilityID)
	if err != nil {
		s.log.Error("Failed to find facility", logger.Fields{
			"error":       err.Error(),
			"facility_id": visit.FacilityID.String(),
		})
		return nil, errorx.Wrap(err, "failed to find facility")
	}

	// Create status details
	details := &VisitStatusDetails{
		Visit:           visit,
		MotherName:      mother.FirstName + " " + mother.LastName,
		FacilityName:    facility.Name,
		TimeSinceUpdate: time.Since(visit.UpdatedAt),
	}

	// Check if visit is upcoming or in the past
	now := time.Now()
	details.IsUpcoming = visit.ScheduledTime.After(now)
	details.IsMissed = visit.IsMissed(now)
	
	// Calculate duration if available
	if visit.CheckInTime != nil && visit.CheckOutTime != nil {
		duration := visit.CheckOutTime.Sub(*visit.CheckInTime)
		details.Duration = &duration
	}

	return details, nil
}

// GetOverdueVisits retrieves visits that should have been completed already
func (s *Service) GetOverdueVisits(
	ctx context.Context,
	limit, offset int,
) ([]*model.Visit, error) {
	// Get current time as reference
	now := time.Now()
	
	// Create a date for the start of the window (e.g., 30 days ago to avoid very old visits)
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	
	// Find scheduled visits between 30 days ago and now that weren't checked in
	options := repository.NewVisitQueryOptions().
		WithStatus(model.VisitStatusScheduled).
		WithDateRange(thirtyDaysAgo, now).
		WithLimit(limit).
		WithOffset(offset).
		WithOrder("scheduled_time", "ASC")

	visits, err := s.visitRepo.GetByDateRange(ctx, thirtyDaysAgo, now, options)
	if err != nil {
		s.log.Error("Failed to get overdue visits", logger.Fields{
			"error": err.Error(),
			"from":  thirtyDaysAgo.Format(time.RFC3339),
			"to":    now.Format(time.RFC3339),
		})
		return nil, errorx.Wrap(err, "failed to get overdue visits")
	}

	return visits, nil
}

// ScheduleFollowUp schedules a follow-up visit
func (s *Service) ScheduleFollowUp(
	ctx context.Context,
	originalVisitID uuid.UUID,
	scheduledTime time.Time,
	notes string,
) (*model.Visit, error) {
	// Retrieve original visit
	originalVisit, err := s.visitRepo.GetByID(ctx, originalVisitID)
	if err != nil {
		s.log.Error("Failed to find original visit", logger.Fields{
			"error":    err.Error(),
			"visit_id": originalVisitID.String(),
		})
		return nil, errorx.Wrap(err, "failed to find original visit")
	}

	// Check if scheduled time is in the future
	if scheduledTime.Before(time.Now()) {
		return nil, errorx.New(errorx.BadRequest, "scheduled time must be in the future")
	}

	// Create follow-up visit
	visitID := uuid.New()
	visit := model.NewVisit(
		visitID,
		originalVisit.MotherID,
		originalVisit.FacilityID,
		scheduledTime,
		model.VisitTypeFollowUp,
	)
	
	// Copy CHW and clinician assignments if available
	if originalVisit.CHWID != nil {
		visit.WithCHW(*originalVisit.CHWID)
	}
	
	if originalVisit.ClinicianID != nil {
		visit.WithClinician(*originalVisit.ClinicianID)
	}
	
	// Add notes referencing original visit
	followUpNotes := "Follow-up to visit on " + originalVisit.ScheduledTime.Format("2006-01-02") + ". " + notes
	visit.WithNotes(followUpNotes)

	// Save visit
	if err := s.visitRepo.Create(ctx, visit); err != nil {
		s.log.Error("Failed to create follow-up visit", logger.Fields{
			"error":           err.Error(),
			"mother_id":       originalVisit.MotherID.String(),
			"original_visit_id": originalVisitID.String(),
		})
		return nil, errorx.Wrap(err, "failed to create follow-up visit")
	}

	s.log.Info("Follow-up visit scheduled successfully", logger.Fields{
		"visit_id":         visitID.String(),
		"mother_id":        originalVisit.MotherID.String(),
		"facility_id":      originalVisit.FacilityID.String(),
		"scheduled_at":     scheduledTime.Format(time.RFC3339),
		"original_visit_id": originalVisitID.String(),
	})

	return visit, nil
}

// VisitStatusDetails contains detailed status information for a visit
type VisitStatusDetails struct {
	Visit           *model.Visit `json:"visit"`
	MotherName      string       `json:"mother_name"`
	FacilityName    string       `json:"facility_name"`
	IsUpcoming      bool         `json:"is_upcoming"`
	IsMissed        bool         `json:"is_missed"`
	Duration        *time.Duration `json:"duration,omitempty"`
	TimeSinceUpdate time.Duration `json:"time_since_update"`
}

// GetCompletedVisits retrieves completed visits for reporting
func (s *Service) GetCompletedVisits(
	ctx context.Context,
	startDate, endDate time.Time,
	facilityID *uuid.UUID,
	limit, offset int,
) ([]*model.Visit, error) {
	options := repository.NewVisitQueryOptions().
		WithStatus(model.VisitStatusCompleted).
		WithDateRange(startDate, endDate).
		WithLimit(limit).
		WithOffset(offset).
		WithOrder("scheduled_time", "DESC")

	var visits []*model.Visit
	var err error

	if facilityID != nil {
		visits, err = s.visitRepo.GetByFacilityID(ctx, *facilityID, options)
	} else {
		visits, err = s.visitRepo.GetByDateRange(ctx, startDate, endDate, options)
	}

	if err != nil {
		s.log.Error("Failed to get completed visits", logger.Fields{
			"error":      err.Error(),
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
		})
		return nil, errorx.Wrap(err, "failed to get completed visits")
	}

	return visits, nil
}

// UpdateVisitNotes updates the notes for a visit
func (s *Service) UpdateVisitNotes(
	ctx context.Context,
	visitID uuid.UUID,
	notes string,
) (*model.Visit, error) {
	// Retrieve visit
	visit, err := s.visitRepo.GetByID(ctx, visitID)
	if err != nil {
		s.log.Error("Failed to find visit", logger.Fields{
			"error":    err.Error(),
			"visit_id": visitID.String(),
		})
		return nil, errorx.Wrap(err, "failed to find visit")
	}

	// Update notes
	visit.VisitNotes = notes
	visit.UpdatedAt = time.Now()

	// Update visit
	if err := s.visitRepo.Update(ctx, visit); err != nil {
		s.log.Error("Failed to update visit notes", logger.Fields{
			"error":    err.Error(),
			"visit_id": visitID.String(),
		})
		return nil, errorx.Wrap(err, "failed to update visit notes")
	}

	s.log.Info("Visit notes updated successfully", logger.Fields{
		"visit_id":  visitID.String(),
		"mother_id": visit.MotherID.String(),
	})

	return visit, nil
}
