package handlers

import (
	"net/http"
	"strconv"

	"github.com/ticket-system/internal/middleware"
	"github.com/ticket-system/internal/models"
	"github.com/ticket-system/internal/services"
	"github.com/ticket-system/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// TicketHandler holds dependencies for ticket HTTP handlers.
type TicketHandler struct {
	ticketService services.TicketService
	validate      *validator.Validate
}

// NewTicketHandler constructs a TicketHandler.
func NewTicketHandler(ticketService services.TicketService) *TicketHandler {
	return &TicketHandler{
		ticketService: ticketService,
		validate:      validator.New(),
	}
}

// CreateTicket handles POST /tickets
//
// This is a protected route — AuthMiddleware runs first and injects userID.
// The handler:
//   1. Gets the authenticated user's ID from context
//   2. Parses and validates the request body
//   3. Calls the service to create the ticket
//   4. Returns 201 Created with the new ticket
func (h *TicketHandler) CreateTicket(c *gin.Context) {
	// Get the authenticated user's ID that AuthMiddleware stored in context.
	// This is always safe on protected routes — middleware has already validated the token.
	userID := middleware.GetUserID(c)

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

// GetUserTickets handles GET /tickets
//
// Returns all tickets belonging to the authenticated user.
// Returns an empty array (not null) if the user has no tickets.
// This is important — frontend clients should not need to handle null vs empty array.
func (h *TicketHandler) GetUserTickets(c *gin.Context) {
	userID := middleware.GetUserID(c)

	tickets, err := h.ticketService.GetUserTickets(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// If tickets is nil (which shouldn't happen with our repo implementation),
	// we ensure we return an empty slice, not null in the JSON response.
	// JSON null vs [] is a common source of frontend bugs.
	if tickets == nil {
		tickets = []models.Ticket{}
	}

	utils.Success(c, http.StatusOK, "Tickets retrieved successfully", tickets)
}

// GetTicketByID handles GET /tickets/:id
//
// Returns a single ticket by its ID.
// The ownership check is done inside the service/repository layer —
// the handler just calls the service and maps errors to HTTP status codes.
func (h *TicketHandler) GetTicketByID(c *gin.Context) {
	userID := middleware.GetUserID(c)

	// c.Param("id") extracts the :id path parameter from the URL.
	// It returns a string — we must parse it to uint.
	// In Django: kwargs['pk']. In Express: req.params.id
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid ticket ID: must be a positive integer")
		return
	}

	ticket, err := h.ticketService.GetTicketByID(uint(ticketID), userID)
	if err != nil {
		if err.Error() == "ticket not found" {
			// 404 Not Found — could mean doesn't exist OR belongs to another user.
			// We intentionally don't distinguish these cases.
			utils.Error(c, http.StatusNotFound, "Ticket not found")
			return
		}
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Ticket retrieved successfully", ticket)
}

// UpdateTicketStatus handles PATCH /tickets/:id/status
//
// Updates only the status field of a ticket.
// Full business rule validation happens in the service layer.
// This handler's job: parse URL params, parse body, call service, map errors.
func (h *TicketHandler) UpdateTicketStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)

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
		// Map different business errors to appropriate HTTP status codes.
		switch err.Error() {
		case "ticket not found":
			utils.Error(c, http.StatusNotFound, "Ticket not found")
		case "invalid status. Must be one of: open, in_progress, closed":
			utils.Error(c, http.StatusBadRequest, err.Error())
		default:
			// Any transition error or unexpected error
			if len(err.Error()) > 0 {
				utils.Error(c, http.StatusUnprocessableEntity, err.Error())
			} else {
				utils.Error(c, http.StatusInternalServerError, "Failed to update ticket status")
			}
		}
		return
	}

	utils.Success(c, http.StatusOK, "Ticket status updated successfully", ticket)
}