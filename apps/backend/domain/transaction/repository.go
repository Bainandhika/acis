package transaction

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
)

type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, exec DBExecutor, tx *Transaction) error
	GetTransactionsByFamilyID(ctx context.Context, familyID string) ([]Transaction, error)
	CreateProposal(ctx context.Context, prop *Proposal) error
	GetProposalsByFamilyID(ctx context.Context, familyID string) ([]Proposal, error)
	GetProposalForUpdate(ctx context.Context, exec DBExecutor, proposalID string) (*Proposal, error)
	GetWalletForUpdate(ctx context.Context, exec DBExecutor, walletID string) (*Wallet, error)
	UpdateWalletBalance(ctx context.Context, exec DBExecutor, walletID string, newBalance float64) error
	UpdateProposalStatus(ctx context.Context, exec DBExecutor, proposalID, status, reviewerID string) error
}

type txRepoImpl struct {
	db *database.AppDB
}

func NewRepository(db *database.AppDB) TransactionRepository {
	return &txRepoImpl{db: db}
}

func (r *txRepoImpl) CreateTransaction(ctx context.Context, exec DBExecutor, tx *Transaction) error {
	query := `INSERT INTO transactions (id, wallet_id, created_by, type, amount, category, description) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING created_at`
	return exec.QueryRowContext(ctx, query, tx.ID, tx.WalletID, tx.CreatedBy, tx.Type, tx.Amount, tx.Category, tx.Description).Scan(&tx.CreatedAt)
}

func (r *txRepoImpl) GetTransactionsByFamilyID(ctx context.Context, familyID string) ([]Transaction, error) {
	query := `SELECT t.id, t.wallet_id, t.created_by, t.type, t.amount, t.category, t.description, t.created_at
			  FROM transactions t
			  JOIN wallets w ON t.wallet_id = w.id
			  WHERE w.family_id = $1 ORDER BY t.created_at DESC`
	var list []Transaction
	err := r.db.SelectContext(ctx, &list, query, familyID)
	return list, err
}

func (r *txRepoImpl) CreateProposal(ctx context.Context, p *Proposal) error {
	query := `INSERT INTO proposals (id, wallet_id, proposed_by, title, amount, description, status) 
			  VALUES ($1, $2, $3, $4, $5, $6, 'pending') RETURNING created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, p.ID, p.WalletID, p.ProposedBy, p.Title, p.Amount, p.Description).Scan(&p.CreatedAt, &p.UpdatedAt)
}

func (r *txRepoImpl) GetProposalsByFamilyID(ctx context.Context, familyID string) ([]Proposal, error) {
	query := `SELECT p.id, p.wallet_id, p.proposed_by, p.title, p.amount, p.description, p.status, p.reviewed_by, p.reviewed_at, p.created_at, p.updated_at
			  FROM proposals p
			  JOIN wallets w ON p.wallet_id = w.id
			  WHERE w.family_id = $1 ORDER BY p.created_at DESC`
	var list []Proposal
	err := r.db.SelectContext(ctx, &list, query, familyID)
	return list, err
}

func (r *txRepoImpl) GetProposalForUpdate(ctx context.Context, exec DBExecutor, proposalID string) (*Proposal, error) {
	query := `SELECT id, wallet_id, proposed_by, title, amount, description, status, reviewed_by, reviewed_at, created_at, updated_at 
			  FROM proposals WHERE id = $1 FOR UPDATE`
	var p Proposal
	err := exec.QueryRowContext(ctx, query, proposalID).Scan(
		&p.ID, &p.WalletID, &p.ProposedBy, &p.Title, &p.Amount, &p.Description,
		&p.Status, &p.ReviewedBy, &p.ReviewedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &p, err
}

func (r *txRepoImpl) GetWalletForUpdate(ctx context.Context, exec DBExecutor, walletID string) (*Wallet, error) {
	query := `SELECT id, family_id, name, description, initial_balance, current_balance, minimum_limit, created_by, created_at, updated_at 
			  FROM wallets WHERE id = $1 FOR UPDATE`
	var w Wallet
	err := exec.QueryRowContext(ctx, query, walletID).Scan(
		&w.ID, &w.FamilyID, &w.Name, &w.Description, &w.InitialBalance, &w.CurrentBalance, &w.MinimumLimit, &w.CreatedBy, &w.CreatedAt, &w.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &w, err
}

func (r *txRepoImpl) UpdateWalletBalance(ctx context.Context, exec DBExecutor, walletID string, newBalance float64) error {
	query := `UPDATE wallets SET current_balance = $1, updated_at = NOW() WHERE id = $2`
	_, err := exec.ExecContext(ctx, query, newBalance, walletID)
	return err
}

func (r *txRepoImpl) UpdateProposalStatus(ctx context.Context, exec DBExecutor, proposalID, status, reviewerID string) error {
	query := `UPDATE proposals SET status = $1, reviewed_by = $2, reviewed_at = NOW(), updated_at = NOW() WHERE id = $3`
	_, err := exec.ExecContext(ctx, query, status, reviewerID, proposalID)
	return err
}
