package validation

import (
	"errors"
	"time"
	"unicode/utf8"

	"folosy-backend/internal/domain"
)

// ValidateTransaction checks a transaction's user-supplied fields. occurred_at
// may be in the future (timezone skew, scheduled entries). category_id isn't
// checked here — the repo's insert gate verifies it exists and is owned.
func ValidateTransaction(amountMinor int64, direction domain.Direction, merchant string, occurredAt time.Time) error {
	// amount is a positive magnitude in cents; the sign lives in direction.
	if amountMinor <= 0 {
		return errors.New("amount must be greater than zero")
	}

	if direction != domain.DirectionIncome && direction != domain.DirectionExpense {
		return errors.New("direction must be income (1) or expense (2)")
	}

	// merchant is optional (DEFAULT ''); only cap its length (app-level, the
	// column is unbounded TEXT).
	if utf8.RuneCountInString(merchant) > 255 {
		return errors.New("merchant must be at most 255 characters")
	}

	if occurredAt.IsZero() {
		return errors.New("occurred_at is required")
	}

	return nil
}
