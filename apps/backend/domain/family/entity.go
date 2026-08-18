package family

import "time"

type Family struct {
	ID             string    `db:"id" json:"id"`
	Name           string    `db:"name" json:"name"`
	InviteCode     string    `db:"invite_code" json:"invite_code"`
	TelegramChatID *int64    `db:"telegram_chat_id" json:"telegram_chat_id,omitempty"`
	MonthlyIncome  float64   `db:"monthly_income" json:"monthly_income"`
	PrimaryBalance float64   `db:"primary_balance" json:"primary_balance"`
	WalletCounter  int       `db:"wallet_counter" json:"wallet_counter"`
	CreatedBy      *string   `db:"created_by" json:"created_by"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type FamilyMember struct {
	ID       string    `db:"id" json:"id"`
	FamilyID string    `db:"family_id" json:"family_id"`
	UserID   string    `db:"user_id" json:"user_id"`
	UserName *string   `db:"user_name" json:"user_name,omitempty"`
	Role     string    `db:"role" json:"role"`
	JoinedAt time.Time `db:"joined_at" json:"joined_at"`
}

type Wallet struct {
	ID             string    `db:"id" json:"id"`
	ShortID        string    `db:"short_id" json:"short_id"`
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
