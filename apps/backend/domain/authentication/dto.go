package authentication

import "time"

type ProvisionRequest struct {
	Name      string  `json:"name"`
	Username  string  `json:"username"`
	AvatarURL *string `json:"avatar_url"`
}

type UserProfileResponse struct {
	ID             string              `json:"id"`
	Username       string              `json:"username"`
	Name           string              `json:"name"`
	AvatarURL      *string             `json:"avatar_url,omitempty"`
	TelegramChatID *int64              `json:"telegram_chat_id,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	Memberships    []FamilyMembership  `json:"memberships,omitempty"`
}

type FamilyMembership struct {
	FamilyID   string `json:"family_id"`
	FamilyName string `json:"family_name"`
	Role       string `json:"role"`
}
