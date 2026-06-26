package services

import (
	"errors"

	"github.com/Zeny1303/ticket-system/internal/models"
	"github.com/Zeny1303/ticket-system/internal/repository"
	"github.com/Zeny1303/ticket-system/pkg/apperrors"
	"gorm.io/gorm"
)

type TicketService interface {
	CreateTicket(req *models.CreateTicketRequest, userID uint) (*models.Ticket, error)
	GetUserTickets(userID uint) ([]models.Ticket, error)
	GetTicketByID(id, userID uint) (*models.Ticket, error)
	UpdateTicketStatus(id, userID uint, req *models.UpdateTicketStatusRequest) (*models.Ticket, error)
}

type ticketService struct {
	ticketRepo repository.TicketRepository
}

func NewTicketService(ticketRepo repository.TicketRepository) TicketService {
	return &ticketService{ticketRepo: ticketRepo}
}

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

func (s *ticketService) GetUserTickets(userID uint) ([]models.Ticket, error) {
	tickets, err := s.ticketRepo.FindAllByUserID(userID)
	if err != nil {
		return nil, errors.New("failed to retrieve tickets")
	}
	return tickets, nil
}

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

func (s *ticketService) UpdateTicketStatus(id, userID uint, req *models.UpdateTicketStatusRequest) (*models.Ticket, error) {
	ticket, err := s.ticketRepo.FindByIDAndUserID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTicketNotFound
		}
		return nil, errors.New("failed to retrieve ticket")
	}

	if !models.IsValidStatus(req.Status) {
		return nil, apperrors.ErrInvalidStatus
	}

	if ticket.Status == req.Status {
		return nil, apperrors.ErrSameStatus
	}

	if !ticket.IsValidTransition(req.Status) {
		return nil, apperrors.ErrInvalidTransition
	}

	if err := s.ticketRepo.UpdateStatus(ticket, req.Status); err != nil {
		return nil, errors.New("failed to update ticket status")
	}

	return ticket, nil
}
