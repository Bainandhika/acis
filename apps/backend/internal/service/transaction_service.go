package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bainandhika/acis/apps/backend/internal/database"
	"github.com/Bainandhika/acis/apps/backend/internal/domain"
	"github.com/Bainandhika/acis/apps/backend/internal/dto"
	"github.com/Bainandhika/acis/apps/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type TransactionService interface {
	CreateTransaction(ctx context.Context, req dto.CreateTransactionRequest, createdBy string) (*dto.TransactionResponse, error)
	GetTransactions(ctx context.Context, walletID string, limit, offset int) ([]dto.TransactionResponse, error)
}

type transactionService struct {
	txRepo     repository.TransactionRepository
	walletRepo repository.WalletRepository
	db         *database.AppDB
}

func NewTransactionService(
	txRepo repository.TransactionRepository,
	walletRepo repository.WalletRepository,
	db *database.AppDB,
) TransactionService {
	return &transactionService{
		txRepo:     txRepo,
		walletRepo: walletRepo,
		db:         db,
	}
}

func (s *transactionService) CreateTransaction(ctx context.Context, req dto.CreateTransactionRequest, createdBy string) (*dto.TransactionResponse, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("TransactionService.CreateTransaction: failed to begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	wallet, err := s.walletRepo.GetByID(ctx, tx, req.WalletID)
	if err != nil {
		return nil, errors.New("wallet not found")
	}

	var balanceChange float64
	if req.Type == "expense" {
		if wallet.CurrentBalance < req.Amount {
			return nil, errors.New("insufficient wallet balance for expense")
		}
		balanceChange = -req.Amount
	} else if req.Type == "income" {
		balanceChange = req.Amount
	} else {
		return nil, errors.New("invalid transaction type")
	}

	// Update wallet balance
	if err = s.walletRepo.UpdateBalance(ctx, tx, req.WalletID, balanceChange); err != nil {
		return nil, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	txEntity := &domain.Transaction{
		ID:          uuid.New().String(),
		WalletID:    req.WalletID,
		Amount:      req.Amount,
		Type:        req.Type,
		Description: &req.Description,
		CreatedBy:   &createdBy,
	}

	if err = s.txRepo.Create(ctx, tx, txEntity); err != nil {
		return nil, fmt.Errorf("failed to record transaction: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info().Str("transaction_id", txEntity.ID).Str("type", req.Type).Msg("Transaction created successfully")

	return &dto.TransactionResponse{
		ID:          txEntity.ID,
		WalletID:    txEntity.WalletID,
		Amount:      txEntity.Amount,
		Type:        txEntity.Type,
		Description: req.Description,
		CreatedBy:   createdBy,
		CreatedAt:   txEntity.CreatedAt,
	}, nil
}

func (s *transactionService) GetTransactions(ctx context.Context, walletID string, limit, offset int) ([]dto.TransactionResponse, error) {
	list, err := s.txRepo.GetByWalletID(ctx, s.db, walletID, limit, offset)
	if err != nil {
		return nil, err
	}

	var res []dto.TransactionResponse
	for _, item := range list {
		desc := ""
		if item.Description != nil {
			desc = *item.Description
		}
		creator := ""
		if item.CreatedBy != nil {
			creator = *item.CreatedBy
		}

		res = append(res, dto.TransactionResponse{
			ID:          item.ID,
			WalletID:    item.WalletID,
			Amount:      item.Amount,
			Type:        item.Type,
			Description: desc,
			CreatedBy:   creator,
			CreatedAt:   item.CreatedAt,
		})
	}
	return res, nil
}
