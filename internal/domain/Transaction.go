package domain

import "time"

// Direction is the income-vs-expense axis of a transaction, stored as a SMALLINT
// (1 = income, 2 = expense).
type Direction int16

const (
	DirectionIncome  Direction = 1
	DirectionExpense Direction = 2
)

// Transaction is a single money movement in a user's ledger. It mirrors real
// money: an expense lowers the user's balance, an income raises it.
type Transaction struct {
	ID     string `json:"id"`
	UserID string `json:"-"` // internal: ownership + scanning, never echoed

	// CategoryID is null for uncategorized — a valid, routine state. NullString is
	// our JSON-clean sql.NullString: a value type that
	// scans/inserts a true NULL and serializes as the id string or null.
	CategoryID NullString `json:"category_id"`

	// AmountMinor is the MAGNITUDE in integer minor units (cents)
	AmountMinor int64     `json:"amount_minor"`
	Direction   Direction `json:"direction"`

	Merchant string `json:"merchant"` // opaque label: "Starbucks", "Salary Deposit"

	// OccurredAt = when the money actually moved (may be backdated by the user or
	// set by the SMS parser). Distinct from CreatedAt = when we recorded the row.
	OccurredAt time.Time `json:"occurred_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
