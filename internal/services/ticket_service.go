package services

import (
	"errors"

	"github.com/Zeny1303/ticket-system/internal/models"
	"github.com/Zeny1303/ticket-system/internal/repository"
	"github.com/Zeny1303/ticket-system/pkg/apperrors"
	"gorm.io/gorm"
)

// TicketService defines the interface for ticket business logic.
type TicketService interface {
	CreateTicket(req *models.CreateTicketRequest, userID uint) (*models.Ticket, error)
	GetUserTickets(userID uint) ([]models.Ticket, error)
	GetTicketByID(id, userID uint) (*models.Ticket, error)
	UpdateTicketStatus(id, userID uint, req *models.UpdateTicketStatusRequest) (*models.Ticket, error)
}

type ticketService struct {
	ticketRepo repository.TicketRepository
}

// NewTicketService constructs a ticketService with its repository dependency.
func NewTicketService(ticketRepo repository.TicketRepository) TicketService {
	return &ticketService{ticketRepo: ticketRepo}
}

// CreateTicket creates a new ticket owned by the given user.
// Status always starts as "open" — the user cannot choose the initial status.
// UserID is set from the authenticated JWT, never from the request body.
func (s *ticketService) CreateTicket(req *models.CreateTicketRequest, userID uint) (*models.Ticket, error) {
	ticket := &models.Ticket{
		Title:       req.Title,
		Description: req.Description,
		Status:      models.StatusOpen,
		UserID:      userID,
	}

	if err := s.ticketRepo.Create(ticket); err != nil {
		return nil, errors.New("failed to create ticket")
	}

	return ticket, nil
}

// GetUserTickets returns all tickets for a given user.
// An empty list is a valid response — not an error.
func (s *ticketService) GetUserTickets(userID uint) ([]models.Ticket, error) {
	tickets, err := s.ticketRepo.FindAllByUserID(userID)
	if err != nil {
		return nil, errors.New("failed to retrieve tickets")
	}
	return tickets, nil
}

// GetTicketByID retrieves a single ticket, enforcing ownership.
// Returns ErrTicketNotFound if the ticket doesn't exist OR belongs to another user.
// We do not distinguish between the two cases to avoid leaking information.
func (s *ticketService) GetTicketByID(id, userID uint) (*models.Ticket, error) {
	ticket, err := s.ticketRepo.FindByIDAndUserID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTicketNotFound
		}
		return nil, errors.New("failed to retrieve ticket")
	}
	return ticket, nil
}

// UpdateTicketStatus updates the status of a ticket with full validation.
// Business rules enforced:
//  1. Ticket must exist and belong to the requesting user (ownership)
//  2. New status must be one of: open, in_progress, closed
//  3. Issue #7 fix: same-status no-op returns a clear error
//  4. Transition must follow: open -> in_progress -> closed
//  5. Closed tickets cannot be reopened
func (s *ticketService) UpdateTicketStatus(id, userID uint, req *models.UpdateTicketStatusRequest) (*models.Ticket, error) {
	// Rule 1: fetch and verify ownership.
	ticket, err := s.ticketRepo.FindByIDAndUserID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTicketNotFound
		}
		return nil, errors.New("failed to retrieve ticket")
	}

	// Rule 2: requested status must be a valid value.
	if !models.IsValidStatus(req.Status) {
		return nil, apperrors.ErrInvalidStatus
	}

	// Rule 3 (Issue #7 fix): reject same-status no-op with a clear message.
	if ticket.Status == req.Status {
		return nil, apperrors.ErrSameStatus
	}

	// Rule 4 & 5: check allowed transition.
	if !ticket.IsValidTransition(req.Status) {
		return nil, apperrors.ErrInvalidTransition
	}

	// All rules passed — persist the new status.
	if err := s.ticketRepo.UpdateStatus(ticket, req.Status); err != nil {
		return nil, errors.New("failed to update ticket status")
	}

	return ticket, nil
}
