package family

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
)

// DBExecutor abstracts *sql.DB and *sql.Tx so repo methods can run inside a transaction.
type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type FamilyRepository interface {
	CreateFamily(ctx context.Context, exec DBExecutor, family *Family) error
	FindByInviteCode(ctx context.Context, inviteCode string) (*Family, error)
	FindFamilyByUserID(ctx context.Context, userID string) (*Family, error)
	FindMemberByUserID(ctx context.Context, userID string) (*FamilyMember, error)
	AddMember(ctx context.Context, exec DBExecutor, member *FamilyMember) error
	GetMembers(ctx context.Context, familyID string) ([]FamilyMember, error)

	CreateWallet(ctx context.Context, wallet *Wallet) error
	GetWalletsByFamilyID(ctx context.Context, familyID string) ([]Wallet, error)
	GetWalletByID(ctx context.Context, walletID string) (*Wallet, error)
	GetLowBalanceWallets(ctx context.Context) ([]LowBalanceWalletDTO, error)
}

type familyRepoImpl struct {
	db *database.AppDB
}

func NewRepository(db *database.AppDB) FamilyRepository {
	return &familyRepoImpl{db: db}
}

func (r *familyRepoImpl) CreateFamily(ctx context.Context, exec DBExecutor, f *Family) error {
	query := `INSERT INTO families (id, name, invite_code, created_by) VALUES ($1, $2, $3, $4) RETURNING created_at, updated_at`
	return exec.QueryRowContext(ctx, query, f.ID, f.Name, f.InviteCode, f.CreatedBy).Scan(&f.CreatedAt, &f.UpdatedAt)
}

func (r *familyRepoImpl) FindByInviteCode(ctx context.Context, inviteCode string) (*Family, error) {
	query := `SELECT id, name, invite_code, created_by, created_at, updated_at FROM families WHERE invite_code = $1`
	var f Family
	err := r.db.GetContext(ctx, &f, query, inviteCode)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &f, err
}

func (r *familyRepoImpl) FindFamilyByUserID(ctx context.Context, userID string) (*Family, error) {
	query := `SELECT f.id, f.name, f.invite_code, f.created_by, f.created_at, f.updated_at 
			  FROM families f
			  JOIN family_members fm ON f.id = fm.family_id
			  WHERE fm.user_id = $1 LIMIT 1`
	var f Family
	err := r.db.GetContext(ctx, &f, query, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &f, err
}

func (r *familyRepoImpl) FindMemberByUserID(ctx context.Context, userID string) (*FamilyMember, error) {
	query := `SELECT id, family_id, user_id, role, joined_at FROM family_members WHERE user_id = $1 LIMIT 1`
	var m FamilyMember
	err := r.db.GetContext(ctx, &m, query, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &m, err
}

func (r *familyRepoImpl) AddMember(ctx context.Context, exec DBExecutor, m *FamilyMember) error {
	query := `INSERT INTO family_members (id, family_id, user_id, role) VALUES ($1, $2, $3, $4) RETURNING joined_at`
	return exec.QueryRowContext(ctx, query, m.ID, m.FamilyID, m.UserID, m.Role).Scan(&m.JoinedAt)
}

func (r *familyRepoImpl) GetMembers(ctx context.Context, familyID string) ([]FamilyMember, error) {
	query := `SELECT id, family_id, user_id, role, joined_at FROM family_members WHERE family_id = $1`
	var members []FamilyMember
	err := r.db.SelectContext(ctx, &members, query, familyID)
	return members, err
}

func (r *familyRepoImpl) CreateWallet(ctx context.Context, w *Wallet) error {
	query := `INSERT INTO wallets (id, family_id, name, description, initial_balance, current_balance, minimum_limit, created_by) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, w.ID, w.FamilyID, w.Name, w.Description, w.InitialBalance, w.CurrentBalance, w.MinimumLimit, w.CreatedBy).
		Scan(&w.CreatedAt, &w.UpdatedAt)
}

func (r *familyRepoImpl) GetWalletsByFamilyID(ctx context.Context, familyID string) ([]Wallet, error) {
	query := `SELECT id, family_id, name, description, initial_balance, current_balance, minimum_limit, created_at, updated_at 
			  FROM wallets WHERE family_id = $1 ORDER BY name ASC`
	var wallets []Wallet
	err := r.db.SelectContext(ctx, &wallets, query, familyID)
	return wallets, err
}

func (r *familyRepoImpl) GetWalletByID(ctx context.Context, walletID string) (*Wallet, error) {
	query := `SELECT id, family_id, name, description, initial_balance, current_balance, minimum_limit, created_at, updated_at 
			  FROM wallets WHERE id = $1`
	var w Wallet
	err := r.db.GetContext(ctx, &w, query, walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &w, err
}

func (r *familyRepoImpl) GetLowBalanceWallets(ctx context.Context) ([]LowBalanceWalletDTO, error) {
	query := `SELECT id AS wallet_id, name AS wallet_name, family_id, current_balance, minimum_limit 
			  FROM wallets WHERE current_balance <= minimum_limit`
	var list []LowBalanceWalletDTO
	err := r.db.SelectContext(ctx, &list, query)
	return list, err
}
