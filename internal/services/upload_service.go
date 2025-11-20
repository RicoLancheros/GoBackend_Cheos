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
		return nil, fmt.Errorf("error inicializando cloudinary: %w", err)
	}

	return &UploadService{
		cfg:        cfg,
		cloudinary: cld,
	}, nil
}

// UploadImage sube una imagen a Cloudinary y devuelve la URL pública
func (s *UploadService) UploadImage(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	// Validar tipo de archivo
	if !s.isValidImageType(fileHeader.Filename) {
		return "", errors.New("tipo de archivo inválido. Solo se permiten imágenes (jpg, jpeg, png, webp, gif)")
	}

	// Validar tamaño (máximo 10MB)
	if fileHeader.Size > 10*1024*1024 {
		return "", errors.New("el archivo es demasiado grande. Tamaño máximo: 10MB")
	}

	// Generar public_id único para Cloudinary
	timestamp := time.Now().Unix()
	baseFilename := strings.TrimSuffix(fileHeader.Filename, filepath.Ext(fileHeader.Filename))
	publicID := fmt.Sprintf("cheos-cafe/gallery/%d_%s", timestamp, s.sanitizeFilename(baseFilename))

	// Subir imagen a Cloudinary
	uploadParams := uploader.UploadParams{
		PublicID:     publicID,
		Folder:       "cheos-cafe/gallery",
		ResourceType: "image",
		Context: map[string]string{
			"uploaded_at":   time.Now().Format(time.RFC3339),
			"original_name": fileHeader.Filename,
		},
	}

	result, err := s.cloudinary.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", fmt.Errorf("error al subir imagen a cloudinary: %w", err)
	}

	// Retornar la URL segura de Cloudinary
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

// sanitizeFilename limpia el nombre del archivo para uso en Firebase Storage
func (s *UploadService) sanitizeFilename(filename string) string {
	// Reemplazar espacios y caracteres especiales
	name := strings.ReplaceAll(filename, " ", "_")
	name = strings.ReplaceAll(name, "#", "")
	name = strings.ReplaceAll(name, "?", "")
	name = strings.ReplaceAll(name, "&", "")
	name = strings.ToLower(name)
	return name
}
