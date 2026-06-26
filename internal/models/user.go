package models

import (
	"time"
)

// User represents the users table in PostgreSQL.
// GORM reads the struct tags (gorm:"...") to understand column names,
// constraints, and relationships.
// The json tags control what gets serialized when we return JSON responses.
type User struct {
	// gorm:"primaryKey" — tells GORM this is the primary key column
	// autoIncrement is handled by PostgreSQL SERIAL type
	ID uint `json:"id" gorm:"primaryKey;autoIncrement"`

	// uniqueIndex — GORM creates a unique index on this column
	// not null — the column cannot be NULL in the database
	// json:"email" — this field appears as "email" in JSON output
	Email string `json:"email" gorm:"uniqueIndex;not null"`

	// json:"-" means this field is NEVER included in JSON output.
	// This is critical — we must never return password hashes in API responses.
	// The "-" is a Go convention for "exclude from marshaling".
	Password string `json:"-" gorm:"not null"`

	// autoCreateTime — GORM automatically sets this on INSERT
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	// autoUpdateTime — GORM automatically updates this on every UPDATE
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName tells GORM the exact PostgreSQL table name to use.
// Without this, GORM would default to "users" (pluralized snake_case of "User").
// Explicit is better than implicit in production code.
func (User) TableName() string {
	return "users"
}

// RegisterRequest is the shape of the JSON body for POST /auth/register.
// It is NOT the database model — it is a DTO (Data Transfer Object).
// We separate this from the User model because:
//   - The request has "password" (plain text) — the model has "password" (hashed)
//   - The request should never have id, created_at, updated_at
//   - Validation rules belong on the request, not the DB model
//
// validate tags are read by the go-playground/validator package.
// "required" — field must be present and non-zero
// "email" — must be a valid email format
// "min=6" — minimum 6 characters
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginRequest is the shape of the JSON body for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse is what we return after successful register or login.
// We return the token and basic user info.
// We never return the password — the User model's json:"-" ensures this.
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}