package repository

import (
	"context"

	"github.com/Bainandhika/acis/apps/backend/internal/database"
	"github.com/Bainandhika/acis/apps/backend/internal/domain"
)

// WalletRepository defines the contract for wallet data operations
type WalletRepository interface {
	Create(ctx context.Context, wallet *domain.Wallet) error
	GetByFamilyID(ctx context.Context, familyID string) ([]domain.Wallet, error)
	UpdateBalance(ctx context.Context, walletID string, amount float64) error
}

type walletRepo struct {
	db *database.AppDB
}

func NewWalletRepository(db *database.AppDB) WalletRepository {
	return &walletRepo{db: db}
}

func (r *walletRepo) Create(ctx context.Context, wallet *domain.Wallet) error {
	query := `INSERT INTO wallets (id, family_id, name, description, initial_balance, current_balance, minimum_limit, created_by) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`
	
	err := r.db.QueryRowContext(ctx, query, 
		wallet.ID, wallet.FamilyID, wallet.Name, wallet.Description, 
		wallet.InitialBalance, wallet.CurrentBalance, wallet.MinimumLimit, wallet.CreatedBy).
		Scan(&wallet.ID, &wallet.CreatedAt, &wallet.UpdatedAt)
		
	return err
}

func (r *walletRepo) GetByFamilyID(ctx context.Context, familyID string) ([]domain.Wallet, error) {
	query := `SELECT id, family_id, name, description, initial_balance, current_balance, minimum_limit, created_by, created_at, updated_at 
			  FROM wallets WHERE family_id = $1 ORDER BY created_at DESC`
	
	var wallets []domain.Wallet
	err := r.db.SelectContext(ctx, &wallets, query, familyID)
	
	if err != nil {
		return nil, err
	}
	return wallets, nil
}

// UpdateBalance updates the current_balance of a wallet
func (r *walletRepo) UpdateBalance(ctx context.Context, walletID string, amount float64) error {
	query := `UPDATE wallets SET current_balance = current_balance + $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, amount, walletID)
	return err
}