package report

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

// Service provides functionality for generating visit reports
type Service struct {
	visitRepo    repository.VisitRepository
	motherRepo   repository.MotherRepository
	facilityRepo repository.FacilityRepository
	userRepo     repository.UserRepository
	log          logger.Logger
}

// NewService creates a new visit report service
func NewService(
	visitRepo repository.VisitRepository,
	motherRepo repository.MotherRepository,
	facilityRepo repository.FacilityRepository,
	userRepo repository.UserRepository,
	log logger.Logger,
) *Service {
	return &Service{
		visitRepo:    visitRepo,
		motherRepo:   motherRepo,
		facilityRepo: facilityRepo,
		userRepo:     userRepo,
		log:          log,
	}
}

// GenerateVisitReport generates a detailed report for a specific visit
func (s *Service) GenerateVisitReport(
	ctx context.Context,
	visitID uuid.UUID,
) (*VisitReport, error) {
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

	// Create report
	report := &VisitReport{
		Visit:          visit,
		MotherName:     fmt.Sprintf("%s %s", mother.FirstName, mother.LastName),
		MotherAge:      calculateAge(mother.DateOfBirth),
		MotherContact:  mother.PhoneNumber,
		FacilityName:   facility.Name,
		FacilityType:   facility.Type,
		FacilityContact: facility.PhoneNumber,
		GeneratedAt:    time.Now(),
	}

	// Add CHW and clinician info if available
	if visit.CHWID != nil {
		chw, err := s.userRepo.GetByID(ctx, *visit.CHWID)
		if err == nil {
			report.CHWName = fmt.Sprintf("%s %s", chw.FirstName, chw.LastName)
			report.CHWContact = chw.PhoneNumber
		}
	}

	if visit.ClinicianID != nil {
		clinician, err := s.userRepo.GetByID(ctx, *visit.ClinicianID)
		if err == nil {
			report.ClinicianName = fmt.Sprintf("%s %s", clinician.FirstName, clinician.LastName)
			report.ClinicianRole = clinician.Role
		}
	}

	// Add pregnancy details if available
	if mother.LMP != nil {
		report.LastMenstrualPeriod = *mother.LMP
		report.EstimatedDeliveryDate = mother.LMP.AddDate(0, 0, 280) // 280 days = 40 weeks
		
		// Calculate gestational age in weeks
		gestationalAgeInDays := int(time.Since(*mother.LMP).Hours() / 24)
		report.GestationalAge = gestationalAgeInDays / 7
	}

	// Add visit timeline if visit has check-in and check-out
	if visit.CheckInTime != nil {
		report.CheckInTime = *visit.CheckInTime
		
		if visit.CheckOutTime != nil {
			report.CheckOutTime = *visit.CheckOutTime
			report.VisitDuration = visit.CheckOutTime.Sub(*visit.CheckInTime)
		}
	}

	s.log.Info("Generated visit report", logger.Fields{
		"visit_id":     visitID.String(),
		"mother_id":    visit.MotherID.String(),
		"facility_id":  visit.FacilityID.String(),
		"visit_status": string(visit.Status),
	})

	return report, nil
}

// GenerateFacilitySummary generates a summary report for visits at a facility
func (s *Service) GenerateFacilitySummary(
	ctx context.Context,
	facilityID uuid.UUID,
	startDate, endDate time.Time,
) (*FacilitySummaryReport, error) {
	// Retrieve facility
	facility, err := s.facilityRepo.GetByID(ctx, facilityID)
	if err != nil {
		s.log.Error("Failed to find facility", logger.Fields{
			"error":       err.Error(),
			"facility_id": facilityID.String(),
		})
		return nil, errorx.Wrap(err, "failed to find facility")
	}

	// Get all visits for the facility in the date range
	options := repository.NewVisitQueryOptions().
		WithDateRange(startDate, endDate)

	visits, err := s.visitRepo.GetByFacilityID(ctx, facilityID, options)
	if err != nil {
		s.log.Error("Failed to get facility visits", logger.Fields{
			"error":       err.Error(),
			"facility_id": facilityID.String(),
			"start_date":  startDate.Format(time.RFC3339),
			"end_date":    endDate.Format(time.RFC3339),
		})
		return nil, errorx.Wrap(err, "failed to get facility visits")
	}

	// Create summary report
	summary := &FacilitySummaryReport{
		FacilityID:      facilityID,
		FacilityName:    facility.Name,
		StartDate:       startDate,
		EndDate:         endDate,
		GeneratedAt:     time.Now(),
		TotalVisits:     len(visits),
		VisitsByType:    make(map[model.VisitType]int),
		VisitsByStatus:  make(map[model.VisitStatus]int),
		VisitsByDay:     make(map[string]int),
		AverageVisitDuration: 0,
	}

	// Counters
	var completedVisitsCount int
	var totalVisitDuration time.Duration

	// Analyze visits
	for _, visit := range visits {
		// Count by type
		summary.VisitsByType[visit.VisitType]++
		
		// Count by status
		summary.VisitsByStatus[visit.Status]++
		
		// Count by day
		dayKey := visit.ScheduledTime.Format("2006-01-02")
		summary.VisitsByDay[dayKey]++
		
		// Calculate visit duration if completed
		if visit.Status == model.VisitStatusCompleted && visit.CheckInTime != nil && visit.CheckOutTime != nil {
			duration := visit.CheckOutTime.Sub(*visit.CheckInTime)
			totalVisitDuration += duration
			completedVisitsCount++
		}
	}

	// Calculate average visit duration
	if completedVisitsCount > 0 {
		summary.AverageVisitDuration = totalVisitDuration.Minutes() / float64(completedVisitsCount)
	}

	// Calculate completion rate
	if summary.TotalVisits > 0 {
		summary.CompletionRate = float64(summary.VisitsByStatus[model.VisitStatusCompleted]) / float64(summary.TotalVisits) * 100
	}

	s.log.Info("Generated facility summary report", logger.Fields{
		"facility_id": facilityID.String(),
		"start_date":  startDate.Format("2006-01-02"),
		"end_date":    endDate.Format("2006-01-02"),
		"total_visits": summary.TotalVisits,
	})

	return summary, nil
}

// GenerateCHWSummary generates a summary report for a CHW's visits
func (s *Service) GenerateCHWSummary(
	ctx context.Context,
	chwID uuid.UUID,
	startDate, endDate time.Time,
) (*CHWSummaryReport, error) {
	// Retrieve CHW
	chw, err := s.userRepo.GetByID(ctx, chwID)
	if err != nil {
		s.log.Error("Failed to find CHW", logger.Fields{
			"error":  err.Error(),
			"chw_id": chwID.String(),
		})
		return nil, errorx.Wrap(err, "failed to find CHW")
	}

	// Get all visits for the CHW in the date range
	options := repository.NewVisitQueryOptions().
		WithDateRange(startDate, endDate)

	visits, err := s.visitRepo.GetByCHW(ctx, chwID, options)
	if err != nil {
		s.log.Error("Failed to get CHW visits", logger.Fields{
			"error":      err.Error(),
			"chw_id":     chwID.String(),
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
		})
		return nil, errorx.Wrap(err, "failed to get CHW visits")
	}

	// Create summary report
	summary := &CHWSummaryReport{
		CHWID:            chwID,
		CHWName:          fmt.Sprintf("%s %s", chw.FirstName, chw.LastName),
		StartDate:        startDate,
		EndDate:          endDate,
		GeneratedAt:      time.Now(),
		TotalVisits:      len(visits),
		CompletedVisits:  0,
		CancelledVisits:  0,
		MissedVisits:     0,
		VisitsByDay:      make(map[string]int),
		AverageVisitDuration: 0,
	}

	// Counters
	var totalVisitDuration time.Duration
	var completedVisitsCount int

	// Get current time for reference
	now := time.Now()

	// Analyze visits
	for _, visit := range visits {
		// Count by day
		dayKey := visit.ScheduledTime.Format("2006-01-02")
		summary.VisitsByDay[dayKey]++
		
		// Count by status
		switch visit.Status {
		case model.VisitStatusCompleted:
			summary.CompletedVisits++
		case model.VisitStatusCancelled:
			summary.CancelledVisits++
		case model.VisitStatusScheduled:
			// Check if visit was missed
			if visit.ScheduledTime.Before(now) {
				summary.MissedVisits++
			}
		}
		
		// Calculate visit duration if completed
		if visit.Status == model.VisitStatusCompleted && visit.CheckInTime != nil && visit.CheckOutTime != nil {
			duration := visit.CheckOutTime.Sub(*visit.CheckInTime)
			totalVisitDuration += duration
			completedVisitsCount++
		}
	}

	// Calculate average visit duration
	if completedVisitsCount > 0 {
		summary.AverageVisitDuration = totalVisitDuration.Minutes() / float64(completedVisitsCount)
	}

	// Calculate completion rate
	if summary.TotalVisits > 0 {
		summary.CompletionRate = float64(summary.CompletedVisits) / float64(summary.TotalVisits) * 100
	}

	s.log.Info("Generated CHW summary report", logger.Fields{
		"chw_id":         chwID.String(),
		"start_date":     startDate.Format("2006-01-02"),
		"end_date":       endDate.Format("2006-01-02"),
		"total_visits":   summary.TotalVisits,
		"completed_visits": summary.CompletedVisits,
	})

	return summary, nil
}

// GenerateDistrictSummary generates a summary report for a district
func (s *Service) GenerateDistrictSummary(
	ctx context.Context,
	district string,
	startDate, endDate time.Time,
) (*DistrictSummaryReport, error) {
	// Create summary report
	summary := &DistrictSummaryReport{
		District:          district,
		StartDate:         startDate,
		EndDate:           endDate,
		GeneratedAt:       time.Now(),
		TotalVisits:       0,
		TotalMothers:      0,
		VisitsByFacility:  make(map[uuid.UUID]int),
		FacilityNames:     make(map[uuid.UUID]string),
		VisitsByType:      make(map[model.VisitType]int),
		VisitsByStatus:    make(map[model.VisitStatus]int),
		CompletionRates:   make(map[uuid.UUID]float64),
	}

	// TODO: Get facilities by district
	// This would require adding district to the Facility model and implementation
	// For now, we'll simulate with hardcoded facility IDs
	facilityIDs := []uuid.UUID{
		uuid.New(),
		uuid.New(),
	}

	// Track unique mothers
	uniqueMothers := make(map[uuid.UUID]bool)

	// Get visits for each facility
	for _, facilityID := range facilityIDs {
		// Get facility name (simulated)
		facility, err := s.facilityRepo.GetByID(ctx, facilityID)
		if err == nil {
			summary.FacilityNames[facilityID] = facility.Name
		} else {
			summary.FacilityNames[facilityID] = "Unknown Facility"
		}

		// Get visits for this facility
		options := repository.NewVisitQueryOptions().
			WithDateRange(startDate, endDate)

		visits, err := s.visitRepo.GetByFacilityID(ctx, facilityID, options)
		if err != nil {
			s.log.Error("Failed to get facility visits", logger.Fields{
				"error":       err.Error(),
				"facility_id": facilityID.String(),
				"district":    district,
			})
			continue
		}

		// Track facility statistics
		summary.VisitsByFacility[facilityID] = len(visits)
		summary.TotalVisits += len(visits)
		
		var facilityCompletedVisits int
		
		for _, visit := range visits {
			// Track unique mothers
			uniqueMothers[visit.MotherID] = true
			
			// Count by type and status
			summary.VisitsByType[visit.VisitType]++
			summary.VisitsByStatus[visit.Status]++
			
			// Count completed visits per facility
			if visit.Status == model.VisitStatusCompleted {
				facilityCompletedVisits++
			}
		}
		
		// Calculate facility completion rate
		if len(visits) > 0 {
			summary.CompletionRates[facilityID] = float64(facilityCompletedVisits) / float64(len(visits)) * 100
		}
	}

	// Count unique mothers
	summary.TotalMothers = len(uniqueMothers)

	// Calculate district completion rate
	if summary.TotalVisits > 0 {
		summary.OverallCompletionRate = float64(summary.VisitsByStatus[model.VisitStatusCompleted]) / float64(summary.TotalVisits) * 100
	}

	s.log.Info("Generated district summary report", logger.Fields{
		"district":     district,
		"start_date":   startDate.Format("2006-01-02"),
		"end_date":     endDate.Format("2006-01-02"),
		"total_visits": summary.TotalVisits,
		"facilities":   len(facilityIDs),
		"total_mothers": summary.TotalMothers,
	})

	return summary, nil
}

// VisitReport represents a detailed report for a single visit
type VisitReport struct {
	Visit              *model.Visit `json:"visit"`
	MotherName         string       `json:"mother_name"`
	MotherAge          int          `json:"mother_age"`
	MotherContact      string       `json:"mother_contact"`
	FacilityName       string       `json:"facility_name"`
	FacilityType       string       `json:"facility_type"`
	FacilityContact    string       `json:"facility_contact"`
	CHWName            string       `json:"chw_name,omitempty"`
	CHWContact         string       `json:"chw_contact,omitempty"`
	ClinicianName      string       `json:"clinician_name,omitempty"`
	ClinicianRole      string       `json:"clinician_role,omitempty"`
	LastMenstrualPeriod time.Time   `json:"last_menstrual_period,omitempty"`
	EstimatedDeliveryDate time.Time `json:"estimated_delivery_date,omitempty"`
	GestationalAge     int          `json:"gestational_age,omitempty"`
	CheckInTime        time.Time    `json:"check_in_time,omitempty"`
	CheckOutTime       time.Time    `json:"check_out_time,omitempty"`
	VisitDuration      time.Duration `json:"visit_duration,omitempty"`
	GeneratedAt        time.Time    `json:"generated_at"`
}

// FacilitySummaryReport represents a summary report for a facility
type FacilitySummaryReport struct {
	FacilityID         uuid.UUID               `json:"facility_id"`
	FacilityName       string                  `json:"facility_name"`
	StartDate          time.Time               `json:"start_date"`
	EndDate            time.Time               `json:"end_date"`
	GeneratedAt        time.Time               `json:"generated_at"`
	TotalVisits        int                     `json:"total_visits"`
	VisitsByType       map[model.VisitType]int `json:"visits_by_type"`
	VisitsByStatus     map[model.VisitStatus]int `json:"visits_by_status"`
	VisitsByDay        map[string]int          `json:"visits_by_day"`
	AverageVisitDuration float64               `json:"average_visit_duration_minutes"`
	CompletionRate     float64                 `json:"completion_rate_percentage"`
}

// CHWSummaryReport represents a summary report for a CHW
type CHWSummaryReport struct {
	CHWID              uuid.UUID    `json:"chw_id"`
	CHWName            string       `json:"chw_name"`
	StartDate          time.Time    `json:"start_date"`
	EndDate            time.Time    `json:"end_date"`
	GeneratedAt        time.Time    `json:"generated_at"`
	TotalVisits        int          `json:"total_visits"`
	CompletedVisits    int          `json:"completed_visits"`
	CancelledVisits    int          `json:"cancelled_visits"`
	MissedVisits       int          `json:"missed_visits"`
	VisitsByDay        map[string]int `json:"visits_by_day"`
	AverageVisitDuration float64     `json:"average_visit_duration_minutes"`
	CompletionRate     float64       `json:"completion_rate_percentage"`
}

// DistrictSummaryReport represents a summary report for a district
type DistrictSummaryReport struct {
	District           string                  `json:"district"`
	StartDate          time.Time               `json:"start_date"`
	EndDate            time.Time               `json:"end_date"`
	GeneratedAt        time.Time               `json:"generated_at"`
	TotalVisits        int                     `json:"total_visits"`
	TotalMothers       int                     `json:"total_mothers"`
	VisitsByFacility   map[uuid.UUID]int       `json:"visits_by_facility"`
	FacilityNames      map[uuid.UUID]string    `json:"facility_names"`
	VisitsByType       map[model.VisitType]int `json:"visits_by_type"`
	VisitsByStatus     map[model.VisitStatus]int `json:"visits_by_status"`
	CompletionRates    map[uuid.UUID]float64   `json:"completion_rates_by_facility"`
	OverallCompletionRate float64              `json:"overall_completion_rate_percentage"`
}

// calculateAge calculates age from birthdate
func calculateAge(birthdate time.Time) int {
	now := time.Now()
	years := now.Year() - birthdate.Year()
	
	// Adjust for months/days
	if now.Month() < birthdate.Month() || 
	   (now.Month() == birthdate.Month() && now.Day() < birthdate.Day()) {
		years--
	}
	
	return years
}
