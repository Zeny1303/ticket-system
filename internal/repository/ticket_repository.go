package repository

import (
	"github.com/Zeny1303/ticket-system/internal/models"
	"gorm.io/gorm"
)

// TicketRepository defines the interface for ticket database operations.
type TicketRepository interface {
	Create(ticket *models.Ticket) error
	FindAllByUserID(userID uint) ([]models.Ticket, error)
	FindByIDAndUserID(id, userID uint) (*models.Ticket, error)
	UpdateStatus(ticket *models.Ticket, newStatus models.TicketStatus) error
}

type ticketRepository struct {
	db *gorm.DB
}

// NewTicketRepository constructs a ticketRepository.
func NewTicketRepository(db *gorm.DB) TicketRepository {
	return &ticketRepository{db: db}
}

// Create inserts a new ticket into the database.
// ticket.UserID must be set before calling — the repository does not set ownership.
func (r *ticketRepository) Create(ticket *models.Ticket) error {
	return r.db.Create(ticket).Error
}

// FindAllByUserID retrieves all tickets belonging to a specific user,
// ordered by most recently created first.
// GORM: SELECT * FROM tickets WHERE user_id = ? ORDER BY created_at DESC
func (r *ticketRepository) FindAllByUserID(userID uint) ([]models.Ticket, error) {
	var tickets []models.Ticket
	result := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&tickets)
	if result.Error != nil {
		return nil, result.Error
	}
	return tickets, nil
}

// FindByIDAndUserID retrieves a single ticket by ID scoped to a specific user.
// The dual WHERE condition (id AND user_id) is the ownership check.
// Returns gorm.ErrRecordNotFound if the ticket doesn't exist or belongs to another user —
// we intentionally do not distinguish between the two cases.
// GORM: SELECT * FROM tickets WHERE id = ? AND user_id = ? LIMIT 1
func (r *ticketRepository) FindByIDAndUserID(id, userID uint) (*models.Ticket, error) {
	var ticket models.Ticket
	result := r.db.Where("id = ? AND user_id = ?", id, userID).First(&ticket)
	if result.Error != nil {
		return nil, result.Error
	}
	return &ticket, nil
}

// UpdateStatus saves the new status and reloads the full record from the DB.
// Issue #8 fix: after the UPDATE, we reload the record so that the in-memory
// struct reflects the DB-generated updated_at timestamp, not the stale value.
// GORM: UPDATE tickets SET status = ?, updated_at = ? WHERE id = ?
func (r *ticketRepository) UpdateStatus(ticket *models.Ticket, newStatus models.TicketStatus) error {
	result := r.db.Model(ticket).Update("status", newStatus)
	if result.Error != nil {
		return result.Error
	}
	// Reload the full record so the caller gets the correct updated_at from DB.
	return r.db.First(ticket, ticket.ID).Error
}
