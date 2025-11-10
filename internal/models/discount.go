package models

import (
	"time"

	"github.com/google/uuid"
)

type DiscountType string

const (
	DiscountPercentage  DiscountType = "PERCENTAGE"
	DiscountFixedAmount DiscountType = "FIXED_AMOUNT"
)

type DiscountCode struct {
	ID          uuid.UUID    `json:"id" firestore:"id"`
	Code        string       `json:"code" firestore:"code"`
	Description string       `json:"description" firestore:"description"`
	Type        DiscountType `json:"type" firestore:"type"`
	Value       float64      `json:"value" firestore:"value"`
	MinPurchase *float64     `json:"min_purchase" firestore:"min_purchase"`
	MaxUses     *int         `json:"max_uses" firestore:"max_uses"`
	UsedCount   int          `json:"used_count" firestore:"used_count"`
	StartDate   time.Time    `json:"start_date" firestore:"start_date"`
	EndDate     time.Time    `json:"end_date" firestore:"end_date"`
	IsActive    bool         `json:"is_active" firestore:"is_active"`
	CreatedAt   time.Time    `json:"created_at" firestore:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" firestore:"updated_at"`
}

// DTOs

type CreateDiscountCodeRequest struct {
	Code        string       `json:"code" validate:"required,min=3,max=20"`
	Description string       `json:"description" validate:"required"`
	Type        DiscountType `json:"type" validate:"required"`
	Value       float64      `json:"value" validate:"required,gt=0"`
	MinPurchase *float64     `json:"min_purchase" validate:"omitempty,gt=0"`
	MaxUses     *int         `json:"max_uses" validate:"omitempty,gt=0"`
	StartDate   time.Time    `json:"start_date" validate:"required"`
	EndDate     time.Time    `json:"end_date" validate:"required"`
	IsActive    bool         `json:"is_active"`
}

type UpdateDiscountCodeRequest struct {
	Code        string       `json:"code" validate:"omitempty,min=3,max=20"`
	Description string       `json:"description"`
	Type        DiscountType `json:"type" validate:"omitempty"`
	Value       float64      `json:"value" validate:"omitempty,gt=0"`
	MinPurchase *float64     `json:"min_purchase" validate:"omitempty,gt=0"`
	MaxUses     *int         `json:"max_uses" validate:"omitempty,gt=0"`
	StartDate   time.Time    `json:"start_date" validate:"omitempty"`
	EndDate     time.Time    `json:"end_date" validate:"omitempty"`
	IsActive    *bool        `json:"is_active" validate:"omitempty"`
}

type ValidateDiscountRequest struct {
	Code          string  `json:"code" validate:"required"`
	PurchaseTotal float64 `json:"purchase_total" validate:"required,gt=0"`
}

type ValidateDiscountResponse struct {
	Valid          bool         `json:"valid"`
	DiscountCode   *DiscountCode `json:"discount_code"`
	DiscountAmount float64      `json:"discount_amount"`
	Message        string       `json:"message"`
}

type DiscountCodeListResponse struct {
	DiscountCodes []DiscountCode `json:"discount_codes"`
	Total         int            `json:"total"`
	Page          int            `json:"page"`
	PageSize      int            `json:"page_size"`
	TotalPages    int            `json:"total_pages"`
}
