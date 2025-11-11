package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/cheoscafe/backend/internal/config"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type UploadService struct {
	cfg        *config.Config
	cloudinary *cloudinary.Cloudinary
}

func NewUploadService(cfg *config.Config) (*UploadService, error) {
	// Inicializar Cloudinary
	cld, err := cloudinary.NewFromParams(
		cfg.CloudinaryCloudName,
		cfg.CloudinaryAPIKey,
		cfg.CloudinaryAPISecret,
	)
	if err != nil {
		return nil, fmt.Errorf("error al inicializar Cloudinary: %w", err)
	}

	return &UploadService{
		cfg:        cfg,
		cloudinary: cld,
	}, nil
}

// UploadImage sube una imagen a Cloudinary y devuelve la URL
func (s *UploadService) UploadImage(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	// Validar tipo de archivo
	if !s.isValidImageType(fileHeader.Filename) {
		return "", errors.New("tipo de archivo inválido. Solo se permiten imágenes (jpg, jpeg, png, webp)")
	}

	// Validar tamaño (máximo 10MB)
	if fileHeader.Size > 10*1024*1024 {
		return "", errors.New("el archivo es demasiado grande. Tamaño máximo: 10MB")
	}

	// Generar nombre único para el archivo
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("gallery_%d_%s", timestamp, fileHeader.Filename)

	// Subir a Cloudinary
	uploadParams := uploader.UploadParams{
		PublicID:     s.generatePublicID(filename),
		Folder:       "cheos-gallery",
		ResourceType: "image",
	}

	result, err := s.cloudinary.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", fmt.Errorf("error al subir imagen a Cloudinary: %w", err)
	}

	// Devolver URL segura
	return result.SecureURL, nil
}

// isValidImageType verifica que el archivo sea una imagen válida
func (s *UploadService) isValidImageType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".gif":  true,
	}
	return validExtensions[ext]
}

// generatePublicID genera un ID público único para Cloudinary
func (s *UploadService) generatePublicID(filename string) string {
	// Eliminar extensión
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	// Reemplazar espacios y caracteres especiales
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ToLower(name)
	return name
}
