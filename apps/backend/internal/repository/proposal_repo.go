package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bainandhika/acis/apps/backend/internal/domain"
)

type ProposalRepository interface {
	Create(ctx context.Context, executor DBExecutor, proposal *domain.Proposal) error
	RejectProposal(ctx context.Context, exec DBExecutor, proposalID string, reviewerID string) error
}

type proposalRepo struct{}

func NewProposalRepository() ProposalRepository {
	return &proposalRepo{}
}

func (r *proposalRepo) Create(ctx context.Context, executor DBExecutor, proposal *domain.Proposal) error {
	query := `INSERT INTO proposals (id, wallet_id, amount, description, status, proposed_by, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING id, created_at, updated_at`

	err := executor.QueryRowContext(ctx, query,
		proposal.ID, proposal.WalletID, proposal.Amount, proposal.Description,
		proposal.Status, proposal.ProposedBy).
		Scan(&proposal.ID, &proposal.CreatedAt, &proposal.UpdatedAt)

	return err
}

func (r *proposalRepo) RejectProposal(ctx context.Context, exec DBExecutor, proposalID string, reviewerID string) error {
	query := `
		UPDATE proposals 
		SET 
			status = 'rejected', 
			reviewed_by = $1, 
			reviewed_at = NOW(), 
			updated_at = NOW()
		WHERE id = $2 AND status = 'pending'
	`

	result, err := exec.ExecContext(ctx, query, reviewerID, proposalID)
	if err != nil {
		return fmt.Errorf("ProposalRepository.RejectProposal: failed to execute query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ProposalRepository.RejectProposal: failed to get rows affected: %w", err)
	}

	// If rowsAffected == 0, it means the proposal doesn't exist OR it's no longer 'pending'
	if rowsAffected == 0 {
		return errors.New("proposal not found or already processed")
	}

	return nil
}
