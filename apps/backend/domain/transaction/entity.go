package transaction

import "time"

type Transaction struct {
	ID          string    `db:"id" json:"id"`
	WalletID    string    `db:"wallet_id" json:"wallet_id"`
	FamilyID    string    `db:"family_id" json:"-"`
	Amount      float64   `db:"amount" json:"amount"`
	Type        string    `db:"type" json:"type"`
	Description *string   `db:"description" json:"description"`
	CreatedBy   *string   `db:"created_by" json:"created_by"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type Wallet struct {
	ID             string    `db:"id" json:"id"`
	FamilyID       string    `db:"family_id" json:"family_id"`
	Name           string    `db:"name" json:"name"`
	Description    *string   `db:"description" json:"description"`
	InitialBalance float64   `db:"initial_balance" json:"initial_balance"`
	CurrentBalance float64   `db:"current_balance" json:"current_balance"`
	MinimumLimit   float64   `db:"minimum_limit" json:"minimum_limit"`
	CreatedBy      *string   `db:"created_by" json:"created_by"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type Proposal struct {
	ID                  string     `db:"id" json:"id"`
	WalletID            string     `db:"wallet_id" json:"wallet_id"`
	Title               *string    `db:"title" json:"title"`
	Amount              float64    `db:"amount" json:"amount"`
	Description         string     `db:"description" json:"description"`
	Status              string     `db:"status" json:"status"`
	RequestType         string     `db:"request_type" json:"request_type"`
	TargetTransactionID *string    `db:"target_transaction_id" json:"target_transaction_id,omitempty"`
	Payload             *string    `db:"payload" json:"payload,omitempty"`
	ProposedBy          *string    `db:"proposed_by" json:"proposed_by"`
	ReviewedBy          *string    `db:"reviewed_by" json:"reviewed_by"`
	ReviewedAt          *time.Time `db:"reviewed_at" json:"reviewed_at"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}
