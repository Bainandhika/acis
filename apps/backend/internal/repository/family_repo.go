package repository

import (
	"context"

	"github.com/Bainandhika/acis/apps/backend/internal/database"
	"github.com/Bainandhika/acis/apps/backend/internal/domain"
)

// FamilyRepository defines the contract for family data operations
type FamilyRepository interface {
	Create(ctx context.Context, exec DBExecutor, family *domain.Family) error
	FindByInviteCode(ctx context.Context, exec DBExecutor, code string) (*domain.Family, error)
	AddMember(ctx context.Context, exec DBExecutor, member *domain.FamilyMember) error
	GetMemberByUserID(ctx context.Context, exec DBExecutor, userID string) (*domain.FamilyMember, error)
	GetFamilyMembers(ctx context.Context, exec DBExecutor, familyID string) ([]domain.FamilyMember, error)
}

type familyRepo struct {
	db *database.AppDB
}

// NewFamilyRepository creates a new instance of FamilyRepository
func NewFamilyRepository(db *database.AppDB) FamilyRepository {
	return &familyRepo{db: db}
}

func (r *familyRepo) Create(ctx context.Context, exec DBExecutor, family *domain.Family) error {
	query := `INSERT INTO families (id, name, invite_code, created_by) 
	          VALUES ($1, $2, $3, $4) RETURNING created_at`
	return exec.QueryRowContext(ctx, query, family.ID, family.Name, family.InviteCode, family.CreatedBy).Scan(&family.CreatedAt)
}

func (r *familyRepo) FindByInviteCode(ctx context.Context, exec DBExecutor, code string) (*domain.Family, error) {
	query := `SELECT id, name, invite_code, created_by, created_at, updated_at 
	          FROM families WHERE invite_code = $1`

	family := &domain.Family{}
	err := exec.GetContext(ctx, family, query, code)
	if err != nil {
		return nil, err
	}
	return family, nil
}

func (r *familyRepo) AddMember(ctx context.Context, exec DBExecutor, member *domain.FamilyMember) error {
	query := `INSERT INTO family_members (id, family_id, user_id, role) 
	          VALUES ($1, $2, $3, $4) RETURNING joined_at`
	return exec.QueryRowContext(ctx, query, member.ID, member.FamilyID, member.UserID, member.Role).Scan(&member.JoinedAt)
}

func (r *familyRepo) GetMemberByUserID(ctx context.Context, exec DBExecutor, userID string) (*domain.FamilyMember, error) {
	query := `SELECT id, family_id, user_id, role, joined_at 
	          FROM family_members WHERE user_id = $1 LIMIT 1`

	member := &domain.FamilyMember{}
	err := exec.GetContext(ctx, member, query, userID)
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (r *familyRepo) GetFamilyMembers(ctx context.Context, exec DBExecutor, familyID string) ([]domain.FamilyMember, error) {
	query := `SELECT fm.id, fm.family_id, fm.user_id, fm.role, fm.joined_at, 
	                 u.name, u.email 
	          FROM family_members fm
	          JOIN users u ON fm.user_id = u.id
	          WHERE fm.family_id = $1 
	          ORDER BY fm.joined_at ASC`

	var members []domain.FamilyMember
	err := exec.SelectContext(ctx, &members, query, familyID)
	return members, err
}
