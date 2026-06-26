package services

import (
	"errors"

	"github.com/ticket-system/internal/models"
	"github.com/ticket-system/internal/repository"
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
// Business rules:
//   - Status always starts as "open" — the user cannot choose initial status
//   - UserID is set from the authenticated user's JWT, not from request body
//
// This enforces ownership at creation time.
func (s *ticketService) CreateTicket(req *models.CreateTicketRequest, userID uint) (*models.Ticket, error) {
	ticket := &models.Ticket{
		Title:       req.Title,
		Description: req.Description,
		Status:      models.StatusOpen, // Always starts open — business rule
		UserID:      userID,            // Ownership set from JWT, never from request body
	}

	if err := s.ticketRepo.Create(ticket); err != nil {
		return nil, errors.New("failed to create ticket")
	}

	return ticket, nil
}

// GetUserTickets returns all tickets for a given user.
// The repository's WHERE clause ensures only this user's tickets are returned.
// An empty list is a valid response — not an error.
func (s *ticketService) GetUserTickets(userID uint) ([]models.Ticket, error) {
	tickets, err := s.ticketRepo.FindAllByUserID(userID)
	if err != nil {
		return nil, errors.New("failed to retrieve tickets")
	}
	return tickets, nil
}

// GetTicketByID retrieves a single ticket, enforcing ownership.
// If the ticket doesn't exist OR belongs to a different user,
// we return a "not found" error — we don't distinguish between the two cases
// to avoid leaking information about other users' tickets.
func (s *ticketService) GetTicketByID(id, userID uint) (*models.Ticket, error) {
	ticket, err := s.ticketRepo.FindByIDAndUserID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ticket not found")
		}
		return nil, errors.New("failed to retrieve ticket")
	}
	return ticket, nil
}

// UpdateTicketStatus updates the status of a ticket with full validation.
// Business rules enforced here:
//   1. The ticket must exist and belong to the requesting user (ownership check)
//   2. The new status must be one of the three valid values
//   3. The transition must follow the allowed flow: open->in_progress->closed
//   4. A closed ticket cannot be reopened (enforced by IsValidTransition)
//
// This is the most business-logic-heavy function in the service layer.
// All four rules are checked before any database write.
func (s *ticketService) UpdateTicketStatus(id, userID uint, req *models.UpdateTicketStatusRequest) (*models.Ticket, error) {
	// Rule 1: Fetch the ticket and verify ownership.
	// FindByIDAndUserID returns not-found if the ticket doesn't belong to the user.
	ticket, err := s.ticketRepo.FindByIDAndUserID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ticket not found")
		}
		return nil, errors.New("failed to retrieve ticket")
	}

	// Rule 2: The requested new status must be a valid status value.
	// Reject anything that isn't "open", "in_progress", or "closed".
	if !models.IsValidStatus(req.Status) {
		return nil, errors.New("invalid status. Must be one of: open, in_progress, closed")
	}

	// Rule 3 & 4: Check if the transition is allowed.
	// ticket.IsValidTransition() encodes the state machine logic:
	//   open -> in_progress ✓
	//   in_progress -> closed ✓
	//   closed -> anything ✗
	//   open -> closed ✗ (must go through in_progress)
	if !ticket.IsValidTransition(req.Status) {
		return nil, errors.New(
			"invalid status transition. Allowed flow: open -> in_progress -> closed. Closed tickets cannot be reopened",
		)
	}

	// All rules passed — save the new status to the database.
	if err := s.ticketRepo.UpdateStatus(ticket, req.Status); err != nil {
		return nil, errors.New("failed to update ticket status")
	}

	return ticket, nil
}