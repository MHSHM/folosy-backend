package validation

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// ValidateCategory checks a category's user-supplied fields. Icon is not
// validated — it's an opaque label the server never interprets.
func ValidateCategory(name string, budgetLimitMinor int64) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name is required")
	}

	// The column is VARCHAR(255), which counts characters (runes), not bytes —
	// so use RuneCountInString, not len(), or a multi-byte name would be measured
	// in bytes and rejected too early.
	if utf8.RuneCountInString(name) > 255 {
		return errors.New("name must be at most 255 characters")
	}

	// A budget limit is a count of cents; negative is meaningless.
	if budgetLimitMinor < 0 {
		return errors.New("budget limit cannot be negative")
	}

	return nil
}
