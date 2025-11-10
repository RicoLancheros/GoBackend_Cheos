package services

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/repository"
	"github.com/google/uuid"
)

type DiscountService struct {
	discountRepo *repository.DiscountRepository
}

func NewDiscountService(discountRepo *repository.DiscountRepository) *DiscountService {
	return &DiscountService{
		discountRepo: discountRepo,
	}
}

// CreateDiscountCode creates a new discount code
func (s *DiscountService) CreateDiscountCode(ctx context.Context, req *models.CreateDiscountCodeRequest) (*models.DiscountCode, error) {
	// Validate dates
	if req.EndDate.Before(req.StartDate) {
		return nil, errors.New("end date must be after start date")
	}

	// Validate code doesn't exist
	existing, _ := s.discountRepo.GetByCode(ctx, req.Code)
	if existing != nil {
		return nil, errors.New("discount code already exists")
	}

	// Validate discount value
	if req.Type == models.DiscountPercentage {
		if req.Value > 100 {
			return nil, errors.New("percentage discount cannot be greater than 100%")
		}
	}

	discount := &models.DiscountCode{
		Code:        req.Code,
		Description: req.Description,
		Type:        req.Type,
		Value:       req.Value,
		MinPurchase: req.MinPurchase,
		MaxUses:     req.MaxUses,
		UsedCount:   0,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		IsActive:    req.IsActive,
	}

	if err := s.discountRepo.Create(ctx, discount); err != nil {
		return nil, err
	}

	return discount, nil
}

// GetDiscountCodeByID gets a discount code by ID
func (s *DiscountService) GetDiscountCodeByID(ctx context.Context, id uuid.UUID) (*models.DiscountCode, error) {
	return s.discountRepo.GetByID(ctx, id)
}

// GetAllDiscountCodes gets all discount codes with pagination
func (s *DiscountService) GetAllDiscountCodes(ctx context.Context, page int, pageSize int) (*models.DiscountCodeListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	limit := pageSize

	discounts, err := s.discountRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	// Convert to slice of DiscountCode
	discountList := make([]models.DiscountCode, 0, len(discounts))
	for _, discount := range discounts {
		discountList = append(discountList, *discount)
	}

	total, err := s.discountRepo.CountDiscountCodes(ctx)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &models.DiscountCodeListResponse{
		DiscountCodes: discountList,
		Total:         int(total),
		Page:          page,
		PageSize:      pageSize,
		TotalPages:    totalPages,
	}, nil
}

// UpdateDiscountCode updates a discount code
func (s *DiscountService) UpdateDiscountCode(ctx context.Context, id uuid.UUID, req *models.UpdateDiscountCodeRequest) (*models.DiscountCode, error) {
	discount, err := s.discountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if present
	if req.Code != "" {
		// Check that new code doesn't exist
		existing, _ := s.discountRepo.GetByCode(ctx, req.Code)
		if existing != nil && existing.ID != id {
			return nil, errors.New("discount code already exists")
		}
		discount.Code = req.Code
	}

	if req.Description != "" {
		discount.Description = req.Description
	}

	if req.Type != "" {
		discount.Type = req.Type
	}

	if req.Value > 0 {
		if req.Type == models.DiscountPercentage && req.Value > 100 {
			return nil, errors.New("percentage discount cannot be greater than 100%")
		}
		discount.Value = req.Value
	}

	if req.MinPurchase != nil {
		discount.MinPurchase = req.MinPurchase
	}

	if req.MaxUses != nil {
		discount.MaxUses = req.MaxUses
	}

	if !req.StartDate.IsZero() {
		discount.StartDate = req.StartDate
	}

	if !req.EndDate.IsZero() {
		discount.EndDate = req.EndDate
	}

	if req.IsActive != nil {
		discount.IsActive = *req.IsActive
	}

	// Validate dates
	if discount.EndDate.Before(discount.StartDate) {
		return nil, errors.New("end date must be after start date")
	}

	if err := s.discountRepo.Update(ctx, discount); err != nil {
		return nil, err
	}

	return discount, nil
}

// DeleteDiscountCode deletes a discount code (soft delete)
func (s *DiscountService) DeleteDiscountCode(ctx context.Context, id uuid.UUID) error {
	return s.discountRepo.Delete(ctx, id)
}

// ValidateDiscountCode validates a discount code and calculates discount
func (s *DiscountService) ValidateDiscountCode(ctx context.Context, req *models.ValidateDiscountRequest) (*models.ValidateDiscountResponse, error) {
	// Get discount code
	discount, err := s.discountRepo.GetByCode(ctx, req.Code)
	if err != nil {
		return &models.ValidateDiscountResponse{
			Valid:          false,
			Message:        "Discount code not found",
			DiscountAmount: 0,
		}, nil
	}

	// Validate if active
	if !discount.IsActive {
		return &models.ValidateDiscountResponse{
			Valid:          false,
			Message:        "Discount code is inactive",
			DiscountAmount: 0,
			DiscountCode:   discount,
		}, nil
	}

	now := time.Now()

	// Validate dates
	if now.Before(discount.StartDate) {
		return &models.ValidateDiscountResponse{
			Valid:          false,
			Message:        "Discount code not yet valid",
			DiscountAmount: 0,
			DiscountCode:   discount,
		}, nil
	}

	if now.After(discount.EndDate) {
		return &models.ValidateDiscountResponse{
			Valid:          false,
			Message:        "Discount code expired",
			DiscountAmount: 0,
			DiscountCode:   discount,
		}, nil
	}

	// Validate max uses
	if discount.MaxUses != nil && discount.UsedCount >= *discount.MaxUses {
		return &models.ValidateDiscountResponse{
			Valid:          false,
			Message:        "Discount code usage limit reached",
			DiscountAmount: 0,
			DiscountCode:   discount,
		}, nil
	}

	// Validate minimum purchase
	if discount.MinPurchase != nil && req.PurchaseTotal < *discount.MinPurchase {
		return &models.ValidateDiscountResponse{
			Valid:          false,
			Message:        "Purchase amount does not meet minimum requirement",
			DiscountAmount: 0,
			DiscountCode:   discount,
		}, nil
	}

	// Calculate discount
	var discountAmount float64
	if discount.Type == models.DiscountPercentage {
		discountAmount = req.PurchaseTotal * (discount.Value / 100)
	} else {
		discountAmount = discount.Value
	}

	// Don't allow discount greater than total
	if discountAmount > req.PurchaseTotal {
		discountAmount = req.PurchaseTotal
	}

	return &models.ValidateDiscountResponse{
		Valid:          true,
		Message:        "Discount code is valid",
		DiscountAmount: discountAmount,
		DiscountCode:   discount,
	}, nil
}

// ApplyDiscountCode marks a code as used (increments counter)
func (s *DiscountService) ApplyDiscountCode(ctx context.Context, id uuid.UUID) error {
	return s.discountRepo.IncrementUsedCount(ctx, id)
}
