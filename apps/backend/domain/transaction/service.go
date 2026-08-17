package transaction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/notification"
	"github.com/google/uuid"
)

type TransactionService interface {
	CreateDirectTransaction(ctx context.Context, req CreateTransactionDTO) (*TransactionDTO, error)
	UpdateTransaction(ctx context.Context, txID string, req UpdateTransactionDTO) (*TransactionDTO, error)
	DeleteTransaction(ctx context.Context, txID string, familyID string) error
	GetTransactions(ctx context.Context, familyID string) ([]TransactionDTO, error)
	CreateProposal(ctx context.Context, req CreateProposalDTO) (*ProposalDTO, error)
	GetProposals(ctx context.Context, familyID string) ([]ProposalDTO, error)
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
		Description: record.Description,
		CreatedAt:   record.CreatedAt,
	}, nil
}

func (s *transactionService) UpdateTransaction(ctx context.Context, txID string, req UpdateTransactionDTO) (*TransactionDTO, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.New("failed to start database transaction")
	}
	defer tx.Rollback()

	record, err := s.repo.GetTransactionByID(ctx, txID)
	if err != nil || record == nil {
		return nil, errors.New("transaction not found")
	}

	wallet, err := s.repo.GetWalletForUpdate(ctx, tx, record.WalletID)
	if err != nil || wallet == nil {
		return nil, errors.New("associated wallet not found")
	}
	if wallet.FamilyID != req.FamilyID {
		return nil, errors.New("unauthorized transaction update")
	}

	// 1. Reverse old transaction effect
	var restoredBalance float64
	if record.Type == "expense" {
		restoredBalance = wallet.CurrentBalance + record.Amount
	} else if record.Type == "income" {
		restoredBalance = wallet.CurrentBalance - record.Amount
	} else {
		restoredBalance = wallet.CurrentBalance
	}

	// 2. Apply new transaction effect
	var newBalance float64
	if req.Type == "expense" {
		if restoredBalance < req.Amount {
			return nil, errors.New("insufficient wallet balance for update")
		}
		newBalance = restoredBalance - req.Amount
	} else if req.Type == "income" {
		newBalance = restoredBalance + req.Amount
	} else {
		return nil, errors.New("invalid transaction type")
	}

	if err := s.repo.UpdateWalletBalance(ctx, tx, wallet.ID, newBalance); err != nil {
		return nil, errors.New("failed to update wallet balance")
	}

	if err := s.repo.UpdateTransactionRecord(ctx, tx, txID, req.Type, req.Amount, req.Description); err != nil {
		return nil, errors.New("failed to update transaction record")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.New("failed to commit transaction update")
	}

	return &TransactionDTO{
		ID:          txID,
		WalletID:    record.WalletID,
		UserID:      record.CreatedBy,
		Type:        req.Type,
		Amount:      req.Amount,
		Description: req.Description,
		CreatedAt:   record.CreatedAt,
	}, nil
}

func (s *transactionService) DeleteTransaction(ctx context.Context, txID string, familyID string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.New("failed to start database transaction")
	}
	defer tx.Rollback()

	record, err := s.repo.GetTransactionByID(ctx, txID)
	if err != nil || record == nil {
		return errors.New("transaction not found")
	}

	wallet, err := s.repo.GetWalletForUpdate(ctx, tx, record.WalletID)
	if err != nil || wallet == nil {
		return errors.New("associated wallet not found")
	}
	if wallet.FamilyID != familyID {
		return errors.New("unauthorized transaction delete")
	}

	// Reverse the transaction's financial effect on the wallet
	var newBalance float64
	if record.Type == "expense" {
		newBalance = wallet.CurrentBalance + record.Amount
	} else if record.Type == "income" {
		newBalance = wallet.CurrentBalance - record.Amount
	} else {
		newBalance = wallet.CurrentBalance
	}

	if err := s.repo.UpdateWalletBalance(ctx, tx, wallet.ID, newBalance); err != nil {
		return errors.New("failed to restore wallet balance")
	}

	if err := s.repo.DeleteTransaction(ctx, tx, txID); err != nil {
		return errors.New("failed to delete transaction")
	}

	return tx.Commit()
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

func (s *transactionService) GetProposals(ctx context.Context, familyID string) ([]ProposalDTO, error) {
	records, err := s.repo.GetProposalsByFamilyID(ctx, familyID)
	if err != nil {
		return nil, errors.New("failed to fetch proposals")
	}

	var dtos []ProposalDTO
	for _, r := range records {
		title := ""
		if r.Title != nil {
			title = *r.Title
		}
		var payload json.RawMessage
		if r.Payload != nil && *r.Payload != "" {
			payload = json.RawMessage(*r.Payload)
		}
		dtos = append(dtos, ProposalDTO{
			ID:                  r.ID,
			WalletID:            r.WalletID,
			ProposedBy:          r.ProposedBy,
			Title:               title,
			Amount:              r.Amount,
			Description:         r.Description,
			Status:              r.Status,
			RequestType:         r.RequestType,
			TargetTransactionID: r.TargetTransactionID,
			Payload:             payload,
			ReviewedBy:          r.ReviewedBy,
			ReviewedAt:          r.ReviewedAt,
			CreatedAt:           r.CreatedAt,
		})
	}
	return dtos, nil
}

func (s *transactionService) CreateProposal(ctx context.Context, req CreateProposalDTO) (*ProposalDTO, error) {
	var payloadStr *string
	if len(req.Payload) > 0 {
		str := string(req.Payload)
		payloadStr = &str
	}

	reqType := req.RequestType
	if reqType == "" {
		reqType = "add_transaction"
	}

	prop := &Proposal{
		ID:                  uuid.NewString(),
		WalletID:            req.WalletID,
		ProposedBy:          &req.ProposedBy,
		Title:               &req.Title,
		Amount:              req.Amount,
		Description:         req.Description,
		Status:              "pending",
		RequestType:         reqType,
		TargetTransactionID: req.TargetTransactionID,
		Payload:             payloadStr,
	}

	if err := s.repo.CreateProposal(ctx, prop); err != nil {
		return nil, errors.New("failed to create proposal")
	}

	return &ProposalDTO{
		ID:                  prop.ID,
		WalletID:            prop.WalletID,
		ProposedBy:          prop.ProposedBy,
		Title:               req.Title,
		Amount:              prop.Amount,
		Description:         prop.Description,
		Status:              prop.Status,
		RequestType:         prop.RequestType,
		TargetTransactionID: prop.TargetTransactionID,
		Payload:             req.Payload,
		CreatedAt:           prop.CreatedAt,
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

	switch proposal.RequestType {
	case "delete_transaction":
		if proposal.TargetTransactionID == nil {
			return errors.New("target transaction id is required for delete request")
		}
		targetTx, err := s.repo.GetTransactionByID(ctx, *proposal.TargetTransactionID)
		if err != nil || targetTx == nil {
			return errors.New("target transaction not found")
		}
		wallet, err := s.repo.GetWalletForUpdate(ctx, tx, targetTx.WalletID)
		if err != nil || wallet == nil {
			return errors.New("wallet not found")
		}
		var newBal float64
		if targetTx.Type == "expense" {
			newBal = wallet.CurrentBalance + targetTx.Amount
		} else if targetTx.Type == "income" {
			newBal = wallet.CurrentBalance - targetTx.Amount
		} else {
			newBal = wallet.CurrentBalance
		}
		if err := s.repo.UpdateWalletBalance(ctx, tx, wallet.ID, newBal); err != nil {
			return errors.New("failed to restore wallet balance")
		}
		if err := s.repo.DeleteTransaction(ctx, tx, targetTx.ID); err != nil {
			return errors.New("failed to delete target transaction")
		}

	case "edit_transaction":
		if proposal.TargetTransactionID == nil {
			return errors.New("target transaction id is required for edit request")
		}
		targetTx, err := s.repo.GetTransactionByID(ctx, *proposal.TargetTransactionID)
		if err != nil || targetTx == nil {
			return errors.New("target transaction not found")
		}
		wallet, err := s.repo.GetWalletForUpdate(ctx, tx, targetTx.WalletID)
		if err != nil || wallet == nil {
			return errors.New("wallet not found")
		}

		var editData struct {
			Type        string  `json:"type"`
			Amount      float64 `json:"amount"`
			Description string  `json:"description"`
		}
		if proposal.Payload != nil && *proposal.Payload != "" {
			_ = json.Unmarshal([]byte(*proposal.Payload), &editData)
		}
		if editData.Type == "" {
			editData.Type = targetTx.Type
		}
		if editData.Amount <= 0 {
			editData.Amount = proposal.Amount
		}
		if editData.Description == "" {
			editData.Description = proposal.Description
		}

		var restoredBal float64
		if targetTx.Type == "expense" {
			restoredBal = wallet.CurrentBalance + targetTx.Amount
		} else if targetTx.Type == "income" {
			restoredBal = wallet.CurrentBalance - targetTx.Amount
		} else {
			restoredBal = wallet.CurrentBalance
		}

		var newBal float64
		if editData.Type == "expense" {
			if restoredBal < editData.Amount {
				return errors.New("insufficient wallet balance for approved edit")
			}
			newBal = restoredBal - editData.Amount
		} else if editData.Type == "income" {
			newBal = restoredBal + editData.Amount
		}

		if err := s.repo.UpdateWalletBalance(ctx, tx, wallet.ID, newBal); err != nil {
			return errors.New("failed to update wallet balance")
		}
		if err := s.repo.UpdateTransactionRecord(ctx, tx, targetTx.ID, editData.Type, editData.Amount, &editData.Description); err != nil {
			return errors.New("failed to update transaction record")
		}

	default: // "add_transaction"
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
	}

	if err := s.repo.UpdateProposalStatus(ctx, tx, proposalID, "approved", reviewerID); err != nil {
		return errors.New("failed to update proposal status")
	}

	if s.outboxRepo != nil && proposal.ProposedBy != nil {
		payload := map[string]interface{}{
			"proposal_id":  proposalID,
			"request_type": proposal.RequestType,
			"description":  proposal.Description,
			"amount":       proposal.Amount,
			"reviewer_id":  reviewerID,
		}
		if err := s.outboxRepo.EnqueueTx(ctx, tx, "proposal_approved", *proposal.ProposedBy, payload); err != nil {
			slog.Warn("Failed to enqueue proposal approval notification", slog.Any("error", err))
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.New("failed to commit proposal approval")
	}

	if s.outboxRepo != nil {
		_ = s.outboxRepo.PublishSignal(ctx)
	}

	slog.Info("Proposal approved atomically", slog.String("proposal_id", proposalID), slog.String("reviewer_id", reviewerID))
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
