package domain

import (
	"database/sql"
	"time"
)

// RefreshToken is a persisted refresh-token row (what we read back from the DB),
//
// RevokedAt is sql.NullTime (a value type: {Time time.Time; Valid bool}) rather
// than *time.Time: it keeps the "when revoked" timestamp but scans inline into
// this struct — no separate heap allocation per row
type RefreshToken struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
	RevokedAt sql.NullTime
}