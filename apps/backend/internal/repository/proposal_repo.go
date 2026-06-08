package repository

import (
	"context"

	"github.com/Bainandhika/acis/apps/backend/internal/domain"
)

type ProposalRepository interface {
	Create(ctx context.Context, executor DBExecutor, proposal *domain.Proposal) error
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
