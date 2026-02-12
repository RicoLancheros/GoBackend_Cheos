package models

import (
	"time"

	"github.com/google/uuid"
)

type Cart struct {
	UserID    uuid.UUID  `json:"user_id" firestore:"user_id"`
	Items     []CartItem `json:"items" firestore:"items"`
	UpdatedAt time.Time  `json:"updated_at" firestore:"updated_at"`
}

type CartItem struct {
	ProductID    uuid.UUID `json:"product_id" firestore:"product_id"`
	ProductName  string    `json:"product_name" firestore:"product_name"`
	ProductPrice float64   `json:"product_price" firestore:"product_price"`
	ProductImage string    `json:"product_image" firestore:"product_image"`
	Quantity     int       `json:"quantity" firestore:"quantity"`
}

// DTOs

type AddToCartRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,gt=0"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" validate:"required,gt=0"`
}

type SyncCartRequest struct {
	Items []AddToCartRequest `json:"items" validate:"required"`
}
