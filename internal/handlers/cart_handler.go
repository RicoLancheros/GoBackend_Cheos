package handlers

import (
	"net/http"

	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/services"
	"github.com/cheoscafe/backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CartHandler struct {
	cartService *services.CartService
}

func NewCartHandler(cartService *services.CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

// getUserID extrae el user_id del contexto (puesto por AuthMiddleware)
func (h *CartHandler) getUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "No autenticado", nil)
		return uuid.Nil, false
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID de usuario inválido", nil)
		return uuid.Nil, false
	}

	return userID, true
}

// GetCart obtiene el carrito del usuario autenticado
func (h *CartHandler) GetCart(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	cart, err := h.cartService.GetCart(c.Request.Context(), userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al obtener carrito", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Carrito obtenido", cart)
}

// AddItem agrega un producto al carrito
func (h *CartHandler) AddItem(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var req models.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.FormatValidationErrors(err))
		return
	}

	cart, err := h.cartService.AddItem(c.Request.Context(), userID, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Error al agregar al carrito", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Producto agregado al carrito", cart)
}

// UpdateItemQuantity actualiza la cantidad de un producto en el carrito
func (h *CartHandler) UpdateItemQuantity(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	productIDStr := c.Param("productId")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID de producto inválido", err.Error())
		return
	}

	var req models.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.FormatValidationErrors(err))
		return
	}

	cart, err := h.cartService.UpdateItemQuantity(c.Request.Context(), userID, productID, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Error al actualizar cantidad", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cantidad actualizada", cart)
}

// RemoveItem elimina un producto del carrito
func (h *CartHandler) RemoveItem(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	productIDStr := c.Param("productId")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID de producto inválido", err.Error())
		return
	}

	cart, err := h.cartService.RemoveItem(c.Request.Context(), userID, productID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al eliminar del carrito", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Producto eliminado del carrito", cart)
}

// ClearCart vacía el carrito completo
func (h *CartHandler) ClearCart(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	if err := h.cartService.ClearCart(c.Request.Context(), userID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al vaciar carrito", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Carrito vaciado", nil)
}

// SyncCart fusiona el carrito de invitado con el guardado
func (h *CartHandler) SyncCart(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var req models.SyncCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.FormatValidationErrors(err))
		return
	}

	cart, err := h.cartService.SyncCart(c.Request.Context(), userID, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al sincronizar carrito", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Carrito sincronizado", cart)
}
