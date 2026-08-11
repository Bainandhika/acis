package dto

import "time"

// CreateTransactionRequest represents the JSON payload to create a direct transaction
type CreateTransactionRequest struct {
	WalletID    string  `json:"wallet_id" binding:"required,uuid"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Type        string  `json:"type" binding:"required,oneof=income expense"`
	Description string  `json:"description" binding:"required,max=255"`
}

// TransactionResponse represents transaction details in JSON response
type TransactionResponse struct {
	ID          string    `json:"id"`
	WalletID    string    `json:"wallet_id"`
	Amount      float64   `json:"amount"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}
