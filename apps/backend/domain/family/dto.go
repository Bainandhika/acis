package family

import "time"

type CreateFamilyReq struct {
	Name          string   `json:"name" binding:"required,min=2,max=255"`
	MonthlyIncome *float64 `json:"monthly_income"`
}

type JoinFamilyReq struct {
	InviteCode string `json:"invite_code" binding:"required,len=6"`
}

type UpdateFamilySettingsReq struct {
	MonthlyIncome *float64 `json:"monthly_income"`
}

type UpdateFamilyReq struct {
	Name string `json:"name" binding:"required,min=2,max=255"`
}

type UpdateWalletReq struct {
	Name         string  `json:"name" binding:"required,min=2,max=255"`
	Description  *string `json:"description"`
	MinimumLimit float64 `json:"minimum_limit"`
}

type FamilyMemberDTO struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	UserName string    `json:"user_name,omitempty"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type FamilyDTO struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	InviteCode     string            `json:"invite_code"`
	TelegramChatID *int64            `json:"telegram_chat_id,omitempty"`
	MonthlyIncome  float64           `json:"monthly_income"`
	WalletCounter  int               `json:"wallet_counter"`
	CreatedBy      *string           `json:"created_by,omitempty"`
	Members        []FamilyMemberDTO `json:"members,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}

type CreateWalletReq struct {
	Name           string  `json:"name" binding:"required,min=2,max=255"`
	Description    *string `json:"description"`
	InitialBalance float64 `json:"initial_balance"`
	MinimumLimit   float64 `json:"minimum_limit"`
}

type WalletDTO struct {
	ID             string    `json:"id"`
	ShortID        string    `json:"short_id"`
	FamilyID       string    `json:"family_id"`
	Name           string    `json:"name"`
	Description    *string   `json:"description,omitempty"`
	InitialBalance float64   `json:"initial_balance"`
	CurrentBalance float64   `json:"current_balance"`
	MinimumLimit   float64   `json:"minimum_limit"`
	CreatedAt      time.Time `json:"created_at"`
}

type WalletBalanceDTO struct {
	WalletID       string  `json:"wallet_id"`
	ShortID        string  `json:"short_id"`
	WalletName     string  `json:"wallet_name"`
	CurrentBalance float64 `json:"current_balance"`
	MinimumLimit   float64 `json:"minimum_limit"`
}

type LowBalanceWalletDTO struct {
	WalletID       string  `db:"wallet_id" json:"wallet_id"`
	ShortID        string  `db:"short_id" json:"short_id"`
	WalletName     string  `db:"wallet_name" json:"wallet_name"`
	FamilyID       string  `db:"family_id" json:"family_id"`
	CurrentBalance float64 `db:"current_balance" json:"current_balance"`
	MinimumLimit   float64 `db:"minimum_limit" json:"minimum_limit"`
	TelegramChatID *int64  `db:"telegram_chat_id" json:"telegram_chat_id"`
}
