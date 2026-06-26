package models

import "time"

type TicketStatus string

const (
	StatusOpen       TicketStatus = "open"
	StatusInProgress TicketStatus = "in_progress"
	StatusClosed     TicketStatus = "closed"
)

type Ticket struct {
	ID          uint         `json:"id" gorm:"primaryKey;autoIncrement"`
	Title       string       `json:"title" gorm:"type:varchar(200);not null"`
	Description string       `json:"description" gorm:"type:text"`
	Status      TicketStatus `json:"status" gorm:"type:varchar(20);default:open;not null"`
	UserID      uint         `json:"user_id" gorm:"not null;index"`
	CreatedAt   time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Ticket) TableName() string {
	return "tickets"
}

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

func IsValidStatus(s TicketStatus) bool {
	return s == StatusOpen || s == StatusInProgress || s == StatusClosed
}

type CreateTicketRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=200"`
	Description string `json:"description" validate:"max=1000"`
}

type UpdateTicketStatusRequest struct {
	Status TicketStatus `json:"status" validate:"required"`
}
