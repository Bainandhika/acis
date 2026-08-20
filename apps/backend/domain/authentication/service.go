package authentication

import (
	"context"
	"fmt"
)

type AuthService interface {
	Provision(ctx context.Context, userID, email string, req ProvisionRequest) (*UserProfileResponse, error)
	GetMe(ctx context.Context, userID, email string) (*UserProfileResponse, error)
}

type authService struct {
	repo AuthRepository
}

func NewService(repo AuthRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) Provision(ctx context.Context, userID, email string, req ProvisionRequest) (*UserProfileResponse, error) {
	name := req.Name
	if name == "" {
		name = "User"
	}

	userToProvision := &User{
		ID:        userID,
		Name:      name,
		Username:  req.Username,
		AvatarURL: req.AvatarURL,
	}

	user, err := s.repo.ProvisionUser(ctx, userID, email, userToProvision)
	if err != nil {
		return nil, fmt.Errorf("failed to provision user: %w", err)
	}

	memberships, err := s.repo.GetUserMemberships(ctx, userID, email)
	if err != nil {
		// Non-fatal, return empty memberships if error
		memberships = []FamilyMembership{}
	}

	return &UserProfileResponse{
		ID:             user.ID,
		Username:       user.Username,
		Name:           user.Name,
		AvatarURL:      user.AvatarURL,
		TelegramChatID: user.TelegramChatID,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
		Memberships:    memberships,
	}, nil
}

func (s *authService) GetMe(ctx context.Context, userID, email string) (*UserProfileResponse, error) {
	user, err := s.repo.FindByUserID(ctx, userID, email)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user profile not found")
	}

	memberships, err := s.repo.GetUserMemberships(ctx, userID, email)
	if err != nil {
		memberships = []FamilyMembership{}
	}

	return &UserProfileResponse{
		ID:             user.ID,
		Username:       user.Username,
		Name:           user.Name,
		AvatarURL:      user.AvatarURL,
		TelegramChatID: user.TelegramChatID,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
		Memberships:    memberships,
	}, nil
}
