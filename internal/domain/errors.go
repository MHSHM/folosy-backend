package domain

import "errors"

// Sentinel errors represent domain-level outcomes that callers may want to
// branch on (for example, the transport layer mapping them to specific HTTP
var (
	// ErrEmailExists is returned when a registration is attempted with an email that is already taken.
	ErrEmailExists = errors.New("email already exists")
)