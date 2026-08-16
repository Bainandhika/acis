package bootstrap

import (
	"context"

	"github.com/Bainandhika/acis/apps/backend/domain/authentication"
	"github.com/Bainandhika/acis/apps/backend/domain/family"
	"github.com/Bainandhika/acis/apps/backend/domain/telegram"
	"github.com/Bainandhika/acis/apps/backend/domain/transaction"
	"github.com/Bainandhika/acis/apps/backend/shared/cache"
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

type authSessionAdapter struct {
	otpCache *cache.OTPCache
	authRepo authentication.AuthRepository
}

func NewAuthSessionAdapter(otpCache *cache.OTPCache, authRepo authentication.AuthRepository) telegram.AuthSessionResolver {
	return &authSessionAdapter{
		otpCache: otpCache,
		authRepo: authRepo,
	}
}

func (a *authSessionAdapter) ResolveAuthSession(ctx context.Context, sessionToken string, chatID int64) (string, error) {
	if a.otpCache == nil {
		return "", nil
	}
	session, err := a.otpCache.GetAuthSession(ctx, sessionToken)
	if err != nil || session == nil {
		return "", err
	}

	// If user exists with email + phone, update their telegram_chat_id
	if a.authRepo != nil {
		user, err := a.authRepo.FindByEmailAndPhone(ctx, session.Email, session.PhoneNumber)
		if err == nil && user != nil {
			_ = a.authRepo.UpdateTelegramChatID(ctx, user.ID, chatID)
		}
	}

	return session.OTP, nil
}

func (a *authSessionAdapter) GetActiveOTP(ctx context.Context, email, phone string, chatID int64) (string, error) {
	if a.otpCache == nil {
		return "", nil
	}
	// If user exists, link chatID
	if a.authRepo != nil {
		user, err := a.authRepo.FindByEmailAndPhone(ctx, email, phone)
		if err == nil && user != nil {
			_ = a.authRepo.UpdateTelegramChatID(ctx, user.ID, chatID)
		}
	}
	return "", nil
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

func (a *telegramFamilyAdapter) FindByTelegramChatID(ctx context.Context, chatID int64) (*telegram.FamilyDTO, error) {
	res, err := a.svc.FindByTelegramChatID(ctx, chatID)
	if err != nil {
		return nil, err
	}
	return &telegram.FamilyDTO{
		ID:             res.ID,
		Name:           res.Name,
		InviteCode:     res.InviteCode,
		TelegramChatID: res.TelegramChatID,
	}, nil
}

func (a *telegramFamilyAdapter) LinkTelegramChatID(ctx context.Context, inviteCode string, chatID int64) error {
	return a.svc.LinkTelegramChatID(ctx, inviteCode, chatID)
}

func (a *telegramFamilyAdapter) GetMembers(ctx context.Context, familyID string) ([]telegram.FamilyMemberDTO, error) {
	members, err := a.svc.GetMembers(ctx, familyID)
	if err != nil {
		return nil, err
	}
	var dtos []telegram.FamilyMemberDTO
	for _, m := range members {
		dtos = append(dtos, telegram.FamilyMemberDTO{
			ID:     m.ID,
			UserID: m.UserID,
			Role:   m.Role,
		})
	}
	return dtos, nil
}
