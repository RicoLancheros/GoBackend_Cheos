package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleCustomer UserRole = "CUSTOMER"
	RoleAdmin    UserRole = "ADMIN"
)

type User struct {
	ID        uuid.UUID `json:"id" firestore:"id"`
	Email     string    `json:"email" firestore:"email"`
	Password  string    `json:"-" firestore:"password"` // No se devuelve en JSON pero se guarda en Firestore
	Name      string    `json:"name" firestore:"name"`
	Phone     string    `json:"phone" firestore:"phone"`
	Role      UserRole  `json:"role" firestore:"role"`
	IsActive  bool      `json:"is_active" firestore:"is_active"`
	CreatedAt time.Time `json:"created_at" firestore:"created_at"`
	UpdatedAt time.Time `json:"updated_at" firestore:"updated_at"`
}

// DTOs (Data Transfer Objects)

type RegisterRequest struct {
	Email    string   `json:"email" validate:"required,email"`
	Password string   `json:"password" validate:"required,min=8"`
	Name     string   `json:"name" validate:"required,min=2"`
	Phone    string   `json:"phone" validate:"required"`
	Role     UserRole `json:"role" validate:"omitempty,oneof=CUSTOMER ADMIN"` // Opcional, por defecto CUSTOMER
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UpdateProfileRequest struct {
	Name  string `json:"name" validate:"omitempty,min=2"`
	Phone string `json:"phone" validate:"omitempty"`
}

type UpdateUserByIDRequest struct {
	Name     *string   `json:"name" validate:"omitempty,min=2"`
	Phone    *string   `json:"phone" validate:"omitempty"`
	Email    *string   `json:"email" validate:"omitempty,email"`
	Password *string   `json:"password" validate:"omitempty,min=8"`
	Role     *UserRole `json:"role" validate:"omitempty,oneof=CUSTOMER ADMIN"`
	IsActive *bool     `json:"is_active" validate:"omitempty"`
}
