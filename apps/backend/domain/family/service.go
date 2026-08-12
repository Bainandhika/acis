package family

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/google/uuid"
	zerolog "github.com/rs/zerolog/log"
)

type FamilyService interface {
	CreateFamily(ctx context.Context, userID, name string) (*FamilyDTO, error)
	JoinFamily(ctx context.Context, userID, inviteCode string) (*FamilyDTO, error)
	GetMyFamily(ctx context.Context, userID string) (*FamilyDTO, error)
	CreateWallet(ctx context.Context, userID, familyID string, req CreateWalletReq) (*WalletDTO, error)
	GetWallets(ctx context.Context, familyID string) ([]WalletDTO, error)
	GetWalletBalances(ctx context.Context, familyID string) ([]WalletBalanceDTO, error)
	GetLowBalanceWallets(ctx context.Context) ([]LowBalanceWalletDTO, error)
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

func (s *familyService) CreateFamily(ctx context.Context, userID, name string) (*FamilyDTO, error) {
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
		ID:         uuid.NewString(),
		Name:       name,
		InviteCode: inviteCode,
		CreatedBy:  &userID,
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
		ID:         family.ID,
		Name:       family.Name,
		InviteCode: family.InviteCode,
		CreatedBy:  family.CreatedBy,
		CreatedAt:  family.CreatedAt,
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
		ID:         family.ID,
		Name:       family.Name,
		InviteCode: family.InviteCode,
		CreatedBy:  family.CreatedBy,
		CreatedAt:  family.CreatedAt,
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
		memberDTOs = append(memberDTOs, FamilyMemberDTO{
			ID:       m.ID,
			UserID:   m.UserID,
			Role:     m.Role,
			JoinedAt: m.JoinedAt,
		})
	}

	return &FamilyDTO{
		ID:         family.ID,
		Name:       family.Name,
		InviteCode: family.InviteCode,
		CreatedBy:  family.CreatedBy,
		Members:    memberDTOs,
		CreatedAt:  family.CreatedAt,
	}, nil
}

func (s *familyService) CreateWallet(ctx context.Context, userID, familyID string, req CreateWalletReq) (*WalletDTO, error) {
	wallet := &Wallet{
		ID:             uuid.NewString(),
		FamilyID:       familyID,
		Name:           req.Name,
		Description:    req.Description,
		InitialBalance: req.InitialBalance,
		CurrentBalance: req.InitialBalance,
		MinimumLimit:   req.MinimumLimit,
		CreatedBy:      &userID,
	}

	if err := s.repo.CreateWallet(ctx, wallet); err != nil {
		zerolog.Error().Err(err).Msg("Failed to create wallet")
		return nil, errors.New("failed to create wallet")
	}

	return &WalletDTO{
		ID:             wallet.ID,
		FamilyID:       wallet.FamilyID,
		Name:           wallet.Name,
		Description:    wallet.Description,
		InitialBalance: wallet.InitialBalance,
		CurrentBalance: wallet.CurrentBalance,
		MinimumLimit:   wallet.MinimumLimit,
		CreatedAt:      wallet.CreatedAt,
	}, nil
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

func (s *familyService) GetWalletBalances(ctx context.Context, familyID string) ([]WalletBalanceDTO, error) {
	wallets, err := s.repo.GetWalletsByFamilyID(ctx, familyID)
	if err != nil {
		return nil, errors.New("failed to fetch wallet balances")
	}

	var dtos []WalletBalanceDTO
	for _, w := range wallets {
		dtos = append(dtos, WalletBalanceDTO{
			WalletID:       w.ID,
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
