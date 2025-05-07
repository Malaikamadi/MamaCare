package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/domain/repository"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// Service provides visit scheduling functionality
type Service struct {
	visitRepo     repository.VisitRepository
	motherRepo    repository.MotherRepository
	facilityRepo  repository.FacilityRepository
	notifyService NotificationService
	log           logger.Logger
}

// NotificationService defines the interface for sending visit notifications
type NotificationService interface {
	// SendVisitReminder sends a reminder for an upcoming visit
	SendVisitReminder(ctx context.Context, visit *model.Visit, motherID uuid.UUID, daysBeforeVisit int) error
}

// NewService creates a new visit scheduler service
func NewService(
	visitRepo repository.VisitRepository,
	motherRepo repository.MotherRepository,
	facilityRepo repository.FacilityRepository,
	notifyService NotificationService,
	log logger.Logger,
) *Service {
	return &Service{
		visitRepo:     visitRepo,
		motherRepo:    motherRepo,
		facilityRepo:  facilityRepo,
		notifyService: notifyService,
		log:           log,
	}
}

// ScheduleVisit schedules a new visit for a mother
func (s *Service) ScheduleVisit(
	ctx context.Context,
	motherID uuid.UUID,
	facilityID uuid.UUID,
	scheduledTime time.Time,
	visitType model.VisitType,
	notes string,
) (*model.Visit, error) {
	// Validate mother exists
	mother, err := s.motherRepo.GetByID(ctx, motherID)
	if err != nil {
		s.log.Error("Failed to find mother", logger.Fields{
			"error":     err.Error(),
			"mother_id": motherID.String(),
		})
		return nil, errorx.Wrap(err, "failed to find mother")
	}

	// Validate facility exists
	facility, err := s.facilityRepo.GetByID(ctx, facilityID)
	if err != nil {
		s.log.Error("Failed to find facility", logger.Fields{
			"error":       err.Error(),
			"facility_id": facilityID.String(),
		})
		return nil, errorx.Wrap(err, "failed to find facility")
	}

	// Check if scheduled time is in the future
	if scheduledTime.Before(time.Now()) {
		return nil, errorx.New(errorx.BadRequest, "scheduled time must be in the future")
	}

	// Create visit
	visitID := uuid.New()
	visit := model.NewVisit(visitID, motherID, facilityID, scheduledTime, visitType)
	
	if notes != "" {
		visit.WithNotes(notes)
	}

	// Save visit
	if err := s.visitRepo.Create(ctx, visit); err != nil {
		s.log.Error("Failed to create visit", logger.Fields{
			"error":       err.Error(),
			"mother_id":   motherID.String(),
			"facility_id": facilityID.String(),
		})
		return nil, errorx.Wrap(err, "failed to create visit")
	}

	s.log.Info("Visit scheduled successfully", logger.Fields{
		"visit_id":      visitID.String(),
		"mother_id":     motherID.String(),
		"facility_id":   facilityID.String(),
		"facility_name": facility.Name,
		"mother_name":   fmt.Sprintf("%s %s", mother.FirstName, mother.LastName),
		"scheduled_at":  scheduledTime.Format(time.RFC3339),
		"visit_type":    string(visitType),
	})

	return visit, nil
}

// RescheduleVisit reschedules an existing visit
func (s *Service) RescheduleVisit(
	ctx context.Context,
	visitID uuid.UUID,
	newScheduledTime time.Time,
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

	// Check if visit can be rescheduled
	if visit.Status != model.VisitStatusScheduled && visit.Status != model.VisitStatusCancelled {
		return nil, errorx.Newf(errorx.BadRequest, "cannot reschedule visit with status %s", visit.Status)
	}

	// Check if new scheduled time is in the future
	if newScheduledTime.Before(time.Now()) {
		return nil, errorx.New(errorx.BadRequest, "new scheduled time must be in the future")
	}

	// Reschedule visit
	visit.Reschedule(newScheduledTime)

	// Update visit
	if err := s.visitRepo.Update(ctx, visit); err != nil {
		s.log.Error("Failed to update visit", logger.Fields{
			"error":    err.Error(),
			"visit_id": visitID.String(),
		})
		return nil, errorx.Wrap(err, "failed to update visit")
	}

	s.log.Info("Visit rescheduled successfully", logger.Fields{
		"visit_id":        visitID.String(),
		"mother_id":       visit.MotherID.String(),
		"facility_id":     visit.FacilityID.String(),
		"new_scheduled_at": newScheduledTime.Format(time.RFC3339),
		"old_scheduled_at": visit.ScheduledTime.Format(time.RFC3339),
		"visit_type":      string(visit.VisitType),
	})

	return visit, nil
}

// CancelVisit cancels a scheduled visit
func (s *Service) CancelVisit(
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

	// Check if visit can be cancelled
	if visit.Status != model.VisitStatusScheduled {
		return nil, errorx.Newf(errorx.BadRequest, "cannot cancel visit with status %s", visit.Status)
	}

	// Cancel visit
	visit.Cancel()

	// Update visit
	if err := s.visitRepo.Update(ctx, visit); err != nil {
		s.log.Error("Failed to update visit", logger.Fields{
			"error":    err.Error(),
			"visit_id": visitID.String(),
		})
		return nil, errorx.Wrap(err, "failed to update visit")
	}

	s.log.Info("Visit cancelled successfully", logger.Fields{
		"visit_id":    visitID.String(),
		"mother_id":   visit.MotherID.String(),
		"facility_id": visit.FacilityID.String(),
		"scheduled_at": visit.ScheduledTime.Format(time.RFC3339),
		"visit_type":  string(visit.VisitType),
	})

	return visit, nil
}

// GetUpcomingVisits retrieves upcoming visits for a mother
func (s *Service) GetUpcomingVisits(
	ctx context.Context,
	motherID uuid.UUID,
	limit int,
) ([]*model.Visit, error) {
	options := repository.NewVisitQueryOptions().
		WithStatus(model.VisitStatusScheduled).
		WithLimit(limit).
		WithOrder("scheduled_time", "ASC")

	// Set date range from now to future
	now := time.Now()
	options.StartDate = &now

	visits, err := s.visitRepo.GetByMotherID(ctx, motherID, options)
	if err != nil {
		s.log.Error("Failed to get upcoming visits", logger.Fields{
			"error":     err.Error(),
			"mother_id": motherID.String(),
		})
		return nil, errorx.Wrap(err, "failed to get upcoming visits")
	}

	return visits, nil
}

// GetVisitHistory retrieves past visits for a mother
func (s *Service) GetVisitHistory(
	ctx context.Context,
	motherID uuid.UUID,
	limit int,
	offset int,
) ([]*model.Visit, error) {
	options := repository.NewVisitQueryOptions().
		WithLimit(limit).
		WithOffset(offset).
		WithOrder("scheduled_time", "DESC")

	// Set date range for past visits
	now := time.Now()
	options.EndDate = &now

	visits, err := s.visitRepo.GetByMotherID(ctx, motherID, options)
	if err != nil {
		s.log.Error("Failed to get visit history", logger.Fields{
			"error":     err.Error(),
			"mother_id": motherID.String(),
		})
		return nil, errorx.Wrap(err, "failed to get visit history")
	}

	return visits, nil
}

// ProcessReminders sends reminders for upcoming visits
func (s *Service) ProcessReminders(ctx context.Context) error {
	// Get visits in the next 3 days
	now := time.Now()
	threeDaysFromNow := now.Add(3 * 24 * time.Hour)
	
	options := repository.NewVisitQueryOptions().
		WithStatus(model.VisitStatusScheduled).
		WithDateRange(now, threeDaysFromNow).
		WithOrder("scheduled_time", "ASC").
		WithLimit(100) // Process in batches to avoid overwhelming the system

	visits, err := s.visitRepo.GetByDateRange(ctx, now, threeDaysFromNow, options)
	if err != nil {
		s.log.Error("Failed to get upcoming visits for reminders", logger.Fields{
			"error": err.Error(),
			"from":  now.Format(time.RFC3339),
			"to":    threeDaysFromNow.Format(time.RFC3339),
		})
		return errorx.Wrap(err, "failed to get upcoming visits for reminders")
	}

	var remindersSent int
	var errorsEncountered int

	for _, visit := range visits {
		// Calculate days until visit
		daysUntilVisit := int(visit.ScheduledTime.Sub(now).Hours() / 24)
		
		// Send reminder
		if err := s.notifyService.SendVisitReminder(ctx, visit, visit.MotherID, daysUntilVisit); err != nil {
			s.log.Error("Failed to send visit reminder", logger.Fields{
				"error":        err.Error(),
				"visit_id":     visit.ID.String(),
				"mother_id":    visit.MotherID.String(),
				"scheduled_at": visit.ScheduledTime.Format(time.RFC3339),
				"days_until":   daysUntilVisit,
			})
			errorsEncountered++
			continue
		}
		
		remindersSent++
	}

	s.log.Info("Processed visit reminders", logger.Fields{
		"total_visits":       len(visits),
		"reminders_sent":     remindersSent,
		"errors_encountered": errorsEncountered,
	})

	if errorsEncountered > 0 {
		return errorx.Newf(errorx.Internal, "encountered %d errors while sending reminders", errorsEncountered)
	}

	return nil
}

// GenerateAutomaticVisits generates visits based on pregnancy stage
func (s *Service) GenerateAutomaticVisits(
	ctx context.Context,
	motherID uuid.UUID,
	facilityID uuid.UUID,
) ([]*model.Visit, error) {
	// Get mother details
	mother, err := s.motherRepo.GetByID(ctx, motherID)
	if err != nil {
		s.log.Error("Failed to find mother", logger.Fields{
			"error":     err.Error(),
			"mother_id": motherID.String(),
		})
		return nil, errorx.Wrap(err, "failed to find mother")
	}

	// Check if mother is pregnant
	if mother.LMP == nil {
		return nil, errorx.New(errorx.BadRequest, "mother does not have a recorded LMP (Last Menstrual Period)")
	}

	// Check if facility exists
	_, err = s.facilityRepo.GetByID(ctx, facilityID)
	if err != nil {
		s.log.Error("Failed to find facility", logger.Fields{
			"error":       err.Error(),
			"facility_id": facilityID.String(),
		})
		return nil, errorx.Wrap(err, "failed to find facility")
	}

	// Calculate expected delivery date (EDD) - about 40 weeks after LMP
	edd := mother.LMP.AddDate(0, 0, 280) // 280 days = 40 weeks

	// Calculate current gestational age in weeks
	now := time.Now()
	gestationalAgeInDays := int(now.Sub(*mother.LMP).Hours() / 24)
	gestationalAgeInWeeks := gestationalAgeInDays / 7

	// Check if already delivered or too early in pregnancy
	if now.After(edd) {
		return nil, errorx.New(errorx.BadRequest, "expected delivery date has passed")
	}

	if gestationalAgeInWeeks < 8 {
		return nil, errorx.New(errorx.BadRequest, "too early in pregnancy to generate standard visit schedule")
	}

	// Define standard visit schedule based on WHO recommendations
	// Format: {week of pregnancy, visit type, notes}
	visitSchedule := []struct {
		Week      int
		VisitType model.VisitType
		Notes     string
	}{
		{12, model.VisitTypeRoutine, "First trimester checkup and basic tests"},
		{20, model.VisitTypeRoutine, "Second trimester checkup with ultrasound scan"},
		{26, model.VisitTypeRoutine, "Routine antenatal check"},
		{30, model.VisitTypeRoutine, "Third trimester follow-up"},
		{34, model.VisitTypeRoutine, "Pre-delivery preparation check"},
		{36, model.VisitTypeRoutine, "Late pregnancy follow-up"},
		{38, model.VisitTypeRoutine, "Pre-birth final check"},
		{40, model.VisitTypeRoutine, "Expected delivery week check"},
	}

	// Collect existing visits to avoid duplicates
	existingVisits, err := s.visitRepo.GetByMotherID(ctx, motherID, repository.NewVisitQueryOptions())
	if err != nil {
		s.log.Error("Failed to get existing visits", logger.Fields{
			"error":     err.Error(),
			"mother_id": motherID.String(),
		})
		return nil, errorx.Wrap(err, "failed to get existing visits")
	}

	// Create a map of existing visits by week of pregnancy
	existingVisitsByWeek := make(map[int]bool)
	for _, visit := range existingVisits {
		// Calculate which week of pregnancy this visit was scheduled for
		daysFromLMP := int(visit.ScheduledTime.Sub(*mother.LMP).Hours() / 24)
		weekOfPregnancy := daysFromLMP / 7
		existingVisitsByWeek[weekOfPregnancy] = true
	}

	// Generate new visits
	var newVisits []*model.Visit

	for _, scheduleItem := range visitSchedule {
		// Skip if visit for this week already exists
		if existingVisitsByWeek[scheduleItem.Week] {
			continue
		}

		// Skip if this week has already passed
		if scheduleItem.Week < gestationalAgeInWeeks {
			continue
		}

		// Calculate the date for this visit (LMP + weeks * 7 days)
		visitDate := mother.LMP.AddDate(0, 0, scheduleItem.Week*7)
		
		// Default visit hour (10:00 AM)
		visitDate = time.Date(
			visitDate.Year(), visitDate.Month(), visitDate.Day(),
			10, 0, 0, 0, visitDate.Location(),
		)

		// Schedule the visit
		visit, err := s.ScheduleVisit(
			ctx,
			motherID,
			facilityID,
			visitDate,
			scheduleItem.VisitType,
			scheduleItem.Notes,
		)
		
		if err != nil {
			s.log.Error("Failed to schedule automatic visit", logger.Fields{
				"error":             err.Error(),
				"mother_id":         motherID.String(),
				"facility_id":       facilityID.String(),
				"week_of_pregnancy": scheduleItem.Week,
			})
			continue
		}
		
		newVisits = append(newVisits, visit)
	}

	s.log.Info("Generated automatic visit schedule", logger.Fields{
		"mother_id":       motherID.String(),
		"facility_id":     facilityID.String(),
		"new_visits":      len(newVisits),
		"existing_visits": len(existingVisits),
		"edd":             edd.Format(time.RFC3339),
		"lmp":             mother.LMP.Format(time.RFC3339),
		"current_week":    gestationalAgeInWeeks,
	})

	return newVisits, nil
}

// GetVisitsByFacility retrieves visits for a facility
func (s *Service) GetVisitsByFacility(
	ctx context.Context,
	facilityID uuid.UUID,
	startDate, endDate time.Time,
	status *model.VisitStatus,
	limit, offset int,
) ([]*model.Visit, error) {
	options := repository.NewVisitQueryOptions().
		WithLimit(limit).
		WithOffset(offset).
		WithOrder("scheduled_time", "ASC").
		WithDateRange(startDate, endDate)

	if status != nil {
		options.WithStatus(*status)
	}

	visits, err := s.visitRepo.GetByFacilityID(ctx, facilityID, options)
	if err != nil {
		s.log.Error("Failed to get visits for facility", logger.Fields{
			"error":       err.Error(),
			"facility_id": facilityID.String(),
		})
		return nil, errorx.Wrap(err, "failed to get visits for facility")
	}

	return visits, nil
}

// GetVisitsByDate retrieves all visits for a specific date
func (s *Service) GetVisitsByDate(
	ctx context.Context,
	date time.Time,
	facilityID *uuid.UUID,
	status *model.VisitStatus,
	limit, offset int,
) ([]*model.Visit, error) {
	// Calculate start and end of the day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	options := repository.NewVisitQueryOptions().
		WithLimit(limit).
		WithOffset(offset).
		WithOrder("scheduled_time", "ASC").
		WithDateRange(startOfDay, endOfDay)

	if status != nil {
		options.WithStatus(*status)
	}

	var visits []*model.Visit
	var err error

	if facilityID != nil {
		visits, err = s.visitRepo.GetByFacilityID(ctx, *facilityID, options)
	} else {
		visits, err = s.visitRepo.GetByDateRange(ctx, startOfDay, endOfDay, options)
	}

	if err != nil {
		s.log.Error("Failed to get visits by date", logger.Fields{
			"error": err.Error(),
			"date":  date.Format("2006-01-02"),
		})
		return nil, errorx.Wrap(err, "failed to get visits by date")
	}

	return visits, nil
}

// FindAvailableSlots finds available time slots for a facility
func (s *Service) FindAvailableSlots(
	ctx context.Context,
	facilityID uuid.UUID,
	date time.Time,
	durationMinutes int,
) ([]time.Time, error) {
	// Get facility operating hours
	facility, err := s.facilityRepo.GetByID(ctx, facilityID)
	if err != nil {
		s.log.Error("Failed to find facility", logger.Fields{
			"error":       err.Error(),
			"facility_id": facilityID.String(),
		})
		return nil, errorx.Wrap(err, "failed to find facility")
	}

	// Default operating hours if not specified
	openingHour := 8  // 8:00 AM
	closingHour := 17 // 5:00 PM

	// Adjust based on facility operating hours if available
	if facility.OpeningHour != nil {
		openingHour = *facility.OpeningHour
	}
	if facility.ClosingHour != nil {
		closingHour = *facility.ClosingHour
	}

	// Check if date is in the past
	now := time.Now()
	if date.Before(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())) {
		return nil, errorx.New(errorx.BadRequest, "cannot find slots for past dates")
	}

	// Calculate start and end of the day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), openingHour, 0, 0, 0, date.Location())
	endOfDay := time.Date(date.Year(), date.Month(), date.Day(), closingHour, 0, 0, 0, date.Location())

	// If looking for today, adjust start time to now + 1 hour (buffer)
	if startOfDay.Year() == now.Year() && startOfDay.Month() == now.Month() && startOfDay.Day() == now.Day() {
		if now.Hour() >= openingHour {
			startOfDay = time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location())
		}
	}

	// Get existing visits for the day
	options := repository.NewVisitQueryOptions().
		WithDateRange(startOfDay, endOfDay)

	existingVisits, err := s.visitRepo.GetByFacilityID(ctx, facilityID, options)
	if err != nil {
		s.log.Error("Failed to get existing visits", logger.Fields{
			"error":       err.Error(),
			"facility_id": facilityID.String(),
			"date":        date.Format("2006-01-02"),
		})
		return nil, errorx.Wrap(err, "failed to get existing visits")
	}

	// Create a map of busy times
	busySlots := make(map[time.Time]bool)
	for _, visit := range existingVisits {
		// Mark the visit time as busy
		slotTime := visit.ScheduledTime
		busySlots[slotTime] = true
	}

	// Generate available slots every 30 minutes
	var availableSlots []time.Time
	slotDuration := 30 * time.Minute
	for slot := startOfDay; slot.Before(endOfDay); slot = slot.Add(slotDuration) {
		// Check if this slot is available
		if !busySlots[slot] {
			availableSlots = append(availableSlots, slot)
		}
	}

	return availableSlots, nil
}
