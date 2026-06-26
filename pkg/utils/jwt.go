package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims defines the data we embed inside the JWT token.
// It embeds jwt.RegisteredClaims which includes standard fields like
// ExpiresAt, IssuedAt, and Issuer — the JWT spec standard fields.
// We add UserID so every protected handler knows which user is making the request.
//
// This is equivalent to the JWT payload in Node's jsonwebtoken:
// { userId: 1, exp: 1234567890, iat: 1234567890 }
type Claims struct {
	UserID uint `json:"user_id"`
	// Embedding RegisteredClaims gives us ExpiresAt, IssuedAt, Issuer for free
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT token for a given user ID.
// It is called after successful login or registration.
//
// Parameters:
//   userID     — the database ID of the authenticated user
//   secret     — the JWT signing secret from config
//   expiryHours — how many hours until the token expires
//
// Returns:
//   string — the signed JWT token string (what we send to the client)
//   error  — non-nil if signing fails
//
// The token is signed with HMAC-SHA256 (HS256) — a symmetric algorithm.
// Both signing and verification use the same secret key.
// This is appropriate for a single-service backend like ours.
func GenerateToken(userID uint, secret string, expiryHours int) (string, error) {
	// time.Now().Add() calculates the expiry time relative to now.
	// time.Hour * time.Duration(expiryHours) converts the int to a duration.
	expiryTime := time.Now().Add(time.Hour * time.Duration(expiryHours))

	// Build the claims struct that will be embedded in the token payload.
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			// ExpiresAt tells the JWT library when to reject this token.
			// jwt.NewNumericDate wraps a time.Time into the JWT numeric date format.
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ticket-system",
		},
	}

	// jwt.NewWithClaims creates an unsigned token with our claims.
	// jwt.SigningMethodHS256 specifies the HMAC-SHA256 algorithm.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// token.SignedString() signs the token with our secret and returns
	// the final JWT string in the format: header.payload.signature
	// []byte(secret) converts the string secret to a byte slice — required by the library.
	return token.SignedString([]byte(secret))
}

// ValidateToken parses and validates a JWT token string.
// It verifies the signature, checks expiry, and returns the embedded claims.
//
// Parameters:
//   tokenString — the raw JWT string from the Authorization header
//   secret      — the same secret used to sign the token
//
// Returns:
//   *Claims — the decoded payload if the token is valid
//   error   — non-nil if the token is invalid, expired, or malformed
//
// This is called by the JWT middleware on every protected request.
func ValidateToken(tokenString string, secret string) (*Claims, error) {
	// jwt.ParseWithClaims parses the token AND validates:
	//   1. The signature matches (prevents tampering)
	//   2. The token hasn't expired
	//   3. The algorithm matches what we expect
	//
	// The callback function (the third argument) is the "key function".
	// It receives the unverified token and must return the signing key.
	// We check the signing method here to prevent algorithm confusion attacks —
	// an attacker could send a token signed with "none" algorithm otherwise.
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// Verify the token uses HS256 — reject anything else.
			// This prevents algorithm substitution attacks.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	// Type assert the parsed claims to our custom Claims struct.
	// The "ok" pattern is Go's safe type assertion — if it fails, ok is false
	// instead of panicking (unlike a direct assertion which would panic).
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}