package handlers

import (
	"net/http"
	"strconv"

	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/services"
	"github.com/cheoscafe/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService *services.ProductService
}

func NewProductHandler(productService *services.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// CreateProduct crea un nuevo producto (solo admin)
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	product, err := h.productService.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al crear producto", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Producto creado exitosamente", product)
}

// GetProduct obtiene un producto por ID
func (h *ProductHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")

	product, err := h.productService.GetProduct(c.Request.Context(), productID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Producto no encontrado", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Producto obtenido exitosamente", product)
}

// GetAllProducts obtiene todos los productos con paginación
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	// Parámetros de paginación
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	response, err := h.productService.GetAllProducts(c.Request.Context(), page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al obtener productos", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Productos obtenidos exitosamente", response)
}

// GetFeaturedProducts obtiene productos destacados
func (h *ProductHandler) GetFeaturedProducts(c *gin.Context) {
	products, err := h.productService.GetFeaturedProducts(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al obtener productos destacados", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Productos destacados obtenidos exitosamente", products)
}

// UpdateProduct actualiza un producto (solo admin)
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	productID := c.Param("id")

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	product, err := h.productService.UpdateProduct(c.Request.Context(), productID, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al actualizar producto", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Producto actualizado exitosamente", product)
}

// DeleteProduct elimina un producto (solo admin)
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	productID := c.Param("id")

	if err := h.productService.DeleteProduct(c.Request.Context(), productID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al eliminar producto", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Producto eliminado exitosamente", nil)
}

// UpdateStock actualiza el stock de un producto
func (h *ProductHandler) UpdateStock(c *gin.Context) {
	productID := c.Param("id")

	var req struct {
		Quantity int `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	if err := h.productService.UpdateStock(c.Request.Context(), productID, req.Quantity); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al actualizar stock", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Stock actualizado exitosamente", nil)
}

// SearchProducts busca productos
func (h *ProductHandler) SearchProducts(c *gin.Context) {
	searchTerm := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	products, err := h.productService.SearchProducts(c.Request.Context(), searchTerm, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Error al buscar productos", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Búsqueda completada", products)
}
