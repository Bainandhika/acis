package authentication

type RequestOTPReq struct {
	Email              string `json:"email" binding:"required,email"`
	PhoneNumber        string `json:"phone_number"`
	TelegramIdentifier string `json:"telegram_identifier"`
}

type RequestOTPResponse struct {
	Message             string `json:"message"`
	AuthSession         string `json:"auth_session"`
	TelegramBotUsername string `json:"telegram_bot_username,omitempty"`
	DirectSent          bool   `json:"direct_sent"`
	IsTestUser          bool   `json:"is_test_user,omitempty"`
	TestOTP             string `json:"test_otp,omitempty"`
}

type VerifyOTPReq struct {
	Email              string `json:"email" binding:"required,email"`
	PhoneNumber        string `json:"phone_number"`
	TelegramIdentifier string `json:"telegram_identifier"`
	OTP                string `json:"otp" binding:"required,len=6"`
}

type UserResponse struct {
	ID             string  `json:"id"`
	Email          string  `json:"email"`
	PhoneNumber    string  `json:"phone_number"`
	Name           string  `json:"name"`
	Role           string  `json:"role"`
	AvatarURL      *string `json:"avatar_url,omitempty"`
	TelegramChatID *int64  `json:"telegram_chat_id,omitempty"`
}

type AuthResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"-"`
	User         UserResponse `json:"user"`
}
