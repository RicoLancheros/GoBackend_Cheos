package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `json:"id" firestore:"id"`
	Name        string    `json:"name" firestore:"name"`
	Description string    `json:"description" firestore:"description"`
	Price       float64   `json:"price" firestore:"price"`
	Weight      int       `json:"weight" firestore:"weight"` // en gramos
	Stock       int       `json:"stock" firestore:"stock"`
	Category    string    `json:"category" firestore:"category"`
	Images      []string  `json:"images" firestore:"images"`
	IsActive    bool      `json:"is_active" firestore:"is_active"`
	IsFeatured  bool      `json:"is_featured" firestore:"is_featured"`
	CreatedAt   time.Time `json:"created_at" firestore:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" firestore:"updated_at"`
}

// DTOs

type CreateProductRequest struct {
	Name        string   `json:"name" validate:"required,min=3"`
	Description string   `json:"description" validate:"required"`
	Price       float64  `json:"price" validate:"required,gt=0"`
	Weight      int      `json:"weight" validate:"required,gt=0"` // en gramos
	Stock       int      `json:"stock" validate:"required,gte=0"`
	Category    string   `json:"category" validate:"required"`
	Images      []string `json:"images" validate:"required,min=1"`
	IsFeatured  bool     `json:"is_featured"`
}

type UpdateProductRequest struct {
	Name        *string   `json:"name" validate:"omitempty,min=3"`
	Description *string   `json:"description" validate:"omitempty"`
	Price       *float64  `json:"price" validate:"omitempty,gt=0"`
	Weight      *int      `json:"weight" validate:"omitempty,gt=0"`
	Stock       *int      `json:"stock" validate:"omitempty,gte=0"`
	Category    *string   `json:"category" validate:"omitempty"`
	Images      []string  `json:"images" validate:"omitempty"`
	IsActive    *bool     `json:"is_active" validate:"omitempty"`
	IsFeatured  *bool     `json:"is_featured" validate:"omitempty"`
}

type UpdateStockRequest struct {
	Quantity int `json:"quantity" validate:"required"`
}

type PaginatedProductsResponse struct {
	Products    []*Product `json:"products"`
	Total       int        `json:"total"`
	Page        int        `json:"page"`
	PageSize    int        `json:"page_size"`
	TotalPages  int        `json:"total_pages"`
	HasNext     bool       `json:"has_next"`
	HasPrevious bool       `json:"has_previous"`
}
