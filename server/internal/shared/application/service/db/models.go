package db

import (
	"time"
)

type User struct {
	ID            string    `json:"id" db:"id"`
	TenantID      string    `json:"tenant_id" db:"tenant_id"`
	Email         string    `json:"email" db:"email"`
	Name          string    `json:"name" db:"name"`
	Role          string    `json:"role" db:"role"`
	PasswordHash  string    `json:"-" db:"password_hash"`
	PhoneNumber   *string   `json:"phone_number,omitempty" db:"phone_number"`
	PhoneVerified bool      `json:"phone_verified" db:"phone_verified"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type Room struct {
	ID        string    `json:"id" db:"id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	Name      string    `json:"name" db:"name"`
	Type      string    `json:"type" db:"type"` // "direct" or "group"
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type RoomMember struct {
	RoomID string `json:"room_id" db:"room_id"`
	UserID string `json:"user_id" db:"user_id"`
}

type Message struct {
	ID       string    `json:"id" db:"id"`
	TenantID string    `json:"tenant_id" db:"tenant_id"`
	RoomID   string    `json:"room_id" db:"room_id"`
	UserID   string    `json:"user_id" db:"user_id"`
	Content  string    `json:"content" db:"content"`
	SentAt   time.Time `json:"sent_at" db:"sent_at"`
}

type CreateRoomRequest struct {
	Name    string   `json:"name" binding:"required"`
	Type    string   `json:"type" binding:"required,oneof=direct group"`
	UserIDs []string `json:"user_ids" binding:"required"`
}

type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

type DirectMessageRequest struct {
	RecipientID string `json:"recipient_id" binding:"required"`
	Content     string `json:"content" binding:"required"`
}

type PhoneVerificationCode struct {
	ID          string    `json:"id" db:"id"`
	PhoneNumber string    `json:"phone_number" db:"phone_number"`
	Code        string    `json:"code" db:"code"`
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at"`
	Used        bool      `json:"used" db:"used"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type PhoneAuthRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

type PhoneVerifyRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Code        string `json:"code" binding:"required"`
}

type PhoneSignupRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
}

type PhoneSigninRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Code        string `json:"code" binding:"required"`
}
