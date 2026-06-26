package repository

import (
	"github.com/ticket-system/internal/models"
	"gorm.io/gorm"
)

// TicketRepository defines the interface for ticket database operations.
// The service layer depends on this interface, not the concrete struct.
type TicketRepository interface {
	Create(ticket *models.Ticket) error
	FindAllByUserID(userID uint) ([]models.Ticket, error)
	FindByIDAndUserID(id, userID uint) (*models.Ticket, error)
	UpdateStatus(ticket *models.Ticket, newStatus models.TicketStatus) error
}

type ticketRepository struct {
	db *gorm.DB
}

// NewTicketRepository is the constructor for ticketRepository.
func NewTicketRepository(db *gorm.DB) TicketRepository {
	return &ticketRepository{db: db}
}

// Create inserts a new ticket into the database.
// The ticket.UserID must be set before calling this — the repository
// does not set ownership. The service layer handles that.
//
// GORM executes: INSERT INTO tickets (title, description, status, user_id, ...) VALUES (...)
func (r *ticketRepository) Create(ticket *models.Ticket) error {
	result := r.db.Create(ticket)
	return result.Error
}

// FindAllByUserID retrieves all tickets belonging to a specific user.
// The WHERE clause on user_id enforces ownership — a user can never
// accidentally see another user's tickets through this query.
//
// GORM executes: SELECT * FROM tickets WHERE user_id = ? ORDER BY created_at DESC
//
// We order by created_at DESC so the most recent tickets appear first.
// This is a sensible default for any list endpoint.
func (r *ticketRepository) FindAllByUserID(userID uint) ([]models.Ticket, error) {
	var tickets []models.Ticket
	result := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&tickets)
	if result.Error != nil {
		return nil, result.Error
	}
	// If no tickets exist, Find() returns an empty slice — not an error.
	// This is correct: an empty list is a valid successful response.
	return tickets, nil
}

// FindByIDAndUserID retrieves a single ticket by ID, scoped to a specific user.
// The DOUBLE condition (id AND user_id) is the ownership check.
//
// WHY two conditions?
// If we only queried by id, user A could request /tickets/5 and see user B's ticket.
// By adding user_id to the WHERE clause, we get gorm.ErrRecordNotFound if the ticket
// doesn't belong to the requesting user — as if the ticket doesn't exist.
// This is intentional: we don't reveal whether a ticket exists to non-owners.
//
// GORM executes: SELECT * FROM tickets WHERE id = ? AND user_id = ? LIMIT 1
func (r *ticketRepository) FindByIDAndUserID(id, userID uint) (*models.Ticket, error) {
	var ticket models.Ticket
	result := r.db.Where("id = ? AND user_id = ?", id, userID).First(&ticket)
	if result.Error != nil {
		return nil, result.Error
	}
	return &ticket, nil
}

// UpdateStatus saves the new status to an existing ticket.
// We use db.Save() which executes a full UPDATE on all fields.
// The ticket object passed in should already have its Status updated
// by the service layer before calling this function.
//
// GORM executes: UPDATE tickets SET status = ?, updated_at = ? WHERE id = ?
func (r *ticketRepository) UpdateStatus(ticket *models.Ticket, newStatus models.TicketStatus) error {
	// db.Model() specifies which record to update.
	// db.Update() updates only the specified column — safer than db.Save()
	// which would update all fields and could overwrite concurrent changes.
	result := r.db.Model(ticket).Update("status", newStatus)
	if result.Error != nil {
		return result.Error
	}
	// Update the in-memory struct to reflect the new status.
	// This way, the service layer can return the updated ticket to the handler.
	ticket.Status = newStatus
	return nil
}