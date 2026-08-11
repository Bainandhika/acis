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

type TransactionService interface {
	CreateDirectTransaction(ctx context.Context, req CreateTransactionDTO) (*TransactionDTO, error)
}

type FamilyService interface {
	GetWalletBalances(ctx context.Context, familyID string) ([]WalletBalanceDTO, error)
	GetLowBalanceWallets(ctx context.Context) ([]LowBalanceWalletDTO, error)
}
