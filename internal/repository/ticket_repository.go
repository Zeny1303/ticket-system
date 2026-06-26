package repository

import (
	"github.com/Zeny1303/ticket-system/internal/models"
	"gorm.io/gorm"
)

type TicketRepository interface {
	Create(ticket *models.Ticket) error
	FindAllByUserID(userID uint) ([]models.Ticket, error)
	FindByIDAndUserID(id, userID uint) (*models.Ticket, error)
	UpdateStatus(ticket *models.Ticket, newStatus models.TicketStatus) error
}

type ticketRepository struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) TicketRepository {
	return &ticketRepository{db: db}
}

func (r *ticketRepository) Create(ticket *models.Ticket) error {
	return r.db.Create(ticket).Error
}

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

func (r *ticketRepository) FindByIDAndUserID(id, userID uint) (*models.Ticket, error) {
	var ticket models.Ticket
	result := r.db.Where("id = ? AND user_id = ?", id, userID).First(&ticket)
	if result.Error != nil {
		return nil, result.Error
	}
	return &ticket, nil
}

func (r *ticketRepository) UpdateStatus(ticket *models.Ticket, newStatus models.TicketStatus) error {
	result := r.db.Model(ticket).Update("status", newStatus)
	if result.Error != nil {
		return result.Error
	}
	return r.db.First(ticket, ticket.ID).Error
}
