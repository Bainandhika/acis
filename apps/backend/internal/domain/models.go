package domain

import "time"

// User represents the 'users' table
type User struct {
	ID        string    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Name      string    `db:"name" json:"name"`
	GoogleID  *string   `db:"google_id" json:"google_id"` // Pointer karena optional
	AvatarURL *string   `db:"avatar_url" json:"avatar_url"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Family represents the 'families' table
type Family struct {
	ID         string    `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	InviteCode string    `db:"invite_code" json:"invite_code"`
	CreatedBy  *string   `db:"created_by" json:"created_by"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

// FamilyMember represents the 'family_members' table
type FamilyMember struct {
	ID       string    `db:"id" json:"id"`
	FamilyID string    `db:"family_id" json:"family_id"`
	UserID   string    `db:"user_id" json:"user_id"`
	Role     string    `db:"role" json:"role"` // 'admin' or 'member'
	JoinedAt time.Time `db:"joined_at" json:"joined_at"`
}

// Wallet represents the 'wallets' table
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

// Transaction represents the 'transactions' table
type Transaction struct {
	ID          string    `db:"id" json:"id"`
	WalletID    string    `db:"wallet_id" json:"wallet_id"`
	Amount      float64   `db:"amount" json:"amount"`
	Type        string    `db:"type" json:"type"` // 'income' or 'expense'
	Description *string   `db:"description" json:"description"`
	CreatedBy   *string   `db:"created_by" json:"created_by"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// Proposal represents the 'proposals' table
type Proposal struct {
	ID         string     `db:"id" json:"id"`
	WalletID   string     `db:"wallet_id" json:"wallet_id"`
	Amount     float64    `db:"amount" json:"amount"`
	Description string    `db:"description" json:"description"`
	Status     string     `db:"status" json:"status"` // 'pending', 'approved', 'rejected'
	ProposedBy *string    `db:"proposed_by" json:"proposed_by"`
	ReviewedBy *string    `db:"reviewed_by" json:"reviewed_by"`
	ReviewedAt *time.Time `db:"reviewed_at" json:"reviewed_at"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
}