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

type ProposalService interface {
	CreateProposal(ctx context.Context, req dto.CreateProposalRequest, proposedBy string) (*dto.ProposalResponse, error)
	ApproveProposal(ctx context.Context, proposalID string, reviewerID string) error
	RejectProposal(ctx context.Context, proposalID string, reviewerID string) error
}

type proposalService struct {
	proposalRepo repository.ProposalRepository
	walletRepo   repository.WalletRepository
	db           *database.AppDB // Needed to begin transactions
}

func NewProposalService(
	proposalRepo repository.ProposalRepository,
	walletRepo repository.WalletRepository,
	db *database.AppDB,
) ProposalService {
	return &proposalService{
		proposalRepo: proposalRepo,
		walletRepo:   walletRepo,
		db:           db,
	}
}

func (s *proposalService) CreateProposal(ctx context.Context, req dto.CreateProposalRequest, proposedBy string) (*dto.ProposalResponse, error) {
	// 1. Begin Atomic Transaction
	tx, err := s.db.Beginx()
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		return nil, errors.New("failed to start transaction")
	}

	// 2. Safety Net: Defer Rollback
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 3. Business Logic: Check Wallet Balance (using tx as executor)
	wallet, err := s.walletRepo.GetByID(ctx, tx, req.WalletID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get wallet")
		return nil, errors.New("wallet not found")
	}

	if wallet.CurrentBalance < req.Amount {
		err = errors.New("insufficient balance")
		return nil, err
	}

	// 4. Create Proposal (using tx as executor)
	proposalID := uuid.New().String()
	proposal := &domain.Proposal{
		ID:          proposalID,
		WalletID:    req.WalletID,
		Amount:      req.Amount,
		Description: req.Description,
		Status:      "pending",
		ProposedBy:  &proposedBy,
	}

	err = s.proposalRepo.Create(ctx, tx, proposal)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create proposal")
		return nil, errors.New("failed to create proposal")
	}

	// 5. Commit Transaction
	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Msg("Failed to commit transaction")
		return nil, errors.New("failed to save proposal")
	}

	// 6. Map to Response DTO
	response := &dto.ProposalResponse{
		ID:          proposal.ID,
		WalletID:    proposal.WalletID,
		Amount:      proposal.Amount,
		Description: proposal.Description,
		Status:      proposal.Status,
		ProposedBy:  proposedBy,
		CreatedAt:   proposal.CreatedAt,
	}

	log.Info().Str("proposal_id", proposalID).Msg("Proposal created successfully")
	return response, nil
}

// ApproveProposal handles atomic proposal approval, wallet balance deduction, and transaction recording.
func (s *proposalService) ApproveProposal(ctx context.Context, proposalID string, reviewerID string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("ProposalService.ApproveProposal: failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1. Lock proposal for update
	proposal, err := s.proposalRepo.GetByIDForUpdate(ctx, tx, proposalID)
	if err != nil {
		return fmt.Errorf("ProposalService.ApproveProposal: proposal not found: %w", err)
	}

	if proposal.Status != "pending" {
		err = errors.New("proposal is not in pending status")
		return err
	}

	// 2. Check wallet balance
	wallet, err := s.walletRepo.GetByID(ctx, tx, proposal.WalletID)
	if err != nil {
		return fmt.Errorf("ProposalService.ApproveProposal: wallet not found: %w", err)
	}

	if wallet.CurrentBalance < proposal.Amount {
		err = errors.New("insufficient wallet balance")
		return err
	}

	// 3. Deduct balance from wallet
	if err = s.walletRepo.UpdateBalance(ctx, tx, proposal.WalletID, -proposal.Amount); err != nil {
		return fmt.Errorf("ProposalService.ApproveProposal: failed to deduct wallet balance: %w", err)
	}

	// 4. Mark proposal approved
	if err = s.proposalRepo.ApproveProposal(ctx, tx, proposalID, reviewerID); err != nil {
		return fmt.Errorf("ProposalService.ApproveProposal: failed to approve proposal: %w", err)
	}

	// 5. Insert transaction log record
	txID := uuid.New().String()
	txQuery := `INSERT INTO transactions (id, wallet_id, amount, type, description, created_by, created_at)
	            VALUES ($1, $2, $3, 'expense', $4, $5, NOW())`
	if _, err = tx.ExecContext(ctx, txQuery, txID, proposal.WalletID, proposal.Amount, proposal.Description, reviewerID); err != nil {
		return fmt.Errorf("ProposalService.ApproveProposal: failed to record transaction: %w", err)
	}

	// 6. Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("ProposalService.ApproveProposal: failed to commit transaction: %w", err)
	}

	log.Info().Str("proposal_id", proposalID).Str("reviewer_id", reviewerID).Msg("Proposal approved successfully")
	return nil
}

// RejectProposal handles the business logic for rejecting a proposal.
func (s *proposalService) RejectProposal(ctx context.Context, proposalID string, reviewerID string) error {
	err := s.proposalRepo.RejectProposal(ctx, s.db, proposalID, reviewerID)
	if err != nil {
		return fmt.Errorf("ProposalService.RejectProposal: %w", err)
	}

	return nil
}
