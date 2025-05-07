package escalation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/incognito25/mamacare/services/go/internal/domain/model"
	"github.com/incognito25/mamacare/services/go/internal/domain/repository"
	"github.com/incognito25/mamacare/services/go/internal/errorx"
	"github.com/incognito25/mamacare/services/go/internal/logger"
)

// Service provides emergency escalation functionality
type Service struct {
	tierRepo        repository.EscalationTierRepository
	contactRepo     repository.ContactRepository
	pathRepo        repository.EscalationPathRepository
	sosRepo         repository.SOSRepository
	notifier        EscalationNotifier
	logger          logger.Logger
}

// EscalationNotifier defines the interface for sending escalation notifications
type EscalationNotifier interface {
	// SendEscalation sends an escalation notification
	SendEscalation(ctx context.Context, sosEvent *model.SOSEvent, tier *model.EscalationTier, pathName string) error
	
	// SendReminder sends a reminder for an unacknowledged escalation
	SendReminder(ctx context.Context, sosEvent *model.SOSEvent, tier *model.EscalationTier, pathName string, attempts int) error
}

// NewService creates a new escalation service
func NewService(
	tierRepo repository.EscalationTierRepository,
	contactRepo repository.ContactRepository,
	pathRepo repository.EscalationPathRepository,
	sosRepo repository.SOSRepository,
	notifier EscalationNotifier,
	logger logger.Logger,
) *Service {
	return &Service{
		tierRepo:    tierRepo,
		contactRepo: contactRepo,
		pathRepo:    pathRepo,
		sosRepo:     sosRepo,
		notifier:    notifier,
		logger:      logger,
	}
}

// CreateEscalationTier creates a new escalation tier
func (s *Service) CreateEscalationTier(
	ctx context.Context,
	name string,
	description string,
	level model.EscalationLevel,
	responseTimeMinutes int,
) (*model.EscalationTier, error) {
	// Validate inputs
	if name == "" {
		return nil, errorx.New(errorx.Validation, "Tier name cannot be empty")
	}
	
	if responseTimeMinutes <= 0 {
		return nil, errorx.New(errorx.Validation, "Response time must be positive")
	}
	
	// Create new tier
	tier := model.NewEscalationTier(uuid.New(), name, description, level, responseTimeMinutes)
	
	// Save to repository
	if err := s.tierRepo.Create(ctx, tier); err != nil {
		s.logger.Error(ctx, "Failed to create escalation tier", logger.FieldsMap{
			"error": err.Error(),
			"name":  name,
			"level": int(level),
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to create escalation tier", err)
	}
	
	s.logger.Info(ctx, "Created escalation tier", logger.FieldsMap{
		"tier_id": tier.ID.String(),
		"name":    name,
		"level":   int(level),
	})
	
	return tier, nil
}

// GetEscalationTierByID retrieves an escalation tier by its ID
func (s *Service) GetEscalationTierByID(ctx context.Context, tierID uuid.UUID) (*model.EscalationTier, error) {
	tier, err := s.tierRepo.GetByID(ctx, tierID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get escalation tier", logger.FieldsMap{
			"error":   err.Error(),
			"tier_id": tierID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "Escalation tier not found", err)
	}
	return tier, nil
}

// UpdateEscalationTier updates an existing escalation tier
func (s *Service) UpdateEscalationTier(
	ctx context.Context,
	tierID uuid.UUID,
	name string,
	description string,
	level model.EscalationLevel,
	responseTimeMinutes int,
) (*model.EscalationTier, error) {
	// Get existing tier
	tier, err := s.tierRepo.GetByID(ctx, tierID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get escalation tier for update", logger.FieldsMap{
			"error":   err.Error(),
			"tier_id": tierID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "Escalation tier not found", err)
	}
	
	// Update fields
	tier.Name = name
	tier.Description = description
	tier.Level = level
	tier.ResponseTime = responseTimeMinutes
	tier.UpdatedAt = time.Now()
	
	// Save to repository
	if err := s.tierRepo.Update(ctx, tier); err != nil {
		s.logger.Error(ctx, "Failed to update escalation tier", logger.FieldsMap{
			"error":   err.Error(),
			"tier_id": tierID.String(),
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to update escalation tier", err)
	}
	
	s.logger.Info(ctx, "Updated escalation tier", logger.FieldsMap{
		"tier_id": tierID.String(),
		"name":    name,
		"level":   int(level),
	})
	
	return tier, nil
}

// LinkTiers links two escalation tiers
func (s *Service) LinkTiers(ctx context.Context, tierID, nextTierID uuid.UUID) (*model.EscalationTier, error) {
	// Get existing tier
	tier, err := s.tierRepo.GetByID(ctx, tierID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get escalation tier for linking", logger.FieldsMap{
			"error":   err.Error(),
			"tier_id": tierID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "Escalation tier not found", err)
	}
	
	// Verify next tier exists
	_, err = s.tierRepo.GetByID(ctx, nextTierID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get next escalation tier for linking", logger.FieldsMap{
			"error":        err.Error(),
			"next_tier_id": nextTierID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "Next escalation tier not found", err)
	}
	
	// Link tiers
	tier.WithNextTier(nextTierID)
	
	// Save to repository
	if err := s.tierRepo.Update(ctx, tier); err != nil {
		s.logger.Error(ctx, "Failed to update escalation tier with link", logger.FieldsMap{
			"error":        err.Error(),
			"tier_id":      tierID.String(),
			"next_tier_id": nextTierID.String(),
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to link escalation tiers", err)
	}
	
	s.logger.Info(ctx, "Linked escalation tiers", logger.FieldsMap{
		"tier_id":      tierID.String(),
		"next_tier_id": nextTierID.String(),
	})
	
	return tier, nil
}

// AddContactToTier adds a contact to an escalation tier
func (s *Service) AddContactToTier(
	ctx context.Context,
	tierID uuid.UUID,
	contactID uuid.UUID,
) (*model.EscalationTier, error) {
	// Get existing tier
	tier, err := s.tierRepo.GetByID(ctx, tierID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get escalation tier for adding contact", logger.FieldsMap{
			"error":   err.Error(),
			"tier_id": tierID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "Escalation tier not found", err)
	}
	
	// Get contact
	contact, err := s.contactRepo.GetByID(ctx, contactID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get contact for escalation tier", logger.FieldsMap{
			"error":      err.Error(),
			"contact_id": contactID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "Contact not found", err)
	}
	
	// Add contact to tier
	tier.AddContact(*contact)
	
	// Save to repository
	if err := s.tierRepo.Update(ctx, tier); err != nil {
		s.logger.Error(ctx, "Failed to update escalation tier with contact", logger.FieldsMap{
			"error":      err.Error(),
			"tier_id":    tierID.String(),
			"contact_id": contactID.String(),
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to add contact to escalation tier", err)
	}
	
	s.logger.Info(ctx, "Added contact to escalation tier", logger.FieldsMap{
		"tier_id":    tierID.String(),
		"contact_id": contactID.String(),
		"contact_name": contact.Name,
	})
	
	return tier, nil
}

// CreateContact creates a new contact
func (s *Service) CreateContact(
	ctx context.Context,
	name string,
	role string,
	phone string,
	email string,
	isEmergency bool,
	isEscalation bool,
	facilityID *uuid.UUID,
) (*model.Contact, error) {
	// Validate inputs
	if name == "" {
		return nil, errorx.New(errorx.Validation, "Contact name cannot be empty")
	}
	
	if phone == "" {
		return nil, errorx.New(errorx.Validation, "Contact phone cannot be empty")
	}
	
	// Create new contact
	contact := model.NewContact(uuid.New(), name, role, phone)
	
	if email != "" {
		contact.WithEmail(email)
	}
	
	if facilityID != nil {
		contact.WithFacility(*facilityID)
	}
	
	if isEmergency {
		contact.MarkAsEmergency()
	}
	
	if isEscalation {
		contact.MarkAsEscalation()
	}
	
	// Save to repository
	if err := s.contactRepo.Create(ctx, contact); err != nil {
		s.logger.Error(ctx, "Failed to create contact", logger.FieldsMap{
			"error": err.Error(),
			"name":  name,
			"role":  role,
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to create contact", err)
	}
	
	s.logger.Info(ctx, "Created contact", logger.FieldsMap{
		"contact_id": contact.ID.String(),
		"name":       name,
		"role":       role,
	})
	
	return contact, nil
}

// CreateEscalationPath creates a new escalation path
func (s *Service) CreateEscalationPath(
	ctx context.Context,
	name string,
	description string,
	tierIDs []uuid.UUID,
	facilityID *uuid.UUID,
	districtID *uuid.UUID,
) (*model.EscalationPath, error) {
	// Validate inputs
	if name == "" {
		return nil, errorx.New(errorx.Validation, "Path name cannot be empty")
	}
	
	// Verify all tiers exist
	for _, tierID := range tierIDs {
		_, err := s.tierRepo.GetByID(ctx, tierID)
		if err != nil {
			s.logger.Error(ctx, "Failed to get escalation tier for path", logger.FieldsMap{
				"error":   err.Error(),
				"tier_id": tierID.String(),
			})
			return nil, errorx.NewWithCause(errorx.NotFound, fmt.Sprintf("Escalation tier %s not found", tierID), err)
		}
	}
	
	// Create new path
	path := model.NewEscalationPath(uuid.New(), name, description)
	
	// Add tiers
	for _, tierID := range tierIDs {
		path.AddTier(tierID)
	}
	
	// Add facility/district if provided
	if facilityID != nil {
		path.WithFacility(*facilityID)
	}
	
	if districtID != nil {
		path.WithDistrict(*districtID)
	}
	
	// Save to repository
	if err := s.pathRepo.Create(ctx, path); err != nil {
		s.logger.Error(ctx, "Failed to create escalation path", logger.FieldsMap{
			"error": err.Error(),
			"name":  name,
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to create escalation path", err)
	}
	
	s.logger.Info(ctx, "Created escalation path", logger.FieldsMap{
		"path_id":    path.ID.String(),
		"name":       name,
		"tier_count": len(tierIDs),
	})
	
	return path, nil
}

// GetEscalationPathByID retrieves an escalation path by its ID
func (s *Service) GetEscalationPathByID(ctx context.Context, pathID uuid.UUID) (*model.EscalationPath, error) {
	path, err := s.pathRepo.GetByID(ctx, pathID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get escalation path", logger.FieldsMap{
			"error":   err.Error(),
			"path_id": pathID.String(),
		})
		return nil, errorx.NewWithCause(errorx.NotFound, "Escalation path not found", err)
	}
	return path, nil
}

// GetEscalationPathsForFacility retrieves escalation paths for a facility
func (s *Service) GetEscalationPathsForFacility(ctx context.Context, facilityID uuid.UUID) ([]*model.EscalationPath, error) {
	paths, err := s.pathRepo.GetByFacility(ctx, facilityID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get escalation paths for facility", logger.FieldsMap{
			"error":       err.Error(),
			"facility_id": facilityID.String(),
		})
		return nil, errorx.NewWithCause(errorx.Internal, "Failed to get escalation paths for facility", err)
	}
	return paths, nil
}

// EscalateSOSEvent starts the escalation process for an SOS event
func (s *Service) EscalateSOSEvent(ctx context.Context, sosID uuid.UUID, pathID uuid.UUID) error {
	// Get SOS event
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for escalation", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}
	
	// Get escalation path
	path, err := s.pathRepo.GetByID(ctx, pathID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get escalation path", logger.FieldsMap{
			"error":   err.Error(),
			"path_id": pathID.String(),
		})
		return errorx.NewWithCause(errorx.NotFound, "Escalation path not found", err)
	}
	
	// Check path is active
	if !path.IsActive {
		return errorx.New(errorx.Validation, "Escalation path is not active")
	}
	
	// Check if path has tiers
	if len(path.TierIDs) == 0 {
		return errorx.New(errorx.Validation, "Escalation path has no tiers")
	}
	
	// Start with first tier
	firstTierID := path.TierIDs[0]
	tier, err := s.tierRepo.GetByID(ctx, firstTierID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get first tier for escalation", logger.FieldsMap{
			"error":   err.Error(),
			"tier_id": firstTierID.String(),
		})
		return errorx.NewWithCause(errorx.Internal, "Failed to get first escalation tier", err)
	}
	
	// Send escalation notification
	if err := s.notifier.SendEscalation(ctx, sosEvent, tier, path.Name); err != nil {
		s.logger.Error(ctx, "Failed to send escalation notification", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
			"path_id": pathID.String(),
			"tier_id": tier.ID.String(),
		})
		return errorx.NewWithCause(errorx.Internal, "Failed to send escalation notification", err)
	}
	
	s.logger.Info(ctx, "Started escalation process for SOS event", logger.FieldsMap{
		"sos_id": sosID.String(),
		"path_id": pathID.String(),
		"path_name": path.Name,
		"tier_id": tier.ID.String(),
		"tier_name": tier.Name,
	})
	
	return nil
}

// EscalateToNextTier escalates an SOS event to the next tier
func (s *Service) EscalateToNextTier(ctx context.Context, sosID uuid.UUID, currentTierID uuid.UUID) error {
	// Get current tier
	currentTier, err := s.tierRepo.GetByID(ctx, currentTierID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get current tier for escalation", logger.FieldsMap{
			"error":   err.Error(),
			"tier_id": currentTierID.String(),
		})
		return errorx.NewWithCause(errorx.NotFound, "Current escalation tier not found", err)
	}
	
	// Check if there is a next tier
	if currentTier.NextTierID == nil {
		return errorx.New(errorx.Validation, "No next tier available for escalation")
	}
	
	// Get next tier
	nextTier, err := s.tierRepo.GetByID(ctx, *currentTier.NextTierID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get next tier for escalation", logger.FieldsMap{
			"error":   err.Error(),
			"tier_id": currentTier.NextTierID.String(),
		})
		return errorx.NewWithCause(errorx.NotFound, "Next escalation tier not found", err)
	}
	
	// Get SOS event
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for escalation", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}
	
	// Send escalation notification for next tier
	// Using a placeholder path name since we don't know which path we're in
	pathName := "Emergency Escalation"
	if err := s.notifier.SendEscalation(ctx, sosEvent, nextTier, pathName); err != nil {
		s.logger.Error(ctx, "Failed to send escalation notification to next tier", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
			"current_tier_id": currentTierID.String(),
			"next_tier_id": nextTier.ID.String(),
		})
		return errorx.NewWithCause(errorx.Internal, "Failed to send escalation notification to next tier", err)
	}
	
	s.logger.Info(ctx, "Escalated SOS event to next tier", logger.FieldsMap{
		"sos_id": sosID.String(),
		"from_tier_id": currentTierID.String(),
		"from_tier_name": currentTier.Name,
		"to_tier_id": nextTier.ID.String(),
		"to_tier_name": nextTier.Name,
	})
	
	return nil
}

// SendEscalationReminder sends a reminder for an unacknowledged escalation
func (s *Service) SendEscalationReminder(ctx context.Context, sosID uuid.UUID, tierID uuid.UUID, attempts int) error {
	// Get SOS event
	sosEvent, err := s.sosRepo.GetByID(ctx, sosID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get SOS event for escalation reminder", logger.FieldsMap{
			"error":   err.Error(),
			"sos_id": sosID.String(),
		})
		return errorx.NewWithCause(errorx.NotFound, "SOS event not found", err)
	}
	
	// Get tier
	tier, err := s.tierRepo.GetByID(ctx, tierID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get tier for escalation reminder", logger.FieldsMap{
			"error":   err.Error(),
			"tier_id": tierID.String(),
		})
		return errorx.NewWithCause(errorx.NotFound, "Escalation tier not found", err)
	}
	
	// Using a placeholder path name since we don't know which path we're in
	pathName := "Emergency Escalation"
	
	// Send reminder
	if err := s.notifier.SendReminder(ctx, sosEvent, tier, pathName, attempts); err != nil {
		s.logger.Error(ctx, "Failed to send escalation reminder", logger.FieldsMap{
			"error":    err.Error(),
			"sos_id":  sosID.String(),
			"tier_id":  tierID.String(),
			"attempts": attempts,
		})
		return errorx.NewWithCause(errorx.Internal, "Failed to send escalation reminder", err)
	}
	
	s.logger.Info(ctx, "Sent escalation reminder", logger.FieldsMap{
		"sos_id":  sosID.String(),
		"tier_id":  tierID.String(),
		"tier_name": tier.Name,
		"attempts": attempts,
	})
	
	return nil
}
