package service

import (
	"context"
	"errors"

	"github.com/Bainandhika/acis/apps/backend/internal/database"
	"github.com/Bainandhika/acis/apps/backend/internal/domain"
	"github.com/Bainandhika/acis/apps/backend/internal/dto"
	"github.com/Bainandhika/acis/apps/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type WalletService interface {
	CreateWallet(ctx context.Context, req dto.CreateWalletRequest, createdBy string) (*dto.WalletResponse, error)
	GetWallets(ctx context.Context, familyID string) ([]dto.WalletResponse, error)
}

type walletService struct {
	walletRepo repository.WalletRepository
	db         *database.AppDB // Added to pass as DBExecutor to repository
}

func NewWalletService(walletRepo repository.WalletRepository, db *database.AppDB) WalletService {
	return &walletService{walletRepo: walletRepo, db: db}
}

func (s *walletService) CreateWallet(ctx context.Context, req dto.CreateWalletRequest, createdBy string) (*dto.WalletResponse, error) {
	if req.Name == "" {
		return nil, errors.New("wallet name cannot be empty")
	}

	walletID := uuid.New().String()

	wallet := &domain.Wallet{
		ID:             walletID,
		FamilyID:       req.FamilyID,
		Name:           req.Name,
		Description:    &req.Description,
		InitialBalance: req.InitialBalance,
		CurrentBalance: req.InitialBalance,
		MinimumLimit:   req.MinimumLimit,
		CreatedBy:      &createdBy,
	}

	// Pass s.db as the executor since this is a non-transactional operation
	if err := s.walletRepo.Create(ctx, s.db, wallet); err != nil {
		log.Error().Err(err).Str("trace_id", ctx.Value("X-Transaction-ID").(string)).Msg("Failed to create wallet in DB")
		return nil, errors.New("failed to create wallet")
	}

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

func (s *walletService) GetWallets(ctx context.Context, familyID string) ([]dto.WalletResponse, error) {
	// Pass s.db as the executor
	wallets, err := s.walletRepo.GetByFamilyID(ctx, s.db, familyID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch wallets")
		return nil, errors.New("failed to fetch wallets")
	}

	var responses []dto.WalletResponse
	for _, w := range wallets {
		desc := ""
		if w.Description != nil {
			desc = *w.Description
		}

		responses = append(responses, dto.WalletResponse{
			ID:             w.ID,
			Name:           w.Name,
			Description:    desc,
			InitialBalance: w.InitialBalance,
			CurrentBalance: w.CurrentBalance,
			MinimumLimit:   w.MinimumLimit,
		})
	}
	return responses, nil
}
