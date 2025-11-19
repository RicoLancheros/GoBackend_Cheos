package repository

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/cheoscafe/backend/internal/database"
	"github.com/cheoscafe/backend/internal/models"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GalleryRepository struct {
	firebase *database.FirebaseClient
}

func NewGalleryRepository(firebase *database.FirebaseClient) *GalleryRepository {
	return &GalleryRepository{
		firebase: firebase,
	}
}

// Create crea una nueva imagen en la galería
func (r *GalleryRepository) Create(ctx context.Context, image *models.GalleryImage) error {
	if image.ID == uuid.Nil {
		image.ID = uuid.New()
	}

	now := time.Now()
	image.CreatedAt = now
	image.UpdatedAt = now
	image.IsActive = true

	_, err := r.firebase.Collection("gallery").Doc(image.ID.String()).Set(ctx, image)
	if err != nil {
		return err
	}

	return nil
}

// GetByID obtiene una imagen por ID
func (r *GalleryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.GalleryImage, error) {
	doc, err := r.firebase.Collection("gallery").Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errors.New("imagen no encontrada")
		}
		return nil, err
	}

	var image models.GalleryImage
	if err := doc.DataTo(&image); err != nil {
		return nil, err
	}

	image.ID = id
	return &image, nil
}

// GetAll obtiene todas las imágenes de la galería
func (r *GalleryRepository) GetAll(ctx context.Context) ([]*models.GalleryImage, error) {
	query := r.firebase.Collection("gallery").OrderBy("created_at", firestore.Desc)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var images []*models.GalleryImage
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var image models.GalleryImage
		if err := doc.DataTo(&image); err != nil {
			continue
		}

		imageID, err := uuid.Parse(doc.Ref.ID)
		if err != nil {
			continue
		}
		image.ID = imageID

		images = append(images, &image)
	}

	return images, nil
}

// GetByType obtiene imágenes por tipo
func (r *GalleryRepository) GetByType(ctx context.Context, imageType models.ImageType) ([]*models.GalleryImage, error) {
	query := r.firebase.Collection("gallery").
		Where("image_type", "==", imageType).
		Where("is_active", "==", true)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var images []*models.GalleryImage
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var image models.GalleryImage
		if err := doc.DataTo(&image); err != nil {
			continue
		}

		imageID, err := uuid.Parse(doc.Ref.ID)
		if err != nil {
			continue
		}
		image.ID = imageID

		images = append(images, &image)
	}

	return images, nil
}

// GetActive obtiene solo las imágenes activas
func (r *GalleryRepository) GetActive(ctx context.Context) ([]*models.GalleryImage, error) {
	query := r.firebase.Collection("gallery").
		Where("is_active", "==", true)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var images []*models.GalleryImage
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var image models.GalleryImage
		if err := doc.DataTo(&image); err != nil {
			continue
		}

		imageID, err := uuid.Parse(doc.Ref.ID)
		if err != nil {
			continue
		}
		image.ID = imageID

		images = append(images, &image)
	}

	return images, nil
}

// Update actualiza una imagen existente
func (r *GalleryRepository) Update(ctx context.Context, image *models.GalleryImage) error {
	image.UpdatedAt = time.Now()

	_, err := r.firebase.Collection("gallery").Doc(image.ID.String()).Set(ctx, image)
	if err != nil {
		return err
	}

	return nil
}

// Delete elimina una imagen físicamente (hard delete)
func (r *GalleryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.firebase.Collection("gallery").Doc(id.String()).Delete(ctx)
	if err != nil {
		return err
	}

	return nil
}

// HardDelete elimina permanentemente una imagen
func (r *GalleryRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.firebase.Collection("gallery").Doc(id.String()).Delete(ctx)
	if err != nil {
		return err
	}

	return nil
}
