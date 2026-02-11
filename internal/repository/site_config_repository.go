package repository

import (
	"context"
	"time"

	"github.com/cheoscafe/backend/internal/database"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SiteConfigRepository struct {
	firebase *database.FirebaseClient
}

func NewSiteConfigRepository(firebase *database.FirebaseClient) *SiteConfigRepository {
	return &SiteConfigRepository{firebase: firebase}
}

// GetCarouselImages obtiene las URLs del carrusel
func (r *SiteConfigRepository) GetCarouselImages(ctx context.Context) ([]string, error) {
	doc, err := r.firebase.Collection("site_config").Doc("carousel").Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return []string{}, nil
		}
		return nil, err
	}

	data := doc.Data()
	images, ok := data["images"]
	if !ok {
		return []string{}, nil
	}

	// Convertir []interface{} a []string
	rawImages, ok := images.([]interface{})
	if !ok {
		return []string{}, nil
	}

	result := make([]string, 0, len(rawImages))
	for _, img := range rawImages {
		if str, ok := img.(string); ok {
			result = append(result, str)
		}
	}

	return result, nil
}

// SetCarouselImages guarda las URLs del carrusel
func (r *SiteConfigRepository) SetCarouselImages(ctx context.Context, images []string) error {
	_, err := r.firebase.Collection("site_config").Doc("carousel").Set(ctx, map[string]interface{}{
		"images":     images,
		"updated_at": time.Now(),
	})
	return err
}
