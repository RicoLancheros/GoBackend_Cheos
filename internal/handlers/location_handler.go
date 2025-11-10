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

type LocationHandler struct {
	locationService *services.LocationService
}

func NewLocationHandler(locationService *services.LocationService) *LocationHandler {
	return &LocationHandler{
		locationService: locationService,
	}
}

// CreateLocation creates a new location (admin only)
func (h *LocationHandler) CreateLocation(c *gin.Context) {
	var req models.CreateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid data", err.Error())
		return
	}

	location, err := h.locationService.CreateLocation(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error creating location", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Location created successfully", location)
}

// GetLocation gets a location by ID (public)
func (h *LocationHandler) GetLocation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid location ID", err.Error())
		return
	}

	location, err := h.locationService.GetLocationByID(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Location not found", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Location retrieved", location)
}

// GetAllLocations gets all locations with pagination (admin only)
func (h *LocationHandler) GetAllLocations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	locations, err := h.locationService.GetAllLocations(c.Request.Context(), page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error getting locations", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Locations retrieved", locations)
}

// GetActiveLocations gets all active locations (public)
func (h *LocationHandler) GetActiveLocations(c *gin.Context) {
	locations, err := h.locationService.GetActiveLocations(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error getting active locations", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Active locations retrieved", locations)
}

// UpdateLocation updates a location (admin only)
func (h *LocationHandler) UpdateLocation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid location ID", err.Error())
		return
	}

	var req models.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid data", err.Error())
		return
	}

	location, err := h.locationService.UpdateLocation(c.Request.Context(), id, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error updating location", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Location updated", location)
}

// DeleteLocation soft deletes a location (admin only)
func (h *LocationHandler) DeleteLocation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid location ID", err.Error())
		return
	}

	if err := h.locationService.DeleteLocation(c.Request.Context(), id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error deleting location", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Location deleted", nil)
}
