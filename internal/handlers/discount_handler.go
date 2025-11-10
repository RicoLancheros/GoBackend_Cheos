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

type DiscountHandler struct {
	discountService *services.DiscountService
}

func NewDiscountHandler(discountService *services.DiscountService) *DiscountHandler {
	return &DiscountHandler{
		discountService: discountService,
	}
}

// CreateDiscountCode creates a new discount code (admin only)
func (h *DiscountHandler) CreateDiscountCode(c *gin.Context) {
	var req models.CreateDiscountCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid data", err.Error())
		return
	}

	discount, err := h.discountService.CreateDiscountCode(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error creating discount code", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Discount code created successfully", discount)
}

// GetDiscountCode gets a discount code by ID (admin only)
func (h *DiscountHandler) GetDiscountCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid discount code ID", err.Error())
		return
	}

	discount, err := h.discountService.GetDiscountCodeByID(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Discount code not found", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Discount code retrieved", discount)
}

// GetAllDiscountCodes gets all discount codes (admin only)
func (h *DiscountHandler) GetAllDiscountCodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	discounts, err := h.discountService.GetAllDiscountCodes(c.Request.Context(), page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error getting discount codes", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Discount codes retrieved", discounts)
}

// UpdateDiscountCode updates a discount code (admin only)
func (h *DiscountHandler) UpdateDiscountCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid discount code ID", err.Error())
		return
	}

	var req models.UpdateDiscountCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid data", err.Error())
		return
	}

	discount, err := h.discountService.UpdateDiscountCode(c.Request.Context(), id, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error updating discount code", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Discount code updated", discount)
}

// DeleteDiscountCode deletes a discount code (soft delete) (admin only)
func (h *DiscountHandler) DeleteDiscountCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid discount code ID", err.Error())
		return
	}

	if err := h.discountService.DeleteDiscountCode(c.Request.Context(), id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error deleting discount code", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Discount code deleted", nil)
}

// ValidateDiscountCode validates a discount code (public)
func (h *DiscountHandler) ValidateDiscountCode(c *gin.Context) {
	var req models.ValidateDiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid data", err.Error())
		return
	}

	result, err := h.discountService.ValidateDiscountCode(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error validating discount code", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, result.Message, result)
}
