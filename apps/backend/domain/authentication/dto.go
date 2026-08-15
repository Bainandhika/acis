package authentication

type RequestOTPReq struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyOTPReq struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

type UserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	AvatarURL *string `json:"avatar_url"`
}

type AuthResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"-"`
	User         UserResponse `json:"user"`
}
