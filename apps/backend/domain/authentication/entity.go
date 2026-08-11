package authentication

import "time"

type User struct {
	ID        string    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Name      string    `db:"name" json:"name"`
	GoogleID  *string   `db:"google_id" json:"google_id"`
	AvatarURL *string   `db:"avatar_url" json:"avatar_url"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type OTPCode struct {
	ID        string    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	CodeHash  string    `db:"code_hash" json:"-"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	IsUsed    bool      `db:"is_used" json:"is_used"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
