package domain

type User struct {
	ID           string  `json:"id"`
	Email        string  `json:"email"`
	Username     string  `json:"username"`
	Password     string  `json:"password"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	TotalBalance float64 `json:"total_balance"`
	Budget       float64 `json:"budget"`
	Spent        float64 `json:"spent"`
	BankID       string  `json:"bank_id"`
	RefreshToken string  `json:"refresh_token"`

	// TODO: Add a list of categories
	// TODO: Add a list of transactions
	// TODO: Add a list of mappings
}
