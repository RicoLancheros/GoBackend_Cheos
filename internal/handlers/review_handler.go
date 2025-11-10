package handlers

import (
	"net/http"
	"strconv"

	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/services"
	"github.com/cheoscafe/backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReviewHandler struct {
	reviewService *services.ReviewService
}

func NewReviewHandler(reviewService *services.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
	}
}

// CreateReview creates a new review (public or authenticated)
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	var req models.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid data", err.Error())
		return
	}

	// Check if user is authenticated
	var userID *uuid.UUID
	if userIDValue, exists := c.Get("user_id"); exists {
		if id, ok := userIDValue.(string); ok {
			parsedID, err := uuid.Parse(id)
			if err == nil {
				userID = &parsedID
			}
		}
	}

	review, err := h.reviewService.CreateReview(c.Request.Context(), &req, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error creating review", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Review created successfully (pending approval)", review)
}

// GetReview gets a review by ID (admin only)
func (h *ReviewHandler) GetReview(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid review ID", err.Error())
		return
	}

	review, err := h.reviewService.GetReviewByID(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Review not found", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Review retrieved", review)
}

// GetAllReviews gets all reviews with pagination (admin only)
func (h *ReviewHandler) GetAllReviews(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	reviews, err := h.reviewService.GetAllReviews(c.Request.Context(), page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error getting reviews", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Reviews retrieved", reviews)
}

// GetProductReviews gets reviews for a specific product (public)
func (h *ReviewHandler) GetProductReviews(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	reviews, err := h.reviewService.GetProductReviews(c.Request.Context(), productID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error getting product reviews", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Product reviews retrieved", reviews)
}

// UpdateReview updates a review (admin only)
func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid review ID", err.Error())
		return
	}

	var req models.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid data", err.Error())
		return
	}

	review, err := h.reviewService.UpdateReview(c.Request.Context(), id, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error updating review", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Review updated", review)
}

// DeleteReview deletes a review (admin only)
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid review ID", err.Error())
		return
	}

	if err := h.reviewService.DeleteReview(c.Request.Context(), id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error deleting review", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Review deleted", nil)
}
