package handlers

import (
	"net/http"

	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/services"
	"github.com/cheoscafe/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type GalleryHandler struct {
	galleryService *services.GalleryService
	uploadService  *services.UploadService
}

func NewGalleryHandler(galleryService *services.GalleryService, uploadService *services.UploadService) *GalleryHandler {
	return &GalleryHandler{
		galleryService: galleryService,
		uploadService:  uploadService,
	}
}

// CreateImage crea una nueva imagen en la galería
func (h *GalleryHandler) CreateImage(c *gin.Context) {
	var req models.CreateGalleryImageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Validar
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.FormatValidationErrors(err))
		return
	}

	image, err := h.galleryService.CreateImage(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Error al crear imagen", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Imagen creada exitosamente", image)
}

// GetImage obtiene una imagen por ID
func (h *GalleryHandler) GetImage(c *gin.Context) {
	id := c.Param("id")

	image, err := h.galleryService.GetImage(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Imagen no encontrada", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Imagen obtenida exitosamente", image)
}

// GetAllImages obtiene todas las imágenes (solo admin)
func (h *GalleryHandler) GetAllImages(c *gin.Context) {
	images, err := h.galleryService.GetAllImages(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al obtener imágenes", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Imágenes obtenidas exitosamente", images)
}

// GetActiveImages obtiene solo las imágenes activas (público)
func (h *GalleryHandler) GetActiveImages(c *gin.Context) {
	images, err := h.galleryService.GetActiveImages(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al obtener imágenes", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Imágenes obtenidas exitosamente", images)
}

// GetImagesByType obtiene imágenes por tipo (público)
func (h *GalleryHandler) GetImagesByType(c *gin.Context) {
	imageType := models.ImageType(c.Param("type"))

	// Validar que el tipo sea válido
	validTypes := map[models.ImageType]bool{
		models.ImageTypeCarousel:   true,
		models.ImageTypeProduct:    true,
		models.ImageTypeBackground: true,
		models.ImageTypeGeneral:    true,
		models.ImageTypeAboutUs:    true,
	}

	if !validTypes[imageType] {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tipo de imagen inválido", "Los tipos válidos son: CAROUSEL, PRODUCT, BACKGROUND, GENERAL, ABOUT_US")
		return
	}

	images, err := h.galleryService.GetImagesByType(c.Request.Context(), imageType)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al obtener imágenes", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Imágenes obtenidas exitosamente", images)
}

// UpdateImage actualiza una imagen existente
func (h *GalleryHandler) UpdateImage(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateGalleryImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Validar
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.FormatValidationErrors(err))
		return
	}

	image, err := h.galleryService.UpdateImage(c.Request.Context(), id, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Error al actualizar imagen", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Imagen actualizada exitosamente", image)
}

// DeleteImage elimina una imagen (soft delete)
func (h *GalleryHandler) DeleteImage(c *gin.Context) {
	id := c.Param("id")

	err := h.galleryService.DeleteImage(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Error al eliminar imagen", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Imagen eliminada exitosamente", nil)
}

// UploadImage sube una imagen a Cloudinary y la guarda en la galería
func (h *GalleryHandler) UploadImage(c *gin.Context) {
	// Obtener archivo del form-data
	file, fileHeader, err := c.Request.FormFile("image")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "No se proporcionó ningún archivo", err.Error())
		return
	}
	defer file.Close()

	// Subir imagen a Cloudinary
	imageURL, err := h.uploadService.UploadImage(c.Request.Context(), file, fileHeader)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Error al subir imagen", err.Error())
		return
	}

	// Obtener metadatos opcionales del form-data
	title := c.PostForm("title")
	description := c.PostForm("description")
	imageTypeStr := c.PostForm("image_type")

	// Validar y establecer tipo de imagen (por defecto GENERAL)
	imageType := models.ImageTypeGeneral
	if imageTypeStr != "" {
		imageType = models.ImageType(imageTypeStr)
	}

	// Crear entrada en la galería
	req := models.CreateGalleryImageRequest{
		URL:         imageURL,
		Title:       title,
		Description: description,
		ImageType:   imageType,
		Tags:        []string{},
	}

	image, err := h.galleryService.CreateImage(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Error al guardar imagen en galería", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Imagen subida y guardada exitosamente", image)
}
