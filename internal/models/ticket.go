package models

import (
	"time"
)

// TicketStatus is a custom string type for ticket statuses.
// Using a named type (instead of plain string) means the compiler
// will catch mistakes like passing an arbitrary string where a
// TicketStatus is expected. This is Go's type safety at work.
type TicketStatus string

// These are the only valid ticket statuses in the system.
// Defining them as constants means they are:
//   - Compile-time checked (no typos)
//   - Autocomplete-friendly in any IDE
//   - Easy to find and change in one place
const (
	StatusOpen       TicketStatus = "open"
	StatusInProgress TicketStatus = "in_progress"
	StatusClosed     TicketStatus = "closed"
)

// Ticket represents the tickets table in PostgreSQL.
// It has a foreign key relationship to the users table via UserID.
type Ticket struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement"`

	// Title of the ticket — required, non-null
	Title string `json:"title" gorm:"not null"`

	// Description — optional, can be empty
	Description string `json:"description"`

	// Status uses our custom TicketStatus type.
	// default:open — GORM sets this as the DB column default.
	// This means even if no status is provided, PostgreSQL assigns "open".
	Status TicketStatus `json:"status" gorm:"type:varchar(20);default:open;not null"`

	// UserID is the foreign key — which user owns this ticket.
	// index — GORM creates an index on this column for fast lookups.
	// When we query "all tickets for user X", this index makes it fast.
	UserID uint `json:"user_id" gorm:"not null;index"`

	// User is a GORM association — it allows us to preload the user data
	// alongside the ticket when needed. The foreign key is UserID above.
	// json:"-" because we don't want to embed the full user object in ticket responses.
	User User `json:"-" gorm:"foreignKey:UserID"`

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
// The assignment specifies:
//   open -> in_progress -> closed
//   closed -> cannot move back
//
// This method lives on the model because it is a PURE BUSINESS RULE.
// It doesn't need a database or HTTP context — it's just logic about status values.
// The service layer calls this before saving any status update.
//
// In Go, methods are defined with a receiver — (t Ticket) means
// this method belongs to the Ticket type, like self in Python.
func (t Ticket) IsValidTransition(newStatus TicketStatus) bool {
	switch t.Status {
	case StatusOpen:
		// From open, you can only move to in_progress
		return newStatus == StatusInProgress
	case StatusInProgress:
		// From in_progress, you can only move to closed
		return newStatus == StatusClosed
	case StatusClosed:
		// Once closed, no transitions are allowed
		return false
	default:
		return false
	}
}

// IsValidStatus checks if a given string is one of the three valid statuses.
// Used for validating incoming PATCH requests before we even check transitions.
func IsValidStatus(s TicketStatus) bool {
	return s == StatusOpen || s == StatusInProgress || s == StatusClosed
}

// CreateTicketRequest is the DTO for POST /tickets.
// The user only provides title and description.
// Status defaults to "open" — the user cannot set the initial status.
// UserID is taken from the JWT token, not from the request body.
type CreateTicketRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=200"`
	Description string `json:"description" validate:"max=1000"`
}

// UpdateTicketStatusRequest is the DTO for PATCH /tickets/:id/status.
// Only the status field can be updated through this endpoint.
// This is intentional — title/description updates are out of scope.
type UpdateTicketStatusRequest struct {
	Status TicketStatus `json:"status" validate:"required"`
}