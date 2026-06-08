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
	// If the function returns an error, or if there's a panic, the transaction will be rolled back.
	// If tx.Commit() is called successfully, it clears the error state, so Rollback() becomes a no-op.
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 3. Business Logic: Check Wallet Balance (using tx as executor)
	wallet, err := s.walletRepo.GetByID(ctx, tx, req.WalletID)
	if err != nil {
		log.Error().Err(err).Str("trace_id", ctx.Value("X-Transaction-ID").(string)).Msg("Failed to get wallet")
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
		log.Error().Err(err).Str("trace_id", ctx.Value("X-Transaction-ID").(string)).Msg("Failed to create proposal")
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

// RejectProposal handles the business logic for rejecting a proposal.
func (s *proposalService) RejectProposal(ctx context.Context, proposalID string, reviewerID string) error {
	// Since this is a single UPDATE query, it is inherently atomic in PostgreSQL.
	// We don't need to explicitly BEGIN/COMMIT a transaction like in ApproveProposal.
	// We just pass the main DB connection (which implements DBExecutor).
	err := s.proposalRepo.RejectProposal(ctx, s.db, proposalID, reviewerID)
	if err != nil {
		return fmt.Errorf("ProposalService.RejectProposal: %w", err)
	}

	// TODO: Trigger notification (e.g., email/Telegram bot) to the proposer later.

	return nil
}
