package models

import (
	"time"

	"github.com/google/uuid"
)

type Schedule struct {
	Monday    string `json:"monday" firestore:"monday"`
	Tuesday   string `json:"tuesday" firestore:"tuesday"`
	Wednesday string `json:"wednesday" firestore:"wednesday"`
	Thursday  string `json:"thursday" firestore:"thursday"`
	Friday    string `json:"friday" firestore:"friday"`
	Saturday  string `json:"saturday" firestore:"saturday"`
	Sunday    string `json:"sunday" firestore:"sunday"`
}

type Location struct {
	ID         uuid.UUID `json:"id" firestore:"id"`
	Name       string    `json:"name" firestore:"name"`
	Address    string    `json:"address" firestore:"address"`
	City       string    `json:"city" firestore:"city"`
	Department string    `json:"department" firestore:"department"`
	Phone      string    `json:"phone" firestore:"phone"`
	Latitude   float64   `json:"latitude" firestore:"latitude"`
	Longitude  float64   `json:"longitude" firestore:"longitude"`
	Schedule   *Schedule `json:"schedule" firestore:"schedule"`
	IsActive   bool      `json:"is_active" firestore:"is_active"`
	CreatedAt  time.Time `json:"created_at" firestore:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" firestore:"updated_at"`
}

// DTOs

type CreateLocationRequest struct {
	Name       string    `json:"name" validate:"required,min=3"`
	Address    string    `json:"address" validate:"required"`
	City       string    `json:"city" validate:"required"`
	Department string    `json:"department" validate:"required"`
	Phone      string    `json:"phone" validate:"required"`
	Latitude   float64   `json:"latitude" validate:"required"`
	Longitude  float64   `json:"longitude" validate:"required"`
	Schedule   *Schedule `json:"schedule" validate:"required"`
	IsActive   bool      `json:"is_active"`
}

type UpdateLocationRequest struct {
	Name       string    `json:"name" validate:"omitempty,min=3"`
	Address    string    `json:"address" validate:"omitempty"`
	City       string    `json:"city" validate:"omitempty"`
	Department string    `json:"department" validate:"omitempty"`
	Phone      string    `json:"phone" validate:"omitempty"`
	Latitude   float64   `json:"latitude" validate:"omitempty"`
	Longitude  float64   `json:"longitude" validate:"omitempty"`
	Schedule   *Schedule `json:"schedule" validate:"omitempty"`
	IsActive   *bool     `json:"is_active" validate:"omitempty"`
}

type LocationListResponse struct {
	Locations  []Location `json:"locations"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}
