package domain

import "time"

type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	Password      string    `json:"password"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	TotalBalance  float64   `json:"total_balance"`
	Budget        float64   `json:"budget"`
	BudgetStartAt time.Time `json:"budget_start_at"`
	BudgetEndAt   time.Time `json:"budget_end_at"`
	Spent         float64   `json:"spent"`
	BankID        string    `json:"bank_id"`
	RefreshToken  string    `json:"refresh_token"`

	// TODO: Add a list of categories
	// TODO: Add a list of transactions
	// TODO: Add a list of mappings
}
