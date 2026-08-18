package transaction

import (
	"encoding/json"
	"time"
)

type CreateTransactionDTO struct {
	WalletID    string  `json:"wallet_id"`
	UserID      string  `json:"-"`
	FamilyID    string  `json:"-"`
	Type        string  `json:"type" binding:"required,oneof=income expense allocation"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description *string `json:"description,omitempty"`
}

type UpdateTransactionDTO struct {
	WalletID    string  `json:"wallet_id"`
	Type        string  `json:"type" binding:"required,oneof=income expense"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description *string `json:"description,omitempty"`
	UserID      string  `json:"-"`
	FamilyID    string  `json:"-"`
}

type TransactionDTO struct {
	ID          string    `json:"id"`
	WalletID    string    `json:"wallet_id"`
	UserID      *string   `json:"user_id,omitempty"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateProposalDTO struct {
	WalletID            string          `json:"wallet_id" binding:"required"`
	ProposedBy          string          `json:"-"`
	Title               string          `json:"title" binding:"required,min=3,max=100"`
	Amount              float64         `json:"amount"`
	Description         string          `json:"description"`
	RequestType         string          `json:"request_type" binding:"required,oneof=add_transaction edit_transaction delete_transaction"`
	TargetTransactionID *string         `json:"target_transaction_id,omitempty"`
	Payload             json.RawMessage `json:"payload,omitempty"`
}

type ProposalDTO struct {
	ID                  string          `json:"id"`
	WalletID            string          `json:"wallet_id"`
	ProposedBy          *string         `json:"proposed_by,omitempty"`
	Title               string          `json:"title"`
	Amount              float64         `json:"amount"`
	Description         string          `json:"description"`
	Status              string          `json:"status"`
	RequestType         string          `json:"request_type"`
	TargetTransactionID *string         `json:"target_transaction_id,omitempty"`
	Payload             json.RawMessage `json:"payload,omitempty"`
	ReviewedBy          *string         `json:"reviewed_by,omitempty"`
	ReviewedAt          *time.Time      `json:"reviewed_at,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
}
