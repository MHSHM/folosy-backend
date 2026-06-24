package domain

import "errors"

// Sentinel errors represent domain-level outcomes that callers may want to
// branch on (for example, the transport layer mapping them to specific HTTP
var (
	// ErrEmailExists is returned when a registration is attempted with an email that is already taken.
	ErrEmailExists = errors.New("email already exists")

	// ErrUserNotFound is returned when a lookup finds no user matching the criteria.
	ErrUserNotFound = errors.New("user not found")

	// ErrInvalidCredentials is returned on a failed login. It is deliberately
	// generic — used for both "no such email" and "wrong password"
	ErrInvalidCredentials = errors.New("invalid email or password")
)