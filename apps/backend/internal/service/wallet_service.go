package service

import (
	"context"
	"errors"

	"github.com/Bainandhika/acis/apps/backend/internal/domain"
	"github.com/Bainandhika/acis/apps/backend/internal/dto"
	"github.com/Bainandhika/acis/apps/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// WalletService defines business logic contracts
type WalletService interface {
	CreateWallet(ctx context.Context, req dto.CreateWalletRequest, createdBy string) (*dto.WalletResponse, error)
}

type walletService struct {
	walletRepo repository.WalletRepository
}

// NewWalletService creates a new WalletService (Manual DI)
func NewWalletService(walletRepo repository.WalletRepository) WalletService {
	return &walletService{walletRepo: walletRepo}
}

// CreateWallet handles the business logic of creating a wallet
func (s *walletService) CreateWallet(ctx context.Context, req dto.CreateWalletRequest, createdBy string) (*dto.WalletResponse, error) {
	// 1. Business Validation
	if req.Name == "" {
		return nil, errors.New("wallet name cannot be empty")
	}

	// 2. Generate UUID for the new wallet
	walletID := uuid.New().String()

	// TODO: Replace with actual IDs from Auth context later
	familyID := "00000000-0000-0000-0000-000000000002"
	createdByUUID := "00000000-0000-0000-0000-000000000001"

	// 3. Map DTO to Domain Model
	wallet := &domain.Wallet{
		ID:             walletID,
		FamilyID:       familyID, // <-- FIX: Isi dengan valid UUID
		Name:           req.Name,
		Description:    &req.Description,
		InitialBalance: req.InitialBalance,
		CurrentBalance: req.InitialBalance,
		MinimumLimit:   req.MinimumLimit,
		CreatedBy:      &createdByUUID, // <-- FIX: Isi dengan valid UUID
	}

	// 4. Call Repository
	if err := s.walletRepo.Create(ctx, wallet); err != nil {
		log.Error().Err(err).Str("trace_id", ctx.Value("X-Transaction-ID").(string)).Msg("Failed to create wallet in DB")
		return nil, errors.New("failed to create wallet")
	}

	// 5. Map Domain back to Response DTO
	response := &dto.WalletResponse{
		ID:             wallet.ID,
		Name:           wallet.Name,
		Description:    req.Description,
		InitialBalance: wallet.InitialBalance,
		CurrentBalance: wallet.CurrentBalance,
		MinimumLimit:   wallet.MinimumLimit,
	}

	log.Info().Str("wallet_id", walletID).Msg("Wallet created successfully")
	return response, nil
}
