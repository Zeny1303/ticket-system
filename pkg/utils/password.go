package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword takes a plain-text password and returns a bcrypt hash.
// The hash includes a random salt automatically — bcrypt handles this internally.
// You never store the plain-text password. You store only this hash.
//
// bcrypt.DefaultCost is 10 — this controls how computationally expensive
// the hash is. Higher cost = harder to brute force, but slower to compute.
// Cost 10 is the industry standard for web applications.
//
// Django equivalent: make_password(password) — Django uses PBKDF2 by default
// but bcrypt is available as a hasher option.
// Express equivalent: bcrypt.hash(password, 10)
func HashPassword(password string) (string, error) {
	// bcrypt.GenerateFromPassword takes a byte slice and a cost factor.
	// It returns the hash as a byte slice which we convert to string for storage.
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compares a plain-text password against a bcrypt hash.
// Returns nil if the password matches, an error if it doesn't.
//
// IMPORTANT: You cannot "decrypt" bcrypt — it is a one-way hash.
// bcrypt.CompareHashAndPassword re-hashes the plain-text password
// using the same salt stored in the hash and compares the results.
//
// Django equivalent: check_password(password, hash)
// Express equivalent: bcrypt.compare(password, hash)
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}