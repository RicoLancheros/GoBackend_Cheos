package services

import (
	"context"
	"math"

	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/repository"
	"github.com/google/uuid"
)

type ReviewService struct {
	reviewRepo  *repository.ReviewRepository
	productRepo *repository.ProductRepository
}

func NewReviewService(reviewRepo *repository.ReviewRepository, productRepo *repository.ProductRepository) *ReviewService {
	return &ReviewService{
		reviewRepo:  reviewRepo,
		productRepo: productRepo,
	}
}

// CreateReview creates a new review
func (s *ReviewService) CreateReview(ctx context.Context, req *models.CreateReviewRequest, userID *uuid.UUID) (*models.Review, error) {
	// Verify product exists
	_, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	review := &models.Review{
		ProductID:     req.ProductID,
		UserID:        userID,
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		Rating:        req.Rating,
		Comment:       req.Comment,
	}

	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}

	return review, nil
}

// GetReviewByID gets a review by ID
func (s *ReviewService) GetReviewByID(ctx context.Context, id uuid.UUID) (*models.Review, error) {
	return s.reviewRepo.GetByID(ctx, id)
}

// GetAllReviews gets all reviews with pagination
func (s *ReviewService) GetAllReviews(ctx context.Context, page int, pageSize int) (*models.ReviewListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	limit := pageSize

	reviews, err := s.reviewRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	// Convert to slice of Review
	reviewList := make([]models.Review, 0, len(reviews))
	for _, review := range reviews {
		reviewList = append(reviewList, *review)
	}

	total, err := s.reviewRepo.CountReviews(ctx)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &models.ReviewListResponse{
		Reviews:    reviewList,
		Total:      int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetProductReviews gets reviews for a specific product with stats
func (s *ReviewService) GetProductReviews(ctx context.Context, productID uuid.UUID) (*models.ProductReviewsResponse, error) {
	// Get all approved reviews for the product
	reviews, err := s.reviewRepo.GetAllByProductID(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Convert to slice and calculate stats
	reviewList := make([]models.Review, 0, len(reviews))
	ratingCounts := make(map[int]int)
	var totalRating float64

	for _, review := range reviews {
		if review.IsApproved {
			reviewList = append(reviewList, *review)
			ratingCounts[review.Rating]++
			totalRating += float64(review.Rating)
		}
	}

	var averageRating float64
	if len(reviewList) > 0 {
		averageRating = totalRating / float64(len(reviewList))
	}

	return &models.ProductReviewsResponse{
		Reviews:       reviewList,
		AverageRating: averageRating,
		TotalReviews:  len(reviewList),
		RatingCounts:  ratingCounts,
	}, nil
}

// UpdateReview updates a review
func (s *ReviewService) UpdateReview(ctx context.Context, id uuid.UUID, req *models.UpdateReviewRequest) (*models.Review, error) {
	review, err := s.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if present
	if req.Rating > 0 {
		review.Rating = req.Rating
	}

	if req.Comment != "" {
		review.Comment = req.Comment
	}

	if req.IsApproved != nil {
		review.IsApproved = *req.IsApproved
	}

	if err := s.reviewRepo.Update(ctx, review); err != nil {
		return nil, err
	}

	return review, nil
}

// DeleteReview deletes a review
func (s *ReviewService) DeleteReview(ctx context.Context, id uuid.UUID) error {
	return s.reviewRepo.Delete(ctx, id)
}
