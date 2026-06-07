package dto

// CreateWalletRequest represents the JSON payload from frontend
type CreateWalletRequest struct {
	Name           string  `json:"name" binding:"required"`
	Description    string  `json:"description"`
	InitialBalance float64 `json:"initial_balance" binding:"required,min=0"`
	MinimumLimit   float64 `json:"minimum_limit" binding:"min=0"`
}

// WalletResponse represents the JSON response to frontend
type WalletResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	InitialBalance float64 `json:"initial_balance"`
	CurrentBalance float64 `json:"current_balance"`
	MinimumLimit   float64 `json:"minimum_limit"`
}
