package service

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/Bainandhika/acis/apps/backend/internal/database"
	"github.com/Bainandhika/acis/apps/backend/internal/domain"
	"github.com/Bainandhika/acis/apps/backend/internal/dto"
	"github.com/Bainandhika/acis/apps/backend/internal/repository"
	"github.com/google/uuid"
)

// FamilyService defines business logic for family operations
type FamilyService interface {
	CreateFamily(ctx context.Context, userID string, req dto.CreateFamilyRequest) (*dto.FamilyResponse, error)
	JoinFamily(ctx context.Context, userID string, req dto.JoinFamilyRequest) (*dto.FamilyResponse, error)
	GetMyFamily(ctx context.Context, userID string) (*dto.FamilyResponse, error)
	GetFamilyMembers(ctx context.Context, familyID string) ([]dto.FamilyMemberResponse, error)
}

type familyService struct {
	familyRepo repository.FamilyRepository
	userRepo   repository.UserRepository
	db         *database.AppDB // for transaction handling
}

// NewFamilyService creates a new instance of FamilyService
func NewFamilyService(familyRepo repository.FamilyRepository, userRepo repository.UserRepository, db *database.AppDB) FamilyService {
	return &familyService{
		familyRepo: familyRepo,
		userRepo:   userRepo,
		db:         db,
	}
}

// generateInviteCode creates a random 6-char uppercase code
func generateInviteCode() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 6)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (s *familyService) CreateFamily(ctx context.Context, userID string, req dto.CreateFamilyRequest) (*dto.FamilyResponse, error) {
	// Start transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.New("failed to start transaction")
	}
	defer tx.Rollback()

	// Create family entity
	family := &domain.Family{
		ID:         uuid.NewString(),
		Name:       req.Name,
		InviteCode: generateInviteCode(),
		CreatedBy:  &userID,
	}

	// Save family to DB
	if err := s.familyRepo.Create(ctx, tx, family); err != nil {
		return nil, err
	}

	// Auto-assign creator as admin member
	member := &domain.FamilyMember{
		ID:       uuid.NewString(),
		FamilyID: family.ID,
		UserID:   userID,
		Role:     "admin",
	}
	if err := s.familyRepo.AddMember(ctx, tx, member); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &dto.FamilyResponse{
		ID:         family.ID,
		Name:       family.Name,
		InviteCode: family.InviteCode,
		CreatedAt:  family.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *familyService) JoinFamily(ctx context.Context, userID string, req dto.JoinFamilyRequest) (*dto.FamilyResponse, error) {
	// Find family by invite code
	family, err := s.familyRepo.FindByInviteCode(ctx, s.db, strings.ToUpper(req.InviteCode))
	if err != nil {
		return nil, errors.New("invalid invite code")
	}

	// Check if user already member
	_, err = s.familyRepo.GetMemberByUserID(ctx, s.db, userID)
	if err == nil {
		return nil, errors.New("user already joined a family")
	}

	// Add user as member
	member := &domain.FamilyMember{
		ID:       uuid.NewString(),
		FamilyID: family.ID,
		UserID:   userID,
		Role:     "member",
	}
	if err := s.familyRepo.AddMember(ctx, s.db, member); err != nil {
		return nil, err
	}

	return &dto.FamilyResponse{
		ID:         family.ID,
		Name:       family.Name,
		InviteCode: family.InviteCode,
		CreatedAt:  family.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *familyService) GetMyFamily(ctx context.Context, userID string) (*dto.FamilyResponse, error) {
	member, err := s.familyRepo.GetMemberByUserID(ctx, s.db, userID)
	if err != nil {
		return nil, errors.New("user not in any family")
	}

	// Fetch full family details (we need a GetByID method, add it to repo if missing)
	query := `SELECT id, name, invite_code, created_by, created_at, updated_at 
	          FROM families WHERE id = $1`
	family := &domain.Family{}
	err = s.db.GetContext(ctx, family, query, member.FamilyID)
	if err != nil {
		return nil, err
	}

	return &dto.FamilyResponse{
		ID:         family.ID,
		Name:       family.Name,
		InviteCode: family.InviteCode,
		CreatedAt:  family.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *familyService) GetFamilyMembers(ctx context.Context, familyID string) ([]dto.FamilyMemberResponse, error) {
	members, err := s.familyRepo.GetFamilyMembers(ctx, s.db, familyID)
	if err != nil {
		return nil, err
	}

	response := make([]dto.FamilyMemberResponse, len(members))
	for i, m := range members {
		// fetch user details for each member
		name := ""
		email := ""
		if user, err := s.userRepo.FindByUserID(ctx, m.UserID); err == nil && user != nil {
			name = user.Name
			email = user.Email
		}

		response[i] = dto.FamilyMemberResponse{
			UserID:   m.UserID,
			Name:     name,
			Email:    email,
			Role:     m.Role,
			JoinedAt: m.JoinedAt.Format(time.RFC3339),
		}
	}
	return response, nil
}
