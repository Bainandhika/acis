package repository

import (
	"context"

	"github.com/Bainandhika/acis/apps/backend/internal/domain"
)

type TransactionRepository interface {
	Create(ctx context.Context, exec DBExecutor, tx *domain.Transaction) error
	GetByWalletID(ctx context.Context, exec DBExecutor, walletID string, limit, offset int) ([]domain.Transaction, error)
}

type transactionRepo struct{}

func NewTransactionRepository() TransactionRepository {
	return &transactionRepo{}
}

func (r *transactionRepo) Create(ctx context.Context, exec DBExecutor, tx *domain.Transaction) error {
	query := `INSERT INTO transactions (id, wallet_id, amount, type, description, created_by, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6, NOW()) RETURNING created_at`

	return exec.QueryRowContext(ctx, query,
		tx.ID, tx.WalletID, tx.Amount, tx.Type, tx.Description, tx.CreatedBy).
		Scan(&tx.CreatedAt)
}

func (r *transactionRepo) GetByWalletID(ctx context.Context, exec DBExecutor, walletID string, limit, offset int) ([]domain.Transaction, error) {
	query := `SELECT id, wallet_id, amount, type, description, created_by, created_at
	          FROM transactions WHERE wallet_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	var list []domain.Transaction
	err := exec.SelectContext(ctx, &list, query, walletID, limit, offset)
	return list, err
}
