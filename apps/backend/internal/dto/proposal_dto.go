package dto

import "time"

type CreateProposalRequest struct {
	WalletID    string  `json:"wallet_id" binding:"required,uuid"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description" binding:"required"`
}

type ProposalResponse struct {
	ID          string    `json:"id"`
	WalletID    string    `json:"wallet_id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	ProposedBy  string    `json:"proposed_by"`
	CreatedAt   time.Time `json:"created_at"`
}
