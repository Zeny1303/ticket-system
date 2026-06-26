package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Zeny1303/ticket-system/internal/middleware"
	"github.com/Zeny1303/ticket-system/internal/models"
	"github.com/Zeny1303/ticket-system/internal/services"
	"github.com/Zeny1303/ticket-system/pkg/apperrors"
	"github.com/Zeny1303/ticket-system/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type TicketHandler struct {
	ticketService services.TicketService
	validate      *validator.Validate
}

func NewTicketHandler(ticketService services.TicketService) *TicketHandler {
	return &TicketHandler{
		ticketService: ticketService,
		validate:      validator.New(),
	}
}

func (h *TicketHandler) CreateTicket(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if err := h.validate.Struct(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	ticket, err := h.ticketService.CreateTicket(&req, userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, http.StatusCreated, "Ticket created successfully", ticket)
}

func (h *TicketHandler) GetUserTickets(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tickets, err := h.ticketService.GetUserTickets(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	if tickets == nil {
		tickets = []models.Ticket{}
	}

	utils.Success(c, http.StatusOK, "Tickets retrieved successfully", tickets)
}

func (h *TicketHandler) GetTicketByID(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid ticket ID: must be a positive integer")
		return
	}

	ticket, err := h.ticketService.GetTicketByID(uint(ticketID), userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrTicketNotFound) {
			utils.Error(c, http.StatusNotFound, "Ticket not found")
			return
		}
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Ticket retrieved successfully", ticket)
}

func (h *TicketHandler) UpdateTicketStatus(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid ticket ID: must be a positive integer")
		return
	}

	var req models.UpdateTicketStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if err := h.validate.Struct(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	ticket, err := h.ticketService.UpdateTicketStatus(uint(ticketID), userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrTicketNotFound):
			utils.Error(c, http.StatusNotFound, err.Error())
		case errors.Is(err, apperrors.ErrInvalidStatus):
			utils.Error(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, apperrors.ErrSameStatus):
			utils.Error(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, apperrors.ErrInvalidTransition):
			utils.Error(c, http.StatusUnprocessableEntity, err.Error())
		default:
			utils.Error(c, http.StatusInternalServerError, "Failed to update ticket status")
		}
		return
	}

	utils.Success(c, http.StatusOK, "Ticket status updated successfully", ticket)
}
