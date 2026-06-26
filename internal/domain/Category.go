package domain

import "time"

// Category is a user-owned spending bucket (e.g. "Groceries", "Rent") with a
// budget limit. "Spent" is NOT stored here — it's derived by summing the
// category's transactions at read time.
type Category struct {
	ID     string `json:"id"`
	UserID string `json:"-"` // internal: ownership + scanning, never echoed to the client
	Name   string `json:"name"`
	Icon   string `json:"icon"` // opaque label the client picks/renders; "" = none

	// BudgetLimitMinor is the budget in integer minor units (cents). 0
	BudgetLimitMinor int64 `json:"budget_limit_minor"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}