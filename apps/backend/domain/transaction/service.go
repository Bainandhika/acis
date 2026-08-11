package transaction

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/notification"
	"github.com/google/uuid"
	zerolog "github.com/rs/zerolog/log"
)

type TransactionService interface {
	CreateDirectTransaction(ctx context.Context, req CreateTransactionDTO) (*TransactionDTO, error)
	GetTransactions(ctx context.Context, familyID string) ([]TransactionDTO, error)
	CreateProposal(ctx context.Context, req CreateProposalDTO) (*ProposalDTO, error)
	ApproveProposal(ctx context.Context, proposalID string, reviewerID string) error
	RejectProposal(ctx context.Context, proposalID string, reviewerID string) error
}

type transactionService struct {
	repo       TransactionRepository
	outboxRepo notification.OutboxRepository
	db         *database.AppDB
}

func NewService(repo TransactionRepository, outboxRepo notification.OutboxRepository, db *database.AppDB) TransactionService {
	return &transactionService{
		repo:       repo,
		outboxRepo: outboxRepo,
		db:         db,
	}
}

func (s *transactionService) CreateDirectTransaction(ctx context.Context, req CreateTransactionDTO) (*TransactionDTO, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.New("failed to start database transaction")
	}
	defer tx.Rollback()

	wallet, err := s.repo.GetWalletForUpdate(ctx, tx, req.WalletID)
	if err != nil || wallet == nil {
		return nil, errors.New("wallet not found")
	}

	var newBalance float64
	if req.Type == "expense" {
		if wallet.CurrentBalance < req.Amount {
			return nil, errors.New("insufficient wallet balance")
		}
		newBalance = wallet.CurrentBalance - req.Amount
	} else if req.Type == "income" {
		newBalance = wallet.CurrentBalance + req.Amount
	} else {
		return nil, errors.New("invalid transaction type")
	}

	if err := s.repo.UpdateWalletBalance(ctx, tx, wallet.ID, newBalance); err != nil {
		return nil, errors.New("failed to update wallet balance")
	}

	record := &Transaction{
		ID:          uuid.NewString(),
		WalletID:    req.WalletID,
		CreatedBy:   &req.UserID,
		Type:        req.Type,
		Amount:      req.Amount,
		Description: req.Description,
	}

	if err := s.repo.CreateTransaction(ctx, tx, record); err != nil {
		return nil, errors.New("failed to record transaction")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.New("failed to commit transaction")
	}

	return &TransactionDTO{
		ID:          record.ID,
		WalletID:    record.WalletID,
		UserID:      record.CreatedBy,
		Type:        record.Type,
		Amount:      record.Amount,
		Category:    req.Category,
		Description: record.Description,
		CreatedAt:   record.CreatedAt,
	}, nil
}

func (s *transactionService) GetTransactions(ctx context.Context, familyID string) ([]TransactionDTO, error) {
	records, err := s.repo.GetTransactionsByFamilyID(ctx, familyID)
	if err != nil {
		return nil, errors.New("failed to fetch transactions")
	}

	var dtos []TransactionDTO
	for _, r := range records {
		dtos = append(dtos, TransactionDTO{
			ID:          r.ID,
			WalletID:    r.WalletID,
			UserID:      r.CreatedBy,
			Type:        r.Type,
			Amount:      r.Amount,
			Description: r.Description,
			CreatedAt:   r.CreatedAt,
		})
	}
	return dtos, nil
}

func (s *transactionService) CreateProposal(ctx context.Context, req CreateProposalDTO) (*ProposalDTO, error) {
	prop := &Proposal{
		ID:          uuid.NewString(),
		WalletID:    req.WalletID,
		ProposedBy:  &req.ProposedBy,
		Amount:      req.Amount,
		Description: req.Description,
		Status:      "pending",
	}

	if err := s.repo.CreateProposal(ctx, prop); err != nil {
		return nil, errors.New("failed to create proposal")
	}

	return &ProposalDTO{
		ID:          prop.ID,
		WalletID:    prop.WalletID,
		ProposedBy:  prop.ProposedBy,
		Title:       req.Title,
		Amount:      prop.Amount,
		Description: prop.Description,
		Status:      prop.Status,
		CreatedAt:   prop.CreatedAt,
	}, nil
}

func (s *transactionService) ApproveProposal(ctx context.Context, proposalID string, reviewerID string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.New("failed to start database transaction")
	}
	defer tx.Rollback()

	proposal, err := s.repo.GetProposalForUpdate(ctx, tx, proposalID)
	if err != nil || proposal == nil {
		return errors.New("proposal not found")
	}

	if proposal.Status != "pending" {
		return fmt.Errorf("proposal already processed with status: %s", proposal.Status)
	}

	wallet, err := s.repo.GetWalletForUpdate(ctx, tx, proposal.WalletID)
	if err != nil || wallet == nil {
		return errors.New("wallet not found")
	}

	if wallet.CurrentBalance < proposal.Amount {
		return errors.New("insufficient wallet balance")
	}

	newBalance := wallet.CurrentBalance - proposal.Amount
	if err := s.repo.UpdateWalletBalance(ctx, tx, wallet.ID, newBalance); err != nil {
		return errors.New("failed to update wallet balance")
	}

	if err := s.repo.UpdateProposalStatus(ctx, tx, proposalID, "approved", reviewerID); err != nil {
		return errors.New("failed to update proposal status")
	}

	record := &Transaction{
		ID:          uuid.NewString(),
		WalletID:    proposal.WalletID,
		CreatedBy:   proposal.ProposedBy,
		Type:        "expense",
		Amount:      proposal.Amount,
		Description: &proposal.Description,
	}
	if err := s.repo.CreateTransaction(ctx, tx, record); err != nil {
		return errors.New("failed to record expense transaction")
	}

	if s.outboxRepo != nil && proposal.ProposedBy != nil {
		payload := map[string]interface{}{
			"proposal_id": proposalID,
			"description": proposal.Description,
			"amount":      proposal.Amount,
			"reviewer_id": reviewerID,
		}
		if err := s.outboxRepo.EnqueueTx(ctx, tx, "proposal_approved", *proposal.ProposedBy, payload); err != nil {
			zerolog.Warn().Err(err).Msg("Failed to enqueue proposal approval notification")
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.New("failed to commit proposal approval")
	}

	zerolog.Info().Str("proposal_id", proposalID).Str("reviewer_id", reviewerID).Msg("Proposal approved atomically")
	return nil
}

func (s *transactionService) RejectProposal(ctx context.Context, proposalID string, reviewerID string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.New("failed to start database transaction")
	}
	defer tx.Rollback()

	proposal, err := s.repo.GetProposalForUpdate(ctx, tx, proposalID)
	if err != nil || proposal == nil {
		return errors.New("proposal not found")
	}

	if proposal.Status != "pending" {
		return fmt.Errorf("proposal already processed with status: %s", proposal.Status)
	}

	if err := s.repo.UpdateProposalStatus(ctx, tx, proposalID, "rejected", reviewerID); err != nil {
		return errors.New("failed to reject proposal")
	}

	return tx.Commit()
}
