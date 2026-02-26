package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/cheoscafe/backend/internal/services"
	"github.com/cheoscafe/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dashboardService *services.DashboardService
}

func NewDashboardHandler(dashboardService *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

// --- Helpers de parsing ---

func parseYearMonth(c *gin.Context) (int, int, error) {
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2020 {
		return 0, 0, errors.New("parametro 'year' invalido")
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		return 0, 0, errors.New("parametro 'month' invalido (1-12)")
	}

	return year, month, nil
}

func parseYear(c *gin.Context) (int, error) {
	yearStr := c.Query("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2020 {
		return 0, errors.New("parametro 'year' invalido")
	}
	return year, nil
}

// --- Endpoints ---

// GetSalesMonthly obtiene ventas mensuales
// GET /api/v1/dashboard/sales/monthly?year=2026&month=2
func (h *DashboardHandler) GetSalesMonthly(c *gin.Context) {
	year, month, err := parseYearMonth(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Parametros invalidos", err.Error())
		return
	}

	data, err := h.dashboardService.GetSalesMonthly(c.Request.Context(), year, month)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al obtener ventas mensuales", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ventas mensuales obtenidas", data)
}

// GetSalesYearly obtiene ventas anuales
// GET /api/v1/dashboard/sales/yearly?year=2026
func (h *DashboardHandler) GetSalesYearly(c *gin.Context) {
	year, err := parseYear(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Parametros invalidos", err.Error())
		return
	}

	data, err := h.dashboardService.GetSalesYearly(c.Request.Context(), year)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al obtener ventas anuales", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ventas anuales obtenidas", data)
}

// GetBuyersMonthly obtiene estadisticas de compradores mensuales
// GET /api/v1/dashboard/buyers/monthly?year=2026&month=2
func (h *DashboardHandler) GetBuyersMonthly(c *gin.Context) {
	year, month, err := parseYearMonth(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Parametros invalidos", err.Error())
		return
	}

	data, err := h.dashboardService.GetBuyersMonthly(c.Request.Context(), year, month)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al obtener estadisticas de compradores mensuales", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Estadisticas de compradores mensuales obtenidas", data)
}

// GetBuyersYearly obtiene estadisticas de compradores anuales
// GET /api/v1/dashboard/buyers/yearly?year=2026
func (h *DashboardHandler) GetBuyersYearly(c *gin.Context) {
	year, err := parseYear(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Parametros invalidos", err.Error())
		return
	}

	data, err := h.dashboardService.GetBuyersYearly(c.Request.Context(), year)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al obtener estadisticas de compradores anuales", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Estadisticas de compradores anuales obtenidas", data)
}

// GetTopProductsMonthly obtiene top productos mensuales
// GET /api/v1/dashboard/products/monthly?year=2026&month=2
func (h *DashboardHandler) GetTopProductsMonthly(c *gin.Context) {
	year, month, err := parseYearMonth(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Parametros invalidos", err.Error())
		return
	}

	data, err := h.dashboardService.GetTopProductsMonthly(c.Request.Context(), year, month)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al obtener top productos mensuales", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Top productos mensuales obtenidos", data)
}

// GetTopProductsYearly obtiene top productos anuales
// GET /api/v1/dashboard/products/yearly?year=2026
func (h *DashboardHandler) GetTopProductsYearly(c *gin.Context) {
	year, err := parseYear(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Parametros invalidos", err.Error())
		return
	}

	data, err := h.dashboardService.GetTopProductsYearly(c.Request.Context(), year)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al obtener top productos anuales", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Top productos anuales obtenidos", data)
}

// GetSummary obtiene el resumen consolidado del dashboard
// GET /api/v1/dashboard/summary
func (h *DashboardHandler) GetSummary(c *gin.Context) {
	data, err := h.dashboardService.GetSummary(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al obtener resumen del dashboard", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Resumen del dashboard obtenido", data)
}

// RecalculateMonth recalcula metricas del mes y ano desde cero
// POST /api/v1/dashboard/recalculate?year=2026&month=2
func (h *DashboardHandler) RecalculateMonth(c *gin.Context) {
	year, month, err := parseYearMonth(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Parametros invalidos", err.Error())
		return
	}

	if err := h.dashboardService.RecalculateMonth(c.Request.Context(), year, month); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al recalcular metricas", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Metricas recalculadas exitosamente", nil)
}
