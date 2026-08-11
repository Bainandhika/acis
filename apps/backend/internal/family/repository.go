package family

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Bainandhika/acis/apps/backend/internal/database"
	"github.com/Bainandhika/acis/apps/backend/internal/domain"
)

type Repository interface {
	CreateFamily(ctx context.Context, family *domain.Family) error
	FindByInviteCode(ctx context.Context, inviteCode string) (*domain.Family, error)
	FindFamilyByUserID(ctx context.Context, userID string) (*domain.Family, error)
	AddMember(ctx context.Context, member *domain.FamilyMember) error
	GetMembers(ctx context.Context, familyID string) ([]domain.FamilyMember, error)
	
	CreateWallet(ctx context.Context, wallet *domain.Wallet) error
	GetWalletsByFamilyID(ctx context.Context, familyID string) ([]domain.Wallet, error)
	GetWalletByID(ctx context.Context, walletID string) (*domain.Wallet, error)
	GetLowBalanceWallets(ctx context.Context) ([]LowBalanceWalletDTO, error)
}

type repository struct {
	db *database.AppDB
}

func NewRepository(db *database.AppDB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateFamily(ctx context.Context, f *domain.Family) error {
	query := `INSERT INTO families (id, name, invite_code, created_by) VALUES ($1, $2, $3, $4) RETURNING created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, f.ID, f.Name, f.InviteCode, f.CreatedBy).Scan(&f.CreatedAt, &f.UpdatedAt)
}

func (r *repository) FindByInviteCode(ctx context.Context, inviteCode string) (*domain.Family, error) {
	query := `SELECT id, name, invite_code, created_by, created_at, updated_at FROM families WHERE invite_code = $1`
	var f domain.Family
	err := r.db.GetContext(ctx, &f, query, inviteCode)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &f, err
}

func (r *repository) FindFamilyByUserID(ctx context.Context, userID string) (*domain.Family, error) {
	query := `SELECT f.id, f.name, f.invite_code, f.created_by, f.created_at, f.updated_at 
			  FROM families f
			  JOIN family_members fm ON f.id = fm.family_id
			  WHERE fm.user_id = $1 LIMIT 1`
	var f domain.Family
	err := r.db.GetContext(ctx, &f, query, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &f, err
}

func (r *repository) AddMember(ctx context.Context, m *domain.FamilyMember) error {
	query := `INSERT INTO family_members (id, family_id, user_id, role) VALUES ($1, $2, $3, $4) RETURNING joined_at`
	return r.db.QueryRowContext(ctx, query, m.ID, m.FamilyID, m.UserID, m.Role).Scan(&m.JoinedAt)
}

func (r *repository) GetMembers(ctx context.Context, familyID string) ([]domain.FamilyMember, error) {
	query := `SELECT id, family_id, user_id, role, joined_at FROM family_members WHERE family_id = $1`
	var members []domain.FamilyMember
	err := r.db.SelectContext(ctx, &members, query, familyID)
	return members, err
}

func (r *repository) CreateWallet(ctx context.Context, w *domain.Wallet) error {
	query := `INSERT INTO wallets (id, family_id, name, description, initial_balance, current_balance, minimum_limit) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, w.ID, w.FamilyID, w.Name, w.Description, w.InitialBalance, w.CurrentBalance, w.MinimumLimit).
		Scan(&w.CreatedAt, &w.UpdatedAt)
}

func (r *repository) GetWalletsByFamilyID(ctx context.Context, familyID string) ([]domain.Wallet, error) {
	query := `SELECT id, family_id, name, description, initial_balance, current_balance, minimum_limit, created_at, updated_at 
			  FROM wallets WHERE family_id = $1 ORDER BY name ASC`
	var wallets []domain.Wallet
	err := r.db.SelectContext(ctx, &wallets, query, familyID)
	return wallets, err
}

func (r *repository) GetWalletByID(ctx context.Context, walletID string) (*domain.Wallet, error) {
	query := `SELECT id, family_id, name, description, initial_balance, current_balance, minimum_limit, created_at, updated_at 
			  FROM wallets WHERE id = $1`
	var w domain.Wallet
	err := r.db.GetContext(ctx, &w, query, walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &w, err
}

func (r *repository) GetLowBalanceWallets(ctx context.Context) ([]LowBalanceWalletDTO, error) {
	query := `SELECT id AS wallet_id, name AS wallet_name, family_id, current_balance, minimum_limit 
			  FROM wallets WHERE current_balance <= minimum_limit`
	var list []LowBalanceWalletDTO
	err := r.db.SelectContext(ctx, &list, query)
	return list, err
}
