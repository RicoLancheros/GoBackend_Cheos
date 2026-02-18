package models

import (
	"time"

	"github.com/google/uuid"
)

// ImageType define los tipos de imágenes disponibles
type ImageType string

const (
	ImageTypeCarousel   ImageType = "CAROUSEL"   // Para carruseles/sliders
	ImageTypeProduct    ImageType = "PRODUCT"    // Para productos
	ImageTypeBackground ImageType = "BACKGROUND" // Fondos
	ImageTypeGeneral    ImageType = "GENERAL"    // Uso general
	ImageTypeAboutUs    ImageType = "ABOUT_US"   // Para sección About Us
)

type GalleryImage struct {
	ID          uuid.UUID `json:"id" firestore:"id"`
	URL         string    `json:"url" firestore:"url"`
	Title       string    `json:"title" firestore:"title"`
	Description string    `json:"description" firestore:"description"`
	ImageType   ImageType `json:"image_type" firestore:"image_type"`
	Tags        []string  `json:"tags" firestore:"tags"`         // Etiquetas para búsqueda
	IsActive    bool      `json:"is_active" firestore:"is_active"`
	CreatedAt   time.Time `json:"created_at" firestore:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" firestore:"updated_at"`
}

// DTOs

type CreateGalleryImageRequest struct {
	URL         string    `json:"url" validate:"required,url"`
	Title       string    `json:"title" validate:"omitempty,max=100"`
	Description string    `json:"description" validate:"omitempty,max=500"`
	ImageType   ImageType `json:"image_type" validate:"required,oneof=CAROUSEL PRODUCT BACKGROUND GENERAL ABOUT_US"`
	Tags        []string  `json:"tags" validate:"omitempty,dive,max=50"`
}

type UpdateGalleryImageRequest struct {
	URL         *string    `json:"url" validate:"omitempty,url"`
	Title       *string    `json:"title" validate:"omitempty,max=100"`
	Description *string    `json:"description" validate:"omitempty,max=500"`
	ImageType   *ImageType `json:"image_type" validate:"omitempty,oneof=CAROUSEL PRODUCT BACKGROUND GENERAL ABOUT_US"`
	Tags        []string   `json:"tags" validate:"omitempty,dive,max=50"`
	IsActive    *bool      `json:"is_active" validate:"omitempty"`
}
