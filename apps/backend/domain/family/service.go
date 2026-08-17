package family

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/google/uuid"
)

type FamilyService interface {
	CreateFamily(ctx context.Context, userID, name string, monthlyIncome float64) (*FamilyDTO, error)
	JoinFamily(ctx context.Context, userID, inviteCode string) (*FamilyDTO, error)
	GetMyFamily(ctx context.Context, userID string) (*FamilyDTO, error)
	UpdateFamilySettings(ctx context.Context, familyID string, req UpdateFamilySettingsReq) error
	UpdateFamilyName(ctx context.Context, familyID string, name string) error
	DisconnectTelegram(ctx context.Context, familyID string) error
	FindByInviteCode(ctx context.Context, inviteCode string) (*FamilyDTO, error)
	FindByTelegramChatID(ctx context.Context, chatID int64) (*FamilyDTO, error)
	LinkTelegramChatID(ctx context.Context, inviteCode string, chatID int64) error
	CreateWallet(ctx context.Context, userID, familyID string, req CreateWalletReq) (*WalletDTO, error)
	UpdateWallet(ctx context.Context, walletID, familyID string, req UpdateWalletReq) (*WalletDTO, error)
	DeleteWallet(ctx context.Context, walletID, familyID string) error
	GetWallets(ctx context.Context, familyID string) ([]WalletDTO, error)
	FindWalletByShortID(ctx context.Context, shortID string) (*WalletDTO, error)
	GetWalletBalances(ctx context.Context, familyID string) ([]WalletBalanceDTO, error)
	GetLowBalanceWallets(ctx context.Context) ([]LowBalanceWalletDTO, error)
	GetMembers(ctx context.Context, familyID string) ([]FamilyMemberDTO, error)
	RemoveMember(ctx context.Context, requesterUserID, memberID, familyID string) error
}

type familyService struct {
	repo FamilyRepository
	db   *database.AppDB
}

func NewService(repo FamilyRepository, db *database.AppDB) FamilyService {
	return &familyService{
		repo: repo,
		db:   db,
	}
}

func (s *familyService) CreateFamily(ctx context.Context, userID, name string, monthlyIncome float64) (*FamilyDTO, error) {
	existing, err := s.repo.FindFamilyByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to check existing family membership")
	}
	if existing != nil {
		return nil, errors.New("user is already a member of a family")
	}

	inviteCode, err := generateInviteCode()
	if err != nil {
		return nil, errors.New("failed to generate invite code")
	}

	family := &Family{
		ID:            uuid.NewString(),
		Name:          name,
		InviteCode:    inviteCode,
		MonthlyIncome: monthlyIncome,
		WalletCounter: 0,
		CreatedBy:     &userID,
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.New("failed to start database transaction")
	}
	defer tx.Rollback()

	if err := s.repo.CreateFamily(ctx, tx, family); err != nil {
		return nil, errors.New("failed to create family record")
	}

	member := &FamilyMember{
		ID:       uuid.NewString(),
		FamilyID: family.ID,
		UserID:   userID,
		Role:     "admin",
	}
	if err := s.repo.AddMember(ctx, tx, member); err != nil {
		return nil, errors.New("failed to add admin family member")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.New("failed to commit family creation")
	}

	return &FamilyDTO{
		ID:            family.ID,
		Name:          family.Name,
		InviteCode:    family.InviteCode,
		MonthlyIncome: family.MonthlyIncome,
		WalletCounter: family.WalletCounter,
		CreatedBy:     family.CreatedBy,
		CreatedAt:     family.CreatedAt,
	}, nil
}

func (s *familyService) JoinFamily(ctx context.Context, userID, inviteCode string) (*FamilyDTO, error) {
	existing, err := s.repo.FindFamilyByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to check existing family membership")
	}
	if existing != nil {
		return nil, errors.New("user is already a member of a family")
	}

	family, err := s.repo.FindByInviteCode(ctx, inviteCode)
	if err != nil || family == nil {
		return nil, errors.New("invalid invite code")
	}

	member := &FamilyMember{
		ID:       uuid.NewString(),
		FamilyID: family.ID,
		UserID:   userID,
		Role:     "member",
	}
	if err := s.repo.AddMember(ctx, s.db, member); err != nil {
		return nil, errors.New("failed to join family")
	}

	return &FamilyDTO{
		ID:             family.ID,
		Name:           family.Name,
		InviteCode:     family.InviteCode,
		TelegramChatID: family.TelegramChatID,
		MonthlyIncome:  family.MonthlyIncome,
		WalletCounter:  family.WalletCounter,
		CreatedBy:      family.CreatedBy,
		CreatedAt:      family.CreatedAt,
	}, nil
}

func (s *familyService) GetMyFamily(ctx context.Context, userID string) (*FamilyDTO, error) {
	family, err := s.repo.FindFamilyByUserID(ctx, userID)
	if err != nil || family == nil {
		return nil, errors.New("family not found for user")
	}

	members, err := s.repo.GetMembers(ctx, family.ID)
	if err != nil {
		return nil, errors.New("failed to fetch family members")
	}

	var memberDTOs []FamilyMemberDTO
	for _, m := range members {
		userName := ""
		if m.UserName != nil {
			userName = *m.UserName
		}
		memberDTOs = append(memberDTOs, FamilyMemberDTO{
			ID:       m.ID,
			UserID:   m.UserID,
			UserName: userName,
			Role:     m.Role,
			JoinedAt: m.JoinedAt,
		})
	}

	return &FamilyDTO{
		ID:             family.ID,
		Name:           family.Name,
		InviteCode:     family.InviteCode,
		TelegramChatID: family.TelegramChatID,
		MonthlyIncome:  family.MonthlyIncome,
		WalletCounter:  family.WalletCounter,
		CreatedBy:      family.CreatedBy,
		Members:        memberDTOs,
		CreatedAt:      family.CreatedAt,
	}, nil
}

func (s *familyService) UpdateFamilySettings(ctx context.Context, familyID string, req UpdateFamilySettingsReq) error {
	if req.MonthlyIncome != nil {
		if *req.MonthlyIncome < 0 {
			return errors.New("monthly income cannot be negative")
		}
		return s.repo.UpdateMonthlyIncome(ctx, familyID, *req.MonthlyIncome)
	}
	return nil
}

func (s *familyService) UpdateFamilyName(ctx context.Context, familyID string, name string) error {
	if name == "" {
		return errors.New("family name cannot be empty")
	}
	return s.repo.UpdateFamilyName(ctx, familyID, name)
}

func (s *familyService) DisconnectTelegram(ctx context.Context, familyID string) error {
	return s.repo.UpdateTelegramChatID(ctx, familyID, nil)
}

func (s *familyService) FindByInviteCode(ctx context.Context, inviteCode string) (*FamilyDTO, error) {
	family, err := s.repo.FindByInviteCode(ctx, inviteCode)
	if err != nil || family == nil {
		return nil, errors.New("family not found")
	}
	return &FamilyDTO{
		ID:             family.ID,
		Name:           family.Name,
		InviteCode:     family.InviteCode,
		TelegramChatID: family.TelegramChatID,
		MonthlyIncome:  family.MonthlyIncome,
		WalletCounter:  family.WalletCounter,
		CreatedBy:      family.CreatedBy,
		CreatedAt:      family.CreatedAt,
	}, nil
}

func (s *familyService) FindByTelegramChatID(ctx context.Context, chatID int64) (*FamilyDTO, error) {
	family, err := s.repo.FindByTelegramChatID(ctx, chatID)
	if err != nil || family == nil {
		return nil, errors.New("family not found for this telegram chat")
	}
	return &FamilyDTO{
		ID:             family.ID,
		Name:           family.Name,
		InviteCode:     family.InviteCode,
		TelegramChatID: family.TelegramChatID,
		MonthlyIncome:  family.MonthlyIncome,
		WalletCounter:  family.WalletCounter,
		CreatedBy:      family.CreatedBy,
		CreatedAt:      family.CreatedAt,
	}, nil
}

func (s *familyService) LinkTelegramChatID(ctx context.Context, inviteCode string, chatID int64) error {
	family, err := s.repo.FindByInviteCode(ctx, inviteCode)
	if err != nil || family == nil {
		return errors.New("invalid invite code")
	}
	return s.repo.UpdateTelegramChatID(ctx, family.ID, &chatID)
}

func (s *familyService) GetMembers(ctx context.Context, familyID string) ([]FamilyMemberDTO, error) {
	members, err := s.repo.GetMembers(ctx, familyID)
	if err != nil {
		return nil, errors.New("failed to fetch members")
	}
	var dtos []FamilyMemberDTO
	for _, m := range members {
		userName := ""
		if m.UserName != nil {
			userName = *m.UserName
		}
		dtos = append(dtos, FamilyMemberDTO{
			ID:       m.ID,
			UserID:   m.UserID,
			UserName: userName,
			Role:     m.Role,
			JoinedAt: m.JoinedAt,
		})
	}
	return dtos, nil
}

func (s *familyService) CreateWallet(ctx context.Context, userID, familyID string, req CreateWalletReq) (*WalletDTO, error) {
	counter, err := s.repo.IncrementWalletCounter(ctx, familyID)
	if err != nil {
		slog.Error("Failed to increment wallet counter", slog.Any("error", err))
		return nil, errors.New("failed to generate wallet sequence")
	}

	fam, err := s.repo.FindFamilyByID(ctx, familyID)
	if err != nil || fam == nil {
		return nil, errors.New("family not found")
	}

	shortID := fmt.Sprintf("%s-%d", fam.InviteCode, counter)

	wallet := &Wallet{
		ID:             uuid.NewString(),
		ShortID:        shortID,
		FamilyID:       familyID,
		Name:           req.Name,
		Description:    req.Description,
		InitialBalance: req.InitialBalance,
		CurrentBalance: req.InitialBalance,
		MinimumLimit:   req.MinimumLimit,
		CreatedBy:      &userID,
	}

	if err := s.repo.CreateWallet(ctx, wallet); err != nil {
		slog.Error("Failed to create wallet", slog.Any("error", err))
		return nil, errors.New("failed to create wallet")
	}

	return &WalletDTO{
		ID:             wallet.ID,
		ShortID:        wallet.ShortID,
		FamilyID:       wallet.FamilyID,
		Name:           wallet.Name,
		Description:    wallet.Description,
		InitialBalance: wallet.InitialBalance,
		CurrentBalance: wallet.CurrentBalance,
		MinimumLimit:   wallet.MinimumLimit,
		CreatedAt:      wallet.CreatedAt,
	}, nil
}

func (s *familyService) UpdateWallet(ctx context.Context, walletID, familyID string, req UpdateWalletReq) (*WalletDTO, error) {
	wallet, err := s.repo.GetWalletByID(ctx, walletID)
	if err != nil || wallet == nil {
		return nil, errors.New("wallet not found")
	}
	if wallet.FamilyID != familyID {
		return nil, errors.New("unauthorized wallet update")
	}

	if err := s.repo.UpdateWallet(ctx, walletID, req.Name, req.Description, req.MinimumLimit); err != nil {
		slog.Error("Failed to update wallet", slog.Any("error", err))
		return nil, errors.New("failed to update wallet")
	}

	updated, err := s.repo.GetWalletByID(ctx, walletID)
	if err != nil || updated == nil {
		return nil, errors.New("failed to retrieve updated wallet")
	}

	return &WalletDTO{
		ID:             updated.ID,
		ShortID:        updated.ShortID,
		FamilyID:       updated.FamilyID,
		Name:           updated.Name,
		Description:    updated.Description,
		InitialBalance: updated.InitialBalance,
		CurrentBalance: updated.CurrentBalance,
		MinimumLimit:   updated.MinimumLimit,
		CreatedAt:      updated.CreatedAt,
	}, nil
}

func (s *familyService) DeleteWallet(ctx context.Context, walletID, familyID string) error {
	wallet, err := s.repo.GetWalletByID(ctx, walletID)
	if err != nil || wallet == nil {
		return errors.New("wallet not found")
	}
	if wallet.FamilyID != familyID {
		return errors.New("unauthorized wallet delete")
	}

	return s.repo.DeleteWallet(ctx, walletID, familyID)
}

func (s *familyService) GetWallets(ctx context.Context, familyID string) ([]WalletDTO, error) {
	wallets, err := s.repo.GetWalletsByFamilyID(ctx, familyID)
	if err != nil {
		return nil, errors.New("failed to fetch wallets")
	}

	var dtos []WalletDTO
	for _, w := range wallets {
		dtos = append(dtos, WalletDTO{
			ID:             w.ID,
			ShortID:        w.ShortID,
			FamilyID:       w.FamilyID,
			Name:           w.Name,
			Description:    w.Description,
			InitialBalance: w.InitialBalance,
			CurrentBalance: w.CurrentBalance,
			MinimumLimit:   w.MinimumLimit,
			CreatedAt:      w.CreatedAt,
		})
	}
	return dtos, nil
}

func (s *familyService) FindWalletByShortID(ctx context.Context, shortID string) (*WalletDTO, error) {
	w, err := s.repo.FindWalletByShortID(ctx, shortID)
	if err != nil || w == nil {
		return nil, errors.New("wallet not found")
	}
	return &WalletDTO{
		ID:             w.ID,
		ShortID:        w.ShortID,
		FamilyID:       w.FamilyID,
		Name:           w.Name,
		Description:    w.Description,
		InitialBalance: w.InitialBalance,
		CurrentBalance: w.CurrentBalance,
		MinimumLimit:   w.MinimumLimit,
		CreatedAt:      w.CreatedAt,
	}, nil
}

func (s *familyService) GetWalletBalances(ctx context.Context, familyID string) ([]WalletBalanceDTO, error) {
	wallets, err := s.repo.GetWalletsByFamilyID(ctx, familyID)
	if err != nil {
		return nil, errors.New("failed to fetch wallet balances")
	}

	var dtos []WalletBalanceDTO
	for _, w := range wallets {
		dtos = append(dtos, WalletBalanceDTO{
			WalletID:       w.ID,
			ShortID:        w.ShortID,
			WalletName:     w.Name,
			CurrentBalance: w.CurrentBalance,
			MinimumLimit:   w.MinimumLimit,
		})
	}
	return dtos, nil
}

func (s *familyService) GetLowBalanceWallets(ctx context.Context) ([]LowBalanceWalletDTO, error) {
	return s.repo.GetLowBalanceWallets(ctx)
}

func (s *familyService) RemoveMember(ctx context.Context, requesterUserID, memberID, familyID string) error {
	member, err := s.repo.FindMemberByID(ctx, memberID)
	if err != nil {
		return err
	}
	if member == nil {
		return errors.New("member not found")
	}

	if member.FamilyID != familyID {
		return errors.New("unauthorized member deletion")
	}

	if member.UserID == requesterUserID {
		return errors.New("cannot remove yourself from the family")
	}

	if member.Role == "admin" {
		members, err := s.repo.GetMembers(ctx, familyID)
		if err != nil {
			return err
		}
		adminCount := 0
		for _, m := range members {
			if m.Role == "admin" {
				adminCount++
			}
		}
		if adminCount <= 1 {
			return errors.New("cannot remove the last admin of the family")
		}
	}

	return s.repo.RemoveMember(ctx, memberID, familyID)
}

func generateInviteCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[num.Int64()]
	}
	return fmt.Sprintf("%s", string(b)), nil
}
