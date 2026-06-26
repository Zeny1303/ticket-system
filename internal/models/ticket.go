package models

import (
	"time"
)

// TicketStatus is a custom string type for ticket statuses.
// Using a named type means the compiler catches arbitrary string assignments.
type TicketStatus string

const (
	StatusOpen       TicketStatus = "open"
	StatusInProgress TicketStatus = "in_progress"
	StatusClosed     TicketStatus = "closed"
)

// Ticket represents the tickets table in PostgreSQL.
type Ticket struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement"`

	// Title — required, max 200 chars enforced at validation layer
	Title string `json:"title" gorm:"type:varchar(200);not null"`

	// Description — optional, stored as text (Issue #28: explicit type)
	Description string `json:"description" gorm:"type:text"`

	// Status uses our custom TicketStatus type.
	// default:open — GORM sets this as the DB column default.
	Status TicketStatus `json:"status" gorm:"type:varchar(20);default:open;not null"`

	// UserID is the foreign key — ownership of the ticket.
	// index — GORM creates a DB index for fast lookups by user.
	UserID uint `json:"user_id" gorm:"not null;index"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName explicitly sets the PostgreSQL table name.
func (Ticket) TableName() string {
	return "tickets"
}

// IsValidTransition checks if moving from the current status to a new status
// is allowed by the business rules.
//
// Allowed flow: open -> in_progress -> closed
// Closed tickets cannot be reopened.
func (t Ticket) IsValidTransition(newStatus TicketStatus) bool {
	switch t.Status {
	case StatusOpen:
		return newStatus == StatusInProgress
	case StatusInProgress:
		return newStatus == StatusClosed
	case StatusClosed:
		return false
	default:
		return false
	}
}

// IsValidStatus checks if a given TicketStatus is one of the three valid values.
func IsValidStatus(s TicketStatus) bool {
	return s == StatusOpen || s == StatusInProgress || s == StatusClosed
}

// CreateTicketRequest is the DTO for POST /tickets.
type CreateTicketRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=200"`
	Description string `json:"description" validate:"max=1000"`
}

// UpdateTicketStatusRequest is the DTO for PATCH /tickets/:id/status.
type UpdateTicketStatusRequest struct {
	Status TicketStatus `json:"status" validate:"required"`
}
