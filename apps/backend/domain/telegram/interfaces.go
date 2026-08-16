package telegram

import "context"

type TransactionDTO struct {
	ID          string  `json:"id"`
	WalletID    string  `json:"wallet_id"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Description *string `json:"description,omitempty"`
}

type CreateTransactionDTO struct {
	WalletID    string  `json:"wallet_id"`
	UserID      string  `json:"-"`
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description *string `json:"description,omitempty"`
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

type FamilyDTO struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	InviteCode     string `json:"invite_code"`
	TelegramChatID *int64 `json:"telegram_chat_id,omitempty"`
}

type FamilyMemberDTO struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type TransactionService interface {
	CreateDirectTransaction(ctx context.Context, req CreateTransactionDTO) (*TransactionDTO, error)
}

type FamilyService interface {
	GetWalletBalances(ctx context.Context, familyID string) ([]WalletBalanceDTO, error)
	GetLowBalanceWallets(ctx context.Context) ([]LowBalanceWalletDTO, error)
	FindByTelegramChatID(ctx context.Context, chatID int64) (*FamilyDTO, error)
	LinkTelegramChatID(ctx context.Context, inviteCode string, chatID int64) error
	GetMembers(ctx context.Context, familyID string) ([]FamilyMemberDTO, error)
}

type AuthSessionResolver interface {
	ResolveAuthSession(ctx context.Context, sessionToken string, chatID int64) (string, error)
	GetActiveOTP(ctx context.Context, email, phone string, chatID int64) (string, error)
}
