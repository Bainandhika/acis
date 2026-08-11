package transaction

import "time"

type CreateTransactionDTO struct {
	WalletID    string  `json:"wallet_id" binding:"required"`
	UserID      string  `json:"-"`
	Type        string  `json:"type" binding:"required,oneof=income expense"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Category    string  `json:"category" binding:"required"`
	Description *string `json:"description,omitempty"`
}

type TransactionDTO struct {
	ID          string    `json:"id"`
	WalletID    string    `json:"wallet_id"`
	UserID      *string   `json:"user_id,omitempty"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateProposalDTO struct {
	WalletID    string  `json:"wallet_id" binding:"required"`
	ProposedBy  string  `json:"-"`
	Title       string  `json:"title" binding:"required,min=3,max=100"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

type ProposalDTO struct {
	ID          string     `json:"id"`
	WalletID    string     `json:"wallet_id"`
	ProposedBy  *string    `json:"proposed_by,omitempty"`
	Title       string     `json:"title"`
	Amount      float64    `json:"amount"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	ReviewedBy  *string    `json:"reviewed_by,omitempty"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
