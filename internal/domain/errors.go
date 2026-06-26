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

	// ErrRefreshTokenNotFound is returned by the repository when no row matches a
	// given token hash. The service translates it into ErrInvalidRefreshToken.
	ErrRefreshTokenNotFound = errors.New("refresh token not found")

	// ErrInvalidRefreshToken is the generic refresh failure surfaced to the
	// client (→ 401). It covers unknown, expired, and revoked/reused tokens
	// alike, so the response never reveals which case occurred.
	ErrInvalidRefreshToken = errors.New("invalid refresh token")

	// ErrInvalidGoogleToken is the generic Google sign-in failure surfaced to the
	// client (→ 401).
	ErrInvalidGoogleToken = errors.New("invalid google token")

	// ErrCategoryNotFound is returned when a category lookup/update/delete matches
	// no row. It covers BOTH "no such category" and "exists but not owned by this
	// user"
	ErrCategoryNotFound = errors.New("category not found")

	// ErrCategoryNameExists is returned when a user already has a category with
	// the same name (the UNIQUE (user_id, name) constraint, SQLSTATE 23505).
	ErrCategoryNameExists = errors.New("category name already exists")
)