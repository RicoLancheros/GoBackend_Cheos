package models

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID            uuid.UUID  `json:"id" firestore:"id"`
	ProductID     uuid.UUID  `json:"product_id" firestore:"product_id"`
	UserID        *uuid.UUID `json:"user_id" firestore:"user_id"`           // Nullable if guest review
	CustomerName  string     `json:"customer_name" firestore:"customer_name"`
	CustomerEmail string     `json:"customer_email" firestore:"customer_email"`
	Rating        int        `json:"rating" firestore:"rating"`             // 1-5 stars
	Comment       string     `json:"comment" firestore:"comment"`
	IsApproved    bool       `json:"is_approved" firestore:"is_approved"`   // For moderation
	CreatedAt     time.Time  `json:"created_at" firestore:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" firestore:"updated_at"`
}

// DTOs

type CreateReviewRequest struct {
	ProductID     uuid.UUID `json:"product_id" validate:"required"`
	CustomerName  string    `json:"customer_name" validate:"required,min=2"`
	CustomerEmail string    `json:"customer_email" validate:"required,email"`
	Rating        int       `json:"rating" validate:"required,min=1,max=5"`
	Comment       string    `json:"comment" validate:"omitempty,max=500"`
}

type UpdateReviewRequest struct {
	Rating     int    `json:"rating" validate:"omitempty,min=1,max=5"`
	Comment    string `json:"comment" validate:"omitempty,max=500"`
	IsApproved *bool  `json:"is_approved" validate:"omitempty"` // Only admin can update
}

type ApproveReviewRequest struct {
	IsApproved bool `json:"is_approved" validate:"required"`
}

type ReviewWithProduct struct {
	Review  Review  `json:"review"`
	Product Product `json:"product"`
}

type ReviewListResponse struct {
	Reviews    []Review `json:"reviews"`
	Total      int      `json:"total"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalPages int      `json:"total_pages"`
}

type ProductReviewsResponse struct {
	Reviews       []Review    `json:"reviews"`
	AverageRating float64     `json:"average_rating"`
	TotalReviews  int         `json:"total_reviews"`
	RatingCounts  map[int]int `json:"rating_counts"` // Count per rating (1-5)
}
