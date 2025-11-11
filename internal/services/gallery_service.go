package services

import (
	"context"
	"errors"

	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/repository"
	"github.com/google/uuid"
)

type GalleryService struct {
	galleryRepo *repository.GalleryRepository
}

func NewGalleryService(galleryRepo *repository.GalleryRepository) *GalleryService {
	return &GalleryService{
		galleryRepo: galleryRepo,
	}
}

// CreateImage crea una nueva imagen en la galería
func (s *GalleryService) CreateImage(ctx context.Context, req *models.CreateGalleryImageRequest) (*models.GalleryImage, error) {
	image := &models.GalleryImage{
		URL:         req.URL,
		Title:       req.Title,
		Description: req.Description,
		ImageType:   req.ImageType,
		Tags:        req.Tags,
		IsActive:    true,
	}

	err := s.galleryRepo.Create(ctx, image)
	if err != nil {
		return nil, err
	}

	return image, nil
}

// GetImage obtiene una imagen por ID
func (s *GalleryService) GetImage(ctx context.Context, id string) (*models.GalleryImage, error) {
	imageID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID de imagen inválido")
	}

	return s.galleryRepo.GetByID(ctx, imageID)
}

// GetAllImages obtiene todas las imágenes
func (s *GalleryService) GetAllImages(ctx context.Context) ([]*models.GalleryImage, error) {
	return s.galleryRepo.GetAll(ctx)
}

// GetActiveImages obtiene solo las imágenes activas
func (s *GalleryService) GetActiveImages(ctx context.Context) ([]*models.GalleryImage, error) {
	return s.galleryRepo.GetActive(ctx)
}

// GetImagesByType obtiene imágenes por tipo
func (s *GalleryService) GetImagesByType(ctx context.Context, imageType models.ImageType) ([]*models.GalleryImage, error) {
	return s.galleryRepo.GetByType(ctx, imageType)
}

// UpdateImage actualiza una imagen existente
func (s *GalleryService) UpdateImage(ctx context.Context, id string, req *models.UpdateGalleryImageRequest) (*models.GalleryImage, error) {
	imageID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID de imagen inválido")
	}

	image, err := s.galleryRepo.GetByID(ctx, imageID)
	if err != nil {
		return nil, err
	}

	// Actualizar campos si se proporcionan
	if req.URL != nil {
		image.URL = *req.URL
	}
	if req.Title != nil {
		image.Title = *req.Title
	}
	if req.Description != nil {
		image.Description = *req.Description
	}
	if req.ImageType != nil {
		image.ImageType = *req.ImageType
	}
	if req.Tags != nil {
		image.Tags = req.Tags
	}
	if req.IsActive != nil {
		image.IsActive = *req.IsActive
	}

	err = s.galleryRepo.Update(ctx, image)
	if err != nil {
		return nil, err
	}

	return image, nil
}

// DeleteImage elimina una imagen (soft delete)
func (s *GalleryService) DeleteImage(ctx context.Context, id string) error {
	imageID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("ID de imagen inválido")
	}

	return s.galleryRepo.Delete(ctx, imageID)
}

// HardDeleteImage elimina permanentemente una imagen
func (s *GalleryService) HardDeleteImage(ctx context.Context, id string) error {
	imageID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("ID de imagen inválido")
	}

	return s.galleryRepo.HardDelete(ctx, imageID)
}
