package transaction

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/jmoiron/sqlx"
)

type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, exec DBExecutor, tx *Transaction) error
	GetTransactionsByFamilyID(ctx context.Context, familyID string) ([]Transaction, error)
	GetTransactionsByFamilyIDAndPeriod(ctx context.Context, familyID string, year int, month int) ([]Transaction, error)
	GetTransactionByID(ctx context.Context, txID string) (*Transaction, error)
	UpdateTransactionRecord(ctx context.Context, exec DBExecutor, txID string, txType string, amount float64, description *string) error
	DeleteTransaction(ctx context.Context, exec DBExecutor, txID string) error
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
	query := `INSERT INTO transactions (id, wallet_id, family_id, created_by, type, amount, description) 
			  VALUES ($1, NULLIF($2, '')::uuid, $3, $4, $5, $6, $7) RETURNING created_at`
	return exec.QueryRowContext(ctx, query, tx.ID, tx.WalletID, tx.FamilyID, tx.CreatedBy, tx.Type, tx.Amount, tx.Description).Scan(&tx.CreatedAt)
}

func (r *txRepoImpl) GetTransactionsByFamilyID(ctx context.Context, familyID string) ([]Transaction, error) {
	return r.GetTransactionsByFamilyIDAndPeriod(ctx, familyID, 0, 0)
}

func (r *txRepoImpl) GetTransactionsByFamilyIDAndPeriod(ctx context.Context, familyID string, year int, month int) ([]Transaction, error) {
	var query string
	var args []interface{}

	if year > 0 && month > 0 {
		query = `SELECT t.id, COALESCE(t.wallet_id::text, '') AS wallet_id, t.family_id, t.created_by, t.type, t.amount, t.description, t.created_at
				 FROM transactions t
				 LEFT JOIN wallets w ON t.wallet_id = w.id
				 WHERE COALESCE(t.family_id, w.family_id) = $1 
				   AND EXTRACT(YEAR FROM t.created_at) = $2
				   AND EXTRACT(MONTH FROM t.created_at) = $3
				 ORDER BY t.created_at DESC`
		args = []interface{}{familyID, year, month}
	} else {
		query = `SELECT t.id, COALESCE(t.wallet_id::text, '') AS wallet_id, t.family_id, t.created_by, t.type, t.amount, t.description, t.created_at
				 FROM transactions t
				 LEFT JOIN wallets w ON t.wallet_id = w.id
				 WHERE COALESCE(t.family_id, w.family_id) = $1 
				 ORDER BY t.created_at DESC`
		args = []interface{}{familyID}
	}

	var list []Transaction
	err := r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		return tx.SelectContext(ctx, &list, query, args...)
	})
	return list, err
}

func (r *txRepoImpl) GetTransactionByID(ctx context.Context, txID string) (*Transaction, error) {
	var txRec Transaction
	err := r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `SELECT id, COALESCE(wallet_id::text, '') AS wallet_id, family_id, created_by, type, amount, description, created_at 
				  FROM transactions WHERE id = $1`
		err := tx.GetContext(ctx, &txRec, query, txID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	if txRec.ID == "" {
		return nil, nil
	}
	return &txRec, nil
}

func (r *txRepoImpl) UpdateTransactionRecord(ctx context.Context, exec DBExecutor, txID string, txType string, amount float64, description *string) error {
	query := `UPDATE transactions SET type = $1, amount = $2, description = $3 WHERE id = $4`
	_, err := exec.ExecContext(ctx, query, txType, amount, description, txID)
	return err
}

func (r *txRepoImpl) DeleteTransaction(ctx context.Context, exec DBExecutor, txID string) error {
	query := `DELETE FROM transactions WHERE id = $1`
	_, err := exec.ExecContext(ctx, query, txID)
	return err
}

func (r *txRepoImpl) CreateProposal(ctx context.Context, p *Proposal) error {
	return r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `INSERT INTO proposals (id, wallet_id, proposed_by, title, amount, description, status, request_type, target_transaction_id, payload) 
				  VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7, $8, $9) RETURNING created_at, updated_at`
		return tx.QueryRowContext(ctx, query, p.ID, p.WalletID, p.ProposedBy, p.Title, p.Amount, p.Description, p.RequestType, p.TargetTransactionID, p.Payload).Scan(&p.CreatedAt, &p.UpdatedAt)
	})
}

func (r *txRepoImpl) GetProposalsByFamilyID(ctx context.Context, familyID string) ([]Proposal, error) {
	var list []Proposal
	err := r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `SELECT p.id, p.wallet_id, p.proposed_by, p.title, p.amount, p.description, p.status, p.request_type, p.target_transaction_id, p.payload, p.reviewed_by, p.reviewed_at, p.created_at, p.updated_at
				  FROM proposals p
				  JOIN wallets w ON p.wallet_id = w.id
				  WHERE w.family_id = $1 ORDER BY p.created_at DESC`
		return tx.SelectContext(ctx, &list, query, familyID)
	})
	return list, err
}

func (r *txRepoImpl) GetProposalForUpdate(ctx context.Context, exec DBExecutor, proposalID string) (*Proposal, error) {
	query := `SELECT id, wallet_id, proposed_by, title, amount, description, status, request_type, target_transaction_id, payload, reviewed_by, reviewed_at, created_at, updated_at 
			  FROM proposals WHERE id = $1 FOR UPDATE`
	var p Proposal
	err := exec.QueryRowContext(ctx, query, proposalID).Scan(
		&p.ID, &p.WalletID, &p.ProposedBy, &p.Title, &p.Amount, &p.Description,
		&p.Status, &p.RequestType, &p.TargetTransactionID, &p.Payload, &p.ReviewedBy, &p.ReviewedAt, &p.CreatedAt, &p.UpdatedAt,
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
