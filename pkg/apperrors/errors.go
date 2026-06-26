package apperrors

import "errors"

var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTicketNotFound     = errors.New("ticket not found")
	ErrInvalidStatus      = errors.New("invalid status. Must be one of: open, in_progress, closed")
	ErrInvalidTransition  = errors.New("invalid status transition. Allowed flow: open -> in_progress -> closed. Closed tickets cannot be reopened")
	ErrSameStatus         = errors.New("ticket is already in this status")
	ErrDatabase           = errors.New("database error")
)
