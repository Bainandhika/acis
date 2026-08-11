package family

import "time"

type CreateFamilyReq struct {
	Name string `json:"name" binding:"required,min=3,max=50"`
}

type JoinFamilyReq struct {
	InviteCode string `json:"invite_code" binding:"required,len=6"`
}

type FamilyDTO struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	InviteCode string            `json:"invite_code"`
	CreatedBy  *string           `json:"created_by,omitempty"`
	Members    []FamilyMemberDTO `json:"members,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

type FamilyMemberDTO struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	UserName string    `json:"user_name"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type CreateWalletReq struct {
	Name           string  `json:"name" binding:"required,min=2,max=50"`
	Description    *string `json:"description,omitempty"`
	InitialBalance float64 `json:"initial_balance" binding:"gte=0"`
	MinimumLimit   float64 `json:"minimum_limit" binding:"gte=0"`
}

type WalletDTO struct {
	ID             string    `json:"id"`
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
	WalletName     string  `json:"wallet_name"`
	CurrentBalance float64 `json:"current_balance"`
	MinimumLimit   float64 `json:"minimum_limit"`
}

type LowBalanceWalletDTO struct {
	WalletID       string  `json:"wallet_id"`
	WalletName     string  `json:"wallet_name"`
	FamilyID       string  `json:"family_id"`
	CurrentBalance float64 `json:"current_balance"`
	MinimumLimit   float64 `json:"minimum_limit"`
	TelegramChatID *int64  `json:"telegram_chat_id,omitempty"`
}
