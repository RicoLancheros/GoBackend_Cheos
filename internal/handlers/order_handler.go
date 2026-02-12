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

type OrderHandler struct {
	orderService *services.OrderService
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// CreateOrder crea una nueva orden
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Obtener userID del contexto si está autenticado
	var userID *uuid.UUID
	if userIDInterface, exists := c.Get("user_id"); exists {
		if id, ok := userIDInterface.(uuid.UUID); ok {
			userID = &id
		}
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), &req, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al crear orden", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Orden creada exitosamente", order)
}

// GetOrder obtiene una orden por ID
func (h *OrderHandler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID de orden inválido", err.Error())
		return
	}

	order, err := h.orderService.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Orden no encontrada", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Orden obtenida", order)
}

// GetOrderByNumber obtiene una orden por número de orden
func (h *OrderHandler) GetOrderByNumber(c *gin.Context) {
	orderNumber := c.Param("number")

	order, err := h.orderService.GetOrderByOrderNumber(c.Request.Context(), orderNumber)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Orden no encontrada", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Orden obtenida", order)
}

// GetUserOrders obtiene las órdenes del usuario autenticado
func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "No autenticado", "user_id not found in context")
		return
	}

	// El middleware auth guarda el user_id como uuid.UUID, no como string
	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID de usuario inválido", "invalid user_id type")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	orders, err := h.orderService.GetUserOrders(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al obtener órdenes", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Órdenes obtenidas", orders)
}

// GetAllOrders obtiene todas las órdenes (solo admin)
func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	orders, err := h.orderService.GetAllOrders(c.Request.Context(), page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al obtener órdenes", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Órdenes obtenidas", orders)
}

// UpdateOrderStatus actualiza el estado de una orden (solo admin)
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID de orden inválido", err.Error())
		return
	}

	var req models.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	order, err := h.orderService.UpdateOrderStatus(c.Request.Context(), id, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al actualizar estado", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Estado de orden actualizado", order)
}

// UpdatePaymentStatus actualiza el estado de pago (solo admin)
func (h *OrderHandler) UpdatePaymentStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID de orden inválido", err.Error())
		return
	}

	var req models.UpdatePaymentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	order, err := h.orderService.UpdatePaymentStatus(c.Request.Context(), id, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al actualizar estado de pago", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Estado de pago actualizado", order)
}
