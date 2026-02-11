package services

import (
	"context"
	"errors"

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
