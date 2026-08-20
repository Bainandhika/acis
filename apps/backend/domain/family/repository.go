package family

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/jmoiron/sqlx"
)

// DBExecutor abstracts *sql.DB and *sql.Tx so repo methods can run inside a transaction.
type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type FamilyRepository interface {
	CreateFamily(ctx context.Context, exec DBExecutor, family *Family) error
	FindFamilyByID(ctx context.Context, id string) (*Family, error)
	FindByInviteCode(ctx context.Context, inviteCode string) (*Family, error)
	FindFamilyByUserID(ctx context.Context, userID string) (*Family, error)
	FindMemberByUserID(ctx context.Context, userID string) (*FamilyMember, error)
	AddMember(ctx context.Context, exec DBExecutor, member *FamilyMember) error
	GetMembers(ctx context.Context, familyID string) ([]FamilyMember, error)
	UpdateTelegramChatID(ctx context.Context, familyID string, chatID *int64) error
	FindByTelegramChatID(ctx context.Context, chatID int64) (*Family, error)
	UpdateFamilyName(ctx context.Context, familyID string, name string) error
	IncrementWalletCounter(ctx context.Context, familyID string) (int, error)

	CreateWallet(ctx context.Context, wallet *Wallet) error
	GetWalletsByFamilyID(ctx context.Context, familyID string) ([]Wallet, error)
	GetWalletByID(ctx context.Context, walletID string) (*Wallet, error)
	FindWalletByShortID(ctx context.Context, shortID string) (*Wallet, error)
	UpdateWallet(ctx context.Context, walletID string, name string, description *string, currentBalance float64, minimumLimit float64) error
	DeleteWallet(ctx context.Context, walletID string, familyID string) error
	GetLowBalanceWallets(ctx context.Context) ([]LowBalanceWalletDTO, error)
	FindMemberByID(ctx context.Context, memberID string) (*FamilyMember, error)
	RemoveMember(ctx context.Context, memberID string, familyID string) error
}

type familyRepoImpl struct {
	db *database.AppDB
}

func NewRepository(db *database.AppDB) FamilyRepository {
	return &familyRepoImpl{db: db}
}

func (r *familyRepoImpl) CreateFamily(ctx context.Context, exec DBExecutor, f *Family) error {
	query := `INSERT INTO families (id, name, invite_code, wallet_counter, created_by) VALUES ($1, $2, $3, $4, $5) RETURNING created_at, updated_at`
	return exec.QueryRowContext(ctx, query, f.ID, f.Name, f.InviteCode, f.WalletCounter, f.CreatedBy).Scan(&f.CreatedAt, &f.UpdatedAt)
}

func (r *familyRepoImpl) FindFamilyByID(ctx context.Context, id string) (*Family, error) {
	var f Family
	err := r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `SELECT id, name, invite_code, telegram_chat_id, wallet_counter, created_by, created_at, updated_at FROM families WHERE id = $1`
		err := tx.GetContext(ctx, &f, query, id)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	if f.ID == "" {
		return nil, nil
	}
	return &f, nil
}

func (r *familyRepoImpl) FindByInviteCode(ctx context.Context, inviteCode string) (*Family, error) {
	var f Family
	query := `SELECT id, name, invite_code, telegram_chat_id, wallet_counter, created_by, created_at, updated_at FROM families WHERE invite_code = $1`
	err := r.db.AdminDB().GetContext(ctx, &f, query, inviteCode)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *familyRepoImpl) FindFamilyByUserID(ctx context.Context, userID string) (*Family, error) {
	var f Family
	err := r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `SELECT f.id, f.name, f.invite_code, f.telegram_chat_id, f.wallet_counter, f.created_by, f.created_at, f.updated_at 
				  FROM families f
				  JOIN family_members fm ON f.id = fm.family_id
				  WHERE fm.user_id = $1 LIMIT 1`
		err := tx.GetContext(ctx, &f, query, userID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	if f.ID == "" {
		return nil, nil
	}
	return &f, nil
}

func (r *familyRepoImpl) IncrementWalletCounter(ctx context.Context, familyID string) (int, error) {
	var counter int
	err := r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `UPDATE families SET wallet_counter = wallet_counter + 1 WHERE id = $1 RETURNING wallet_counter`
		return tx.QueryRowContext(ctx, query, familyID).Scan(&counter)
	})
	return counter, err
}

func (r *familyRepoImpl) FindMemberByUserID(ctx context.Context, userID string) (*FamilyMember, error) {
	var m FamilyMember
	err := r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `SELECT id, family_id, user_id, role, joined_at FROM family_members WHERE user_id = $1 LIMIT 1`
		err := tx.GetContext(ctx, &m, query, userID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	if m.ID == "" {
		return nil, nil
	}
	return &m, nil
}

func (r *familyRepoImpl) AddMember(ctx context.Context, exec DBExecutor, m *FamilyMember) error {
	query := `INSERT INTO family_members (id, family_id, user_id, role) VALUES ($1, $2, $3, $4) RETURNING joined_at`
	return exec.QueryRowContext(ctx, query, m.ID, m.FamilyID, m.UserID, m.Role).Scan(&m.JoinedAt)
}

func (r *familyRepoImpl) GetMembers(ctx context.Context, familyID string) ([]FamilyMember, error) {
	var members []FamilyMember
	err := r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `SELECT fm.id, fm.family_id, fm.user_id, fm.role, fm.joined_at, 
				  COALESCE(NULLIF(u.name, ''), NULLIF(u.username, ''), 'Member') AS user_name 
				  FROM family_members fm 
				  LEFT JOIN users u ON fm.user_id = u.id 
				  WHERE fm.family_id = $1 
				  ORDER BY fm.joined_at ASC`
		return tx.SelectContext(ctx, &members, query, familyID)
	})
	return members, err
}

func (r *familyRepoImpl) UpdateTelegramChatID(ctx context.Context, familyID string, chatID *int64) error {
	return r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `UPDATE families SET telegram_chat_id = $1, updated_at = NOW() WHERE id = $2`
		_, err := tx.ExecContext(ctx, query, chatID, familyID)
		return err
	})
}

func (r *familyRepoImpl) FindByTelegramChatID(ctx context.Context, chatID int64) (*Family, error) {
	query := `SELECT id, name, invite_code, telegram_chat_id, wallet_counter, created_by, created_at, updated_at FROM families WHERE telegram_chat_id = $1`
	var f Family
	err := r.db.AdminDB().GetContext(ctx, &f, query, chatID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *familyRepoImpl) UpdateFamilyName(ctx context.Context, familyID string, name string) error {
	return r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `UPDATE families SET name = $1, updated_at = NOW() WHERE id = $2`
		_, err := tx.ExecContext(ctx, query, name, familyID)
		return err
	})
}

func (r *familyRepoImpl) CreateWallet(ctx context.Context, w *Wallet) error {
	return r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `INSERT INTO wallets (id, family_id, short_id, name, description, initial_balance, current_balance, minimum_limit, created_by) 
				  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING created_at, updated_at`
		return tx.QueryRowContext(ctx, query, w.ID, w.FamilyID, w.ShortID, w.Name, w.Description, w.InitialBalance, w.CurrentBalance, w.MinimumLimit, w.CreatedBy).
			Scan(&w.CreatedAt, &w.UpdatedAt)
	})
}

func (r *familyRepoImpl) GetWalletsByFamilyID(ctx context.Context, familyID string) ([]Wallet, error) {
	var wallets []Wallet
	err := r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `SELECT id, family_id, short_id, name, description, initial_balance, current_balance, minimum_limit, created_at, updated_at 
				  FROM wallets WHERE family_id = $1 ORDER BY name ASC`
		return tx.SelectContext(ctx, &wallets, query, familyID)
	})
	return wallets, err
}

func (r *familyRepoImpl) GetWalletByID(ctx context.Context, walletID string) (*Wallet, error) {
	var w Wallet
	err := r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `SELECT id, family_id, short_id, name, description, initial_balance, current_balance, minimum_limit, created_at, updated_at 
				  FROM wallets WHERE id = $1`
		err := tx.GetContext(ctx, &w, query, walletID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	if w.ID == "" {
		return nil, nil
	}
	return &w, nil
}

func (r *familyRepoImpl) FindWalletByShortID(ctx context.Context, shortID string) (*Wallet, error) {
	query := `SELECT id, family_id, short_id, name, description, initial_balance, current_balance, minimum_limit, created_at, updated_at 
			  FROM wallets WHERE short_id = $1`
	var w Wallet
	err := r.db.AdminDB().GetContext(ctx, &w, query, shortID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *familyRepoImpl) UpdateWallet(ctx context.Context, walletID string, name string, description *string, currentBalance float64, minimumLimit float64) error {
	return r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `UPDATE wallets SET name = $1, description = $2, current_balance = $3, minimum_limit = $4, updated_at = NOW() WHERE id = $5`
		_, err := tx.ExecContext(ctx, query, name, description, currentBalance, minimumLimit, walletID)
		return err
	})
}

func (r *familyRepoImpl) DeleteWallet(ctx context.Context, walletID string, familyID string) error {
	return r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `DELETE FROM wallets WHERE id = $1 AND family_id = $2`
		_, err := tx.ExecContext(ctx, query, walletID, familyID)
		return err
	})
}

func (r *familyRepoImpl) GetLowBalanceWallets(ctx context.Context) ([]LowBalanceWalletDTO, error) {
	query := `SELECT w.id AS wallet_id, w.short_id, w.name AS wallet_name, w.family_id, w.current_balance, w.minimum_limit, f.telegram_chat_id 
			  FROM wallets w
			  JOIN families f ON w.family_id = f.id
			  WHERE w.current_balance <= w.minimum_limit`
	var list []LowBalanceWalletDTO
	err := r.db.AdminDB().SelectContext(ctx, &list, query)
	return list, err
}

func (r *familyRepoImpl) FindMemberByID(ctx context.Context, memberID string) (*FamilyMember, error) {
	var m FamilyMember
	err := r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `SELECT id, family_id, user_id, role, joined_at FROM family_members WHERE id = $1 LIMIT 1`
		err := tx.GetContext(ctx, &m, query, memberID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	if m.ID == "" {
		return nil, nil
	}
	return &m, nil
}

func (r *familyRepoImpl) RemoveMember(ctx context.Context, memberID string, familyID string) error {
	return r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `DELETE FROM family_members WHERE id = $1 AND family_id = $2`
		_, err := tx.ExecContext(ctx, query, memberID, familyID)
		return err
	})
}
