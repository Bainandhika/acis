package bootstrap

import (
	"context"

	"github.com/Bainandhika/acis/apps/backend/domain/authentication"
	"github.com/Bainandhika/acis/apps/backend/domain/family"
	"github.com/Bainandhika/acis/apps/backend/domain/telegram"
	"github.com/Bainandhika/acis/apps/backend/domain/transaction"
)

// roleFinderAdapter bridges authentication.RoleFinder to family.FamilyRepository
type roleFinderAdapter struct {
	repo family.FamilyRepository
}

func NewRoleFinderAdapter(repo family.FamilyRepository) authentication.RoleFinder {
	return &roleFinderAdapter{repo: repo}
}

func (a *roleFinderAdapter) FindRoleByUserID(ctx context.Context, userID string) (string, error) {
	member, err := a.repo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	if member == nil {
		return "", nil
	}
	return member.Role, nil
}

type telegramTxAdapter struct {
	svc transaction.TransactionService
}

func NewTelegramTxAdapter(svc transaction.TransactionService) telegram.TransactionService {
	return &telegramTxAdapter{svc: svc}
}

func (a *telegramTxAdapter) CreateDirectTransaction(ctx context.Context, req telegram.CreateTransactionDTO) (*telegram.TransactionDTO, error) {
	txReq := transaction.CreateTransactionDTO{
		WalletID:    req.WalletID,
		UserID:      req.UserID,
		Type:        req.Type,
		Amount:      req.Amount,
		Category:    req.Category,
		Description: req.Description,
	}

	res, err := a.svc.CreateDirectTransaction(ctx, txReq)
	if err != nil {
		return nil, err
	}

	return &telegram.TransactionDTO{
		ID:          res.ID,
		WalletID:    res.WalletID,
		Amount:      res.Amount,
		Type:        res.Type,
		Description: res.Description,
	}, nil
}

type telegramFamilyAdapter struct {
	svc family.FamilyService
}

func NewTelegramFamilyAdapter(svc family.FamilyService) telegram.FamilyService {
	return &telegramFamilyAdapter{svc: svc}
}

func (a *telegramFamilyAdapter) GetWalletBalances(ctx context.Context, familyID string) ([]telegram.WalletBalanceDTO, error) {
	list, err := a.svc.GetWalletBalances(ctx, familyID)
	if err != nil {
		return nil, err
	}

	var dtos []telegram.WalletBalanceDTO
	for _, item := range list {
		dtos = append(dtos, telegram.WalletBalanceDTO{
			WalletID:       item.WalletID,
			WalletName:     item.WalletName,
			CurrentBalance: item.CurrentBalance,
			MinimumLimit:   item.MinimumLimit,
		})
	}
	return dtos, nil
}

func (a *telegramFamilyAdapter) GetLowBalanceWallets(ctx context.Context) ([]telegram.LowBalanceWalletDTO, error) {
	list, err := a.svc.GetLowBalanceWallets(ctx)
	if err != nil {
		return nil, err
	}

	var dtos []telegram.LowBalanceWalletDTO
	for _, item := range list {
		dtos = append(dtos, telegram.LowBalanceWalletDTO{
			WalletID:       item.WalletID,
			WalletName:     item.WalletName,
			FamilyID:       item.FamilyID,
			CurrentBalance: item.CurrentBalance,
			MinimumLimit:   item.MinimumLimit,
			TelegramChatID: item.TelegramChatID,
		})
	}
	return dtos, nil
}
