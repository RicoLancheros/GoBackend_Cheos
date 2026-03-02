package models

import (
	"time"

	"github.com/google/uuid"
)

type PasswordReset struct {
	ID        uuid.UUID `json:"id" firestore:"id"`
	UserID    uuid.UUID `json:"user_id" firestore:"user_id"`
	Email     string    `json:"email" firestore:"email"`
	Token     string    `json:"token" firestore:"token"`
	Used      bool      `json:"used" firestore:"used"`
	CreatedAt time.Time `json:"created_at" firestore:"created_at"`
	ExpiresAt time.Time `json:"expires_at" firestore:"expires_at"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}
