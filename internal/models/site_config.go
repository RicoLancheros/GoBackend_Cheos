package models

import "time"

// SiteConfig almacena configuraciones globales del sitio
type SiteConfig struct {
	Key       string      `json:"key" firestore:"key"`
	Value     interface{} `json:"value" firestore:"value"`
	UpdatedAt time.Time   `json:"updated_at" firestore:"updated_at"`
}

// CarouselConfig almacena las URLs del carrusel
type CarouselConfig struct {
	Images []string `json:"images" firestore:"images"`
}

// UpdateCarouselRequest DTO para actualizar el carrusel
type UpdateCarouselRequest struct {
	Images []string `json:"images" validate:"required,max=6,dive,url"`
}
