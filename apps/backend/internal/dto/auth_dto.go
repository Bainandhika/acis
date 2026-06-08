package dto

// RequestOTPRequest represents the JSON payload to request an OTP
type RequestOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// VerifyOTPRequest represents the JSON payload to verify an OTP
type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

// AuthResponse represents the successful login response
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// UserResponse is a simplified user object for the frontend
type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"` // 'admin' or 'member'
}
