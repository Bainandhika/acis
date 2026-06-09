package dto

// CreateFamilyRequest represents the payload to create a new family
type CreateFamilyRequest struct {
	Name string `json:"name" binding:"required,min=3,max=100"`
}

// JoinFamilyRequest represents the payload to join a family via invite code
type JoinFamilyRequest struct {
	InviteCode string `json:"invite_code" binding:"required,len=6"`
}

// FamilyResponse represents the family data returned to client
type FamilyResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	InviteCode string `json:"invite_code"`
	CreatedAt  string `json:"created_at"`
}

// FamilyMemberResponse represents a member in the family
type FamilyMemberResponse struct {
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"` // "admin" or "member"
	JoinedAt string `json:"joined_at"`
}
