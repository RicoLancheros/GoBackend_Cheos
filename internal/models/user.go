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

type Gender string

const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
	GenderOther  Gender = "OTHER"
)

type User struct {
	ID           uuid.UUID  `json:"id" firestore:"id"`
	Email        string     `json:"email" firestore:"email"`
	Password     string     `json:"-" firestore:"password"`
	Name         string     `json:"name" firestore:"name"`
	Phone        string     `json:"phone" firestore:"phone"`
	City         *string    `json:"city" firestore:"city"`
	Municipality *string    `json:"municipality" firestore:"municipality"`
	Neighborhood *string    `json:"neighborhood" firestore:"neighborhood"`
	Gender       *Gender    `json:"gender" firestore:"gender"`
	BirthDate    *string    `json:"birth_date" firestore:"birth_date"`
	Role         UserRole   `json:"role" firestore:"role"`
	IsActive     bool       `json:"is_active" firestore:"is_active"`
	CreatedAt    time.Time  `json:"created_at" firestore:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" firestore:"updated_at"`
}

// DTOs (Data Transfer Objects)

type RegisterRequest struct {
	Email        string   `json:"email" validate:"required,email"`
	Password     string   `json:"password" validate:"required,min=6"`
	Name         string   `json:"name" validate:"required,min=2,excludesall=0123456789"`
	Phone        string   `json:"phone" validate:"required"`
	City         *string  `json:"city" validate:"omitempty"`
	Municipality *string  `json:"municipality" validate:"omitempty"`
	Neighborhood *string  `json:"neighborhood" validate:"omitempty"`
	Gender       *Gender  `json:"gender" validate:"omitempty,oneof=MALE FEMALE OTHER"`
	BirthDate    *string  `json:"birth_date" validate:"omitempty"`
	Role         UserRole `json:"role" validate:"omitempty,oneof=CUSTOMER ADMIN"`
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
	Name         string  `json:"name" validate:"omitempty,min=2"`
	Phone        string  `json:"phone" validate:"omitempty"`
	City         *string `json:"city" validate:"omitempty"`
	Municipality *string `json:"municipality" validate:"omitempty"`
	Neighborhood *string `json:"neighborhood" validate:"omitempty"`
	Gender       *Gender `json:"gender" validate:"omitempty,oneof=MALE FEMALE OTHER"`
	BirthDate    *string `json:"birth_date" validate:"omitempty"`
}

type UpdateUserByIDRequest struct {
	Name         *string   `json:"name" validate:"omitempty,min=2"`
	Phone        *string   `json:"phone" validate:"omitempty"`
	Email        *string   `json:"email" validate:"omitempty,email"`
	Password     *string   `json:"password" validate:"omitempty,min=8"`
	City         *string   `json:"city" validate:"omitempty"`
	Municipality *string   `json:"municipality" validate:"omitempty"`
	Neighborhood *string   `json:"neighborhood" validate:"omitempty"`
	Gender       *Gender   `json:"gender" validate:"omitempty,oneof=MALE FEMALE OTHER"`
	BirthDate    *string   `json:"birth_date" validate:"omitempty"`
	Role         *UserRole `json:"role" validate:"omitempty,oneof=CUSTOMER ADMIN"`
	IsActive     *bool     `json:"is_active" validate:"omitempty"`
}
