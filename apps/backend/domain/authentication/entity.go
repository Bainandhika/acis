package authentication

import "time"

type User struct {
	ID             string    `db:"id" json:"id"`
	Email          string    `db:"email" json:"email"`
	PhoneNumber    string    `db:"phone_number" json:"phone_number"`
	Name           string    `db:"name" json:"name"`
	GoogleID       *string   `db:"google_id" json:"google_id,omitempty"`
	AvatarURL      *string   `db:"avatar_url" json:"avatar_url,omitempty"`
	TelegramChatID *int64    `db:"telegram_chat_id" json:"telegram_chat_id,omitempty"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}
