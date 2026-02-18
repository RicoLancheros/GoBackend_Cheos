package handlers

import (
	"net/http"

	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/services"
	"github.com/cheoscafe/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type SiteConfigHandler struct {
	service *services.SiteConfigService
}

func NewSiteConfigHandler(service *services.SiteConfigService) *SiteConfigHandler {
	return &SiteConfigHandler{service: service}
}

// GetCarousel obtiene las imágenes del carrusel (público)
func (h *SiteConfigHandler) GetCarousel(c *gin.Context) {
	images, err := h.service.GetCarouselImages(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al obtener carrusel", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Carrusel obtenido exitosamente", gin.H{
		"images": images,
	})
}

// UpdateCarousel actualiza las imágenes del carrusel (solo admin)
func (h *SiteConfigHandler) UpdateCarousel(c *gin.Context) {
	var req models.UpdateCarouselRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	if err := h.service.SetCarouselImages(c.Request.Context(), req.Images); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Error al actualizar carrusel", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Carrusel actualizado exitosamente", gin.H{
		"images": req.Images,
	})
}

// GetAboutUs obtiene la configuración de "Sobre Nosotros" (público)
func (h *SiteConfigHandler) GetAboutUs(c *gin.Context) {
	aboutUs, err := h.service.GetAboutUs(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al obtener Sobre Nosotros", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sobre Nosotros obtenido exitosamente", aboutUs)
}

// UpdateAboutUs actualiza la configuración de "Sobre Nosotros" (solo admin)
func (h *SiteConfigHandler) UpdateAboutUs(c *gin.Context) {
	var req models.UpdateAboutUsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.FormatValidationErrors(err))
		return
	}

	if err := h.service.SetAboutUs(c.Request.Context(), req.Description, req.Images); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Error al actualizar Sobre Nosotros", err.Error())
		return
	}

	// Obtener los datos actualizados para retornar
	aboutUs, _ := h.service.GetAboutUs(c.Request.Context())
	utils.SuccessResponse(c, http.StatusOK, "Sobre Nosotros actualizado exitosamente", aboutUs)
}
