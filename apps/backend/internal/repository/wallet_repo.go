package repository

import (
	"context"

	"github.com/Bainandhika/acis/apps/backend/internal/domain"
)

// WalletRepository defines the contract for wallet data operations
type WalletRepository interface {
	Create(ctx context.Context, executor DBExecutor, wallet *domain.Wallet) error
	GetByID(ctx context.Context, executor DBExecutor, id string) (*domain.Wallet, error)
	GetByFamilyID(ctx context.Context, executor DBExecutor, familyID string) ([]domain.Wallet, error)
	UpdateBalance(ctx context.Context, executor DBExecutor, walletID string, amount float64) error
}

type walletRepo struct{}

func NewWalletRepository() WalletRepository {
	return &walletRepo{}
}

func (r *walletRepo) Create(ctx context.Context, executor DBExecutor, wallet *domain.Wallet) error {
	query := `INSERT INTO wallets (id, family_id, name, description, initial_balance, current_balance, minimum_limit, created_by)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`

	err := executor.QueryRowContext(ctx, query,
		wallet.ID, wallet.FamilyID, wallet.Name, wallet.Description,
		wallet.InitialBalance, wallet.CurrentBalance, wallet.MinimumLimit, wallet.CreatedBy).
		Scan(&wallet.ID, &wallet.CreatedAt, &wallet.UpdatedAt)

	return err
}

func (r *walletRepo) GetByID(ctx context.Context, executor DBExecutor, id string) (*domain.Wallet, error) {
	query := `SELECT id, family_id, name, description, initial_balance, current_balance, minimum_limit, created_by, created_at, updated_at
	FROM wallets WHERE id = $1`

	var wallet domain.Wallet
	err := executor.QueryRowContext(ctx, query, id).Scan(
		&wallet.ID, &wallet.FamilyID, &wallet.Name, &wallet.Description,
		&wallet.InitialBalance, &wallet.CurrentBalance, &wallet.MinimumLimit,
		&wallet.CreatedBy, &wallet.CreatedAt, &wallet.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepo) GetByFamilyID(ctx context.Context, executor DBExecutor, familyID string) ([]domain.Wallet, error) {
	query := `SELECT id, family_id, name, description, initial_balance, current_balance, minimum_limit, created_by, created_at, updated_at
	FROM wallets WHERE family_id = $1 ORDER BY created_at DESC`

	var wallets []domain.Wallet
	err := executor.SelectContext(ctx, &wallets, query, familyID)
	if err != nil {
		return nil, err
	}
	return wallets, nil
}

func (r *walletRepo) UpdateBalance(ctx context.Context, executor DBExecutor, walletID string, amount float64) error {
	query := `UPDATE wallets SET current_balance = current_balance + $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := executor.ExecContext(ctx, query, amount, walletID)
	return err
}
