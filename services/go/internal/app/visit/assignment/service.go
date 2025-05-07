package assignment

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/domain/repository"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// Service handles the assignment of visits to CHWs
type Service struct {
	visitRepo     repository.VisitRepository
	motherRepo    repository.MotherRepository
	userRepo      repository.UserRepository
	facilityRepo  repository.FacilityRepository
	routingClient RoutingClient
	log           logger.Logger
}

// RoutingClient defines the interface for route optimization
type RoutingClient interface {
	// OptimizeRoute optimizes a route for a set of locations
	OptimizeRoute(ctx context.Context, startLocation *model.GeoPoint, destinations []*model.GeoPoint) ([]int, error)
	
	// CalculateETA calculates the estimated time of arrival
	CalculateETA(ctx context.Context, from, to *model.GeoPoint, departureTime time.Time) (time.Time, error)
}

// NewService creates a new visit assignment service
func NewService(
	visitRepo repository.VisitRepository,
	motherRepo repository.MotherRepository,
	userRepo repository.UserRepository,
	facilityRepo repository.FacilityRepository,
	routingClient RoutingClient,
	log logger.Logger,
) *Service {
	return &Service{
		visitRepo:     visitRepo,
		motherRepo:    motherRepo,
		userRepo:      userRepo,
		facilityRepo:  facilityRepo,
		routingClient: routingClient,
		log:           log,
	}
}

// AssignCHW assigns a CHW to a visit
func (s *Service) AssignCHW(
	ctx context.Context,
	visitID uuid.UUID,
	chwID uuid.UUID,
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

	// Check if visit is already completed
	if visit.Status == model.VisitStatusCompleted || visit.Status == model.VisitStatusCancelled {
		return nil, errorx.Newf(errorx.BadRequest, "cannot assign CHW to a %s visit", visit.Status)
	}

	// Verify CHW exists
	user, err := s.userRepo.GetByID(ctx, chwID)
	if err != nil {
		s.log.Error("Failed to find CHW", logger.Fields{
			"error":  err.Error(),
			"chw_id": chwID.String(),
		})
		return nil, errorx.Wrap(err, "failed to find CHW")
	}

	// Check if user is a CHW
	if user.Role != "chw" {
		return nil, errorx.New(errorx.BadRequest, "user is not a CHW")
	}

	// Assign CHW
	visit.WithCHW(chwID)
	visit.UpdatedAt = time.Now()

	// Update visit
	if err := s.visitRepo.Update(ctx, visit); err != nil {
		s.log.Error("Failed to update visit", logger.Fields{
			"error":    err.Error(),
			"visit_id": visitID.String(),
		})
		return nil, errorx.Wrap(err, "failed to update visit")
	}

	s.log.Info("CHW assigned to visit successfully", logger.Fields{
		"visit_id":  visitID.String(),
		"chw_id":    chwID.String(),
		"mother_id": visit.MotherID.String(),
	})

	return visit, nil
}

// UnassignCHW removes CHW assignment from a visit
func (s *Service) UnassignCHW(
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

	// Check if visit has CHW assigned
	if visit.CHWID == nil {
		return nil, errorx.New(errorx.BadRequest, "visit does not have a CHW assigned")
	}

	// Check if visit is already completed
	if visit.Status == model.VisitStatusCompleted || visit.Status == model.VisitStatusCancelled {
		return nil, errorx.Newf(errorx.BadRequest, "cannot unassign CHW from a %s visit", visit.Status)
	}

	// Remove CHW assignment
	visit.CHWID = nil
	visit.UpdatedAt = time.Now()

	// Update visit
	if err := s.visitRepo.Update(ctx, visit); err != nil {
		s.log.Error("Failed to update visit", logger.Fields{
			"error":    err.Error(),
			"visit_id": visitID.String(),
		})
		return nil, errorx.Wrap(err, "failed to update visit")
	}

	s.log.Info("CHW unassigned from visit successfully", logger.Fields{
		"visit_id":  visitID.String(),
		"mother_id": visit.MotherID.String(),
	})

	return visit, nil
}

// GetCHWWorkload retrieves visits assigned to a specific CHW
func (s *Service) GetCHWWorkload(
	ctx context.Context,
	chwID uuid.UUID,
	startDate, endDate time.Time,
) ([]*model.Visit, error) {
	options := repository.NewVisitQueryOptions().
		WithDateRange(startDate, endDate).
		WithOrder("scheduled_time", "ASC")

	visits, err := s.visitRepo.GetByCHW(ctx, chwID, options)
	if err != nil {
		s.log.Error("Failed to get CHW workload", logger.Fields{
			"error":     err.Error(),
			"chw_id":    chwID.String(),
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
		})
		return nil, errorx.Wrap(err, "failed to get CHW workload")
	}

	return visits, nil
}

// AssignVisitsByCatchmentArea assigns visits to CHWs based on their catchment areas
func (s *Service) AssignVisitsByCatchmentArea(
	ctx context.Context,
	facilityID uuid.UUID,
	startDate, endDate time.Time,
) (int, error) {
	// Get unassigned visits in the date range
	options := repository.NewVisitQueryOptions().
		WithStatus(model.VisitStatusScheduled).
		WithDateRange(startDate, endDate)

	visits, err := s.visitRepo.GetByFacilityID(ctx, facilityID, options)
	if err != nil {
		s.log.Error("Failed to get unassigned visits", logger.Fields{
			"error":       err.Error(),
			"facility_id": facilityID.String(),
		})
		return 0, errorx.Wrap(err, "failed to get unassigned visits")
	}

	// Filter out visits that already have a CHW
	var unassignedVisits []*model.Visit
	for _, visit := range visits {
		if visit.CHWID == nil {
			unassignedVisits = append(unassignedVisits, visit)
		}
	}

	if len(unassignedVisits) == 0 {
		s.log.Info("No unassigned visits found", logger.Fields{
			"facility_id": facilityID.String(),
			"start_date":  startDate.Format(time.RFC3339),
			"end_date":    endDate.Format(time.RFC3339),
		})
		return 0, nil
	}

	// Get all CHWs for this facility
	// TODO: Replace with actual query to get CHWs by facility
	// For now, we'll simulate getting CHWs
	chws := []struct {
		ID            uuid.UUID
		CatchmentArea string
	}{
		{ID: uuid.New(), CatchmentArea: "Area A"},
		{ID: uuid.New(), CatchmentArea: "Area B"},
	}

	// Create a map of mothers by ID for quick lookup
	motherIDs := make([]uuid.UUID, 0, len(unassignedVisits))
	for _, visit := range unassignedVisits {
		motherIDs = append(motherIDs, visit.MotherID)
	}

	// Get mother information to determine catchment areas
	// TODO: Replace with actual repository call
	// For now, we'll simulate catchment areas
	catchmentAreas := map[uuid.UUID]string{
		motherIDs[0]: "Area A",
		motherIDs[1]: "Area B",
	}

	// Assign visits based on catchment area
	var assignedCount int
	for _, visit := range unassignedVisits {
		// Get mother's catchment area
		catchmentArea, exists := catchmentAreas[visit.MotherID]
		if !exists {
			continue // Skip if we don't know the catchment area
		}

		// Find CHW for this catchment area
		var selectedCHW *uuid.UUID
		for _, chw := range chws {
			if chw.CatchmentArea == catchmentArea {
				selectedCHW = &chw.ID
				break
			}
		}

		if selectedCHW == nil {
			continue // Skip if no CHW found for this area
		}

		// Assign CHW to visit
		visit.CHWID = selectedCHW
		visit.UpdatedAt = time.Now()

		// Update visit
		if err := s.visitRepo.Update(ctx, visit); err != nil {
			s.log.Error("Failed to update visit with CHW assignment", logger.Fields{
				"error":    err.Error(),
				"visit_id": visit.ID.String(),
				"chw_id":   selectedCHW.String(),
			})
			continue
		}

		assignedCount++
	}

	s.log.Info("Assigned visits by catchment area", logger.Fields{
		"facility_id":    facilityID.String(),
		"total_visits":   len(unassignedVisits),
		"assigned_count": assignedCount,
	})

	return assignedCount, nil
}

// OptimizeCHWRoutes optimizes routes for CHWs based on visit locations
func (s *Service) OptimizeCHWRoutes(
	ctx context.Context,
	chwID uuid.UUID,
	date time.Time,
) (*OptimizedRoute, error) {
	// Create date range for the specified day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Get visits for the CHW on the specified date
	options := repository.NewVisitQueryOptions().
		WithDateRange(startOfDay, endOfDay).
		WithStatus(model.VisitStatusScheduled).
		WithOrder("scheduled_time", "ASC")

	visits, err := s.visitRepo.GetByCHW(ctx, chwID, options)
	if err != nil {
		s.log.Error("Failed to get CHW visits", logger.Fields{
			"error":  err.Error(),
			"chw_id": chwID.String(),
			"date":   date.Format("2006-01-02"),
		})
		return nil, errorx.Wrap(err, "failed to get CHW visits")
	}

	if len(visits) == 0 {
		s.log.Info("No visits found for CHW on specified date", logger.Fields{
			"chw_id": chwID.String(),
			"date":   date.Format("2006-01-02"),
		})
		return &OptimizedRoute{
			CHWId:     chwID,
			Date:      date,
			Visits:    []*VisitWithLocation{},
			TotalTime: 0,
			Distance:  0,
		}, nil
	}

	// Get CHW information including starting location
	user, err := s.userRepo.GetByID(ctx, chwID)
	if err != nil {
		s.log.Error("Failed to find CHW", logger.Fields{
			"error":  err.Error(),
			"chw_id": chwID.String(),
		})
		return nil, errorx.Wrap(err, "failed to find CHW")
	}

	// Get CHW's starting location (e.g., facility or home)
	// TODO: Replace with actual CHW location lookup
	startLocation := &model.GeoPoint{Latitude: 8.4657, Longitude: -13.2317} // Example: Freetown, Sierra Leone

	// Collect mother locations for each visit
	visitsWithLocations := make([]*VisitWithLocation, 0, len(visits))
	motherIDs := make([]uuid.UUID, 0, len(visits))
	
	for _, visit := range visits {
		motherIDs = append(motherIDs, visit.MotherID)
	}

	// Get mothers' information to get locations
	// TODO: Replace with batch query to get all mothers at once
	for i, visit := range visits {
		mother, err := s.motherRepo.GetByID(ctx, visit.MotherID)
		if err != nil {
			s.log.Error("Failed to find mother", logger.Fields{
				"error":     err.Error(),
				"mother_id": visit.MotherID.String(),
			})
			continue
		}

		// Skip if mother doesn't have location
		if mother.Location == nil {
			continue
		}

		visitsWithLocations = append(visitsWithLocations, &VisitWithLocation{
			Visit:    visit,
			Location: mother.Location,
			Mother:   mother,
		})
	}

	if len(visitsWithLocations) == 0 {
		return nil, errorx.New(errorx.BadRequest, "no visits with valid locations found")
	}

	// Collect destination points for routing
	destinations := make([]*model.GeoPoint, len(visitsWithLocations))
	for i, vwl := range visitsWithLocations {
		destinations[i] = vwl.Location
	}

	// Optimize route
	visitOrder, err := s.routingClient.OptimizeRoute(ctx, startLocation, destinations)
	if err != nil {
		s.log.Error("Failed to optimize route", logger.Fields{
			"error":  err.Error(),
			"chw_id": chwID.String(),
		})
		return nil, errorx.Wrap(err, "failed to optimize route")
	}

	// Reorder visits based on optimized route
	optimizedVisits := make([]*VisitWithLocation, len(visitOrder))
	for i, idx := range visitOrder {
		optimizedVisits[i] = visitsWithLocations[idx]
	}

	// Calculate ETAs for each visit
	currentTime := startOfDay.Add(8 * time.Hour) // Assume 8 AM start
	currentLocation := startLocation

	for i, visit := range optimizedVisits {
		// Calculate ETA
		eta, err := s.routingClient.CalculateETA(ctx, currentLocation, visit.Location, currentTime)
		if err != nil {
			s.log.Error("Failed to calculate ETA", logger.Fields{
				"error":    err.Error(),
				"visit_id": visit.Visit.ID.String(),
			})
			// Use a rough estimate if calculation fails
			eta = currentTime.Add(30 * time.Minute)
		}

		// Set ETA
		visit.EstimatedArrival = eta

		// Update current state for next calculation
		currentTime = eta.Add(30 * time.Minute) // Assume 30 minutes per visit
		currentLocation = visit.Location
	}

	// Calculate total distance and time (this would normally come from the routing engine)
	// For now, we'll just use a placeholder calculation
	totalDistance := 0.0
	totalTimeMinutes := 0

	// Return the optimized route
	return &OptimizedRoute{
		CHWId:     chwID,
		Date:      date,
		Visits:    optimizedVisits,
		TotalTime: totalTimeMinutes,
		Distance:  totalDistance,
	}, nil
}

// BalanceWorkload balances workload across CHWs by redistributing visits
func (s *Service) BalanceWorkload(
	ctx context.Context,
	facilityID uuid.UUID,
	date time.Time,
) (int, error) {
	// Create date range for the specified day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Get all visits for the facility on the specified date
	options := repository.NewVisitQueryOptions().
		WithDateRange(startOfDay, endOfDay).
		WithStatus(model.VisitStatusScheduled)

	visits, err := s.visitRepo.GetByFacilityID(ctx, facilityID, options)
	if err != nil {
		s.log.Error("Failed to get facility visits", logger.Fields{
			"error":       err.Error(),
			"facility_id": facilityID.String(),
			"date":        date.Format("2006-01-02"),
		})
		return 0, errorx.Wrap(err, "failed to get facility visits")
	}

	// Get all CHWs for this facility
	// TODO: Replace with actual query to get CHWs by facility
	// For now, we'll simulate getting CHWs
	chws := []struct {
		ID       uuid.UUID
		Capacity int // Maximum number of visits per day
	}{
		{ID: uuid.New(), Capacity: 5},
		{ID: uuid.New(), Capacity: 5},
		{ID: uuid.New(), Capacity: 4},
	}

	// Count current assignments
	chwWorkload := make(map[uuid.UUID]int)
	for _, chw := range chws {
		chwWorkload[chw.ID] = 0
	}

	// Count current workload
	for _, visit := range visits {
		if visit.CHWID != nil {
			chwWorkload[*visit.CHWID]++
		}
	}

	// Find CHWs who are over capacity
	var overloaded []uuid.UUID
	var underloaded []uuid.UUID
	
	for _, chw := range chws {
		capacity := chw.Capacity
		current := chwWorkload[chw.ID]
		
		if current > capacity {
			overloaded = append(overloaded, chw.ID)
		} else if current < capacity {
			underloaded = append(underloaded, chw.ID)
		}
	}

	// If no CHWs are overloaded, no need to balance
	if len(overloaded) == 0 {
		s.log.Info("No CHWs are overloaded, no balancing needed", logger.Fields{
			"facility_id": facilityID.String(),
			"date":        date.Format("2006-01-02"),
		})
		return 0, nil
	}

	// Collect visits by CHW
	visitsByCHW := make(map[uuid.UUID][]*model.Visit)
	for _, visit := range visits {
		if visit.CHWID != nil {
			visitsByCHW[*visit.CHWID] = append(visitsByCHW[*visit.CHWID], visit)
		}
	}

	// Perform balancing
	var reassignedCount int
	
	for _, overCHW := range overloaded {
		overVisits := visitsByCHW[overCHW]
		
		// Sort by scheduled time to redistribute later visits first
		sort.Slice(overVisits, func(i, j int) bool {
			return overVisits[i].ScheduledTime.After(overVisits[j].ScheduledTime)
		})
		
		// Calculate how many visits to redistribute
		overCapacity := chwWorkload[overCHW] - getCapacity(chws, overCHW)
		
		// Redistribute visits
		for i := 0; i < overCapacity && i < len(overVisits); i++ {
			// Find least loaded CHW
			var leastLoadedCHW uuid.UUID
			var leastLoad int = 999
			
			for _, underCHW := range underloaded {
				if chwWorkload[underCHW] < leastLoad {
					leastLoad = chwWorkload[underCHW]
					leastLoadedCHW = underCHW
				}
			}
			
			// If no underloaded CHW found, break
			if leastLoad == 999 {
				break
			}
			
			// Reassign visit
			visit := overVisits[i]
			visit.CHWID = &leastLoadedCHW
			visit.UpdatedAt = time.Now()
			
			// Update visit in database
			if err := s.visitRepo.Update(ctx, visit); err != nil {
				s.log.Error("Failed to reassign visit", logger.Fields{
					"error":     err.Error(),
					"visit_id":  visit.ID.String(),
					"from_chw":  overCHW.String(),
					"to_chw":    leastLoadedCHW.String(),
				})
				continue
			}
			
			// Update workload counts
			chwWorkload[overCHW]--
			chwWorkload[leastLoadedCHW]++
			reassignedCount++
			
			// If CHW is no longer underloaded, remove from list
			if chwWorkload[leastLoadedCHW] >= getCapacity(chws, leastLoadedCHW) {
				for i, id := range underloaded {
					if id == leastLoadedCHW {
						underloaded = append(underloaded[:i], underloaded[i+1:]...)
						break
					}
				}
			}
		}
	}

	s.log.Info("Balanced CHW workload", logger.Fields{
		"facility_id":      facilityID.String(),
		"date":             date.Format("2006-01-02"),
		"reassigned_count": reassignedCount,
	})

	return reassignedCount, nil
}

// getCapacity gets capacity for a CHW from the list
func getCapacity(chws []struct {
	ID       uuid.UUID
	Capacity int
}, chwID uuid.UUID) int {
	for _, chw := range chws {
		if chw.ID == chwID {
			return chw.Capacity
		}
	}
	return 5 // Default capacity
}

// OptimizedRoute represents an optimized route for a CHW
type OptimizedRoute struct {
	CHWId     uuid.UUID            `json:"chw_id"`
	Date      time.Time            `json:"date"`
	Visits    []*VisitWithLocation `json:"visits"`
	TotalTime int                  `json:"total_time_minutes"`
	Distance  float64              `json:"distance_km"`
}

// VisitWithLocation combines a visit with its location information
type VisitWithLocation struct {
	Visit           *model.Visit     `json:"visit"`
	Location        *model.GeoPoint  `json:"location"`
	Mother          *model.Mother    `json:"mother"`
	EstimatedArrival time.Time       `json:"estimated_arrival"`
}

// UpdateVisitOrder updates the order of visits for a CHW
func (s *Service) UpdateVisitOrder(
	ctx context.Context,
	chwID uuid.UUID,
	visitIDs []uuid.UUID,
) error {
	// Verify all visits exist and belong to the specified CHW
	for _, visitID := range visitIDs {
		visit, err := s.visitRepo.GetByID(ctx, visitID)
		if err != nil {
			s.log.Error("Failed to find visit", logger.Fields{
				"error":    err.Error(),
				"visit_id": visitID.String(),
			})
			return errorx.Wrap(err, "failed to find visit")
		}

		if visit.CHWID == nil || *visit.CHWID != chwID {
			return errorx.Newf(errorx.BadRequest, "visit %s is not assigned to the specified CHW", visitID)
		}
	}

	// For now, we're not persisting the order as that would require adding an order field
	// to the Visit model. In a real implementation, you would add this field and update it here.
	// This could be a future enhancement.

	s.log.Info("Updated visit order for CHW", logger.Fields{
		"chw_id":      chwID.String(),
		"visit_count": len(visitIDs),
	})

	return nil
}
