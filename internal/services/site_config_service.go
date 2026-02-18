package services

import (
	"context"
	"errors"

	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/repository"
)

type SiteConfigService struct {
	repo *repository.SiteConfigRepository
}

func NewSiteConfigService(repo *repository.SiteConfigRepository) *SiteConfigService {
	return &SiteConfigService{repo: repo}
}

// GetCarouselImages obtiene las imágenes del carrusel
func (s *SiteConfigService) GetCarouselImages(ctx context.Context) ([]string, error) {
	return s.repo.GetCarouselImages(ctx)
}

// SetCarouselImages guarda las imágenes del carrusel (máximo 6)
func (s *SiteConfigService) SetCarouselImages(ctx context.Context, images []string) error {
	if len(images) > 6 {
		return errors.New("máximo 6 imágenes en el carrusel")
	}
	return s.repo.SetCarouselImages(ctx, images)
}

// GetAboutUs obtiene la configuración de "Sobre Nosotros"
func (s *SiteConfigService) GetAboutUs(ctx context.Context) (*models.AboutUsConfig, error) {
	description, images, err := s.repo.GetAboutUs(ctx)
	if err != nil {
		return nil, err
	}
	return &models.AboutUsConfig{
		Description: description,
		Images:      images,
	}, nil
}

// SetAboutUs actualiza la configuración de "Sobre Nosotros"
func (s *SiteConfigService) SetAboutUs(ctx context.Context, description *string, images []string) error {
	// Obtener datos actuales para actualización parcial
	currentDesc, currentImages, err := s.repo.GetAboutUs(ctx)
	if err != nil {
		return err
	}

	newDesc := currentDesc
	if description != nil {
		newDesc = *description
	}

	newImages := currentImages
	if images != nil {
		if len(images) > 6 {
			return errors.New("máximo 6 imágenes en Sobre Nosotros")
		}
		newImages = images
	}

	return s.repo.SetAboutUs(ctx, newDesc, newImages)
}
