package repository

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/cheoscafe/backend/internal/database"
	"github.com/cheoscafe/backend/internal/models"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

type LocationRepository struct {
	firebase *firestore.Client
}

func NewLocationRepository(firebase *database.FirebaseClient) *LocationRepository {
	return &LocationRepository{
		firebase: firebase.Firestore,
	}
}

// Create creates a new location
func (r *LocationRepository) Create(ctx context.Context, location *models.Location) error {
	if location.ID == uuid.Nil {
		location.ID = uuid.New()
	}

	now := time.Now()
	location.CreatedAt = now
	location.UpdatedAt = now

	_, err := r.firebase.Collection("locations").Doc(location.ID.String()).Set(ctx, location)
	return err
}

// GetByID gets a location by ID
func (r *LocationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Location, error) {
	doc, err := r.firebase.Collection("locations").Doc(id.String()).Get(ctx)
	if err != nil {
		return nil, err
	}

	var location models.Location
	if err := doc.DataTo(&location); err != nil {
		return nil, err
	}

	return &location, nil
}

// GetAll gets all locations with pagination
func (r *LocationRepository) GetAll(ctx context.Context, limit int, offset int) ([]*models.Location, error) {
	query := r.firebase.Collection("locations").
		OrderBy("created_at", firestore.Desc).
		Limit(limit).
		Offset(offset)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var locations []*models.Location
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var location models.Location
		if err := doc.DataTo(&location); err != nil {
			return nil, err
		}
		locations = append(locations, &location)
	}

	return locations, nil
}

// GetActive gets all active locations (for public display)
func (r *LocationRepository) GetActive(ctx context.Context) ([]*models.Location, error) {
	query := r.firebase.Collection("locations").
		Where("is_active", "==", true)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var locations []*models.Location
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var location models.Location
		if err := doc.DataTo(&location); err != nil {
			return nil, err
		}
		locations = append(locations, &location)
	}

	return locations, nil
}

// Update updates a location
func (r *LocationRepository) Update(ctx context.Context, id uuid.UUID, location *models.Location) error {
	location.UpdatedAt = time.Now()

	updates := []firestore.Update{}

	if location.Name != "" {
		updates = append(updates, firestore.Update{Path: "name", Value: location.Name})
	}
	if location.Address != "" {
		updates = append(updates, firestore.Update{Path: "address", Value: location.Address})
	}
	if location.City != "" {
		updates = append(updates, firestore.Update{Path: "city", Value: location.City})
	}
	if location.Department != "" {
		updates = append(updates, firestore.Update{Path: "department", Value: location.Department})
	}
	if location.Phone != "" {
		updates = append(updates, firestore.Update{Path: "phone", Value: location.Phone})
	}
	if location.Latitude != 0 {
		updates = append(updates, firestore.Update{Path: "latitude", Value: location.Latitude})
	}
	if location.Longitude != 0 {
		updates = append(updates, firestore.Update{Path: "longitude", Value: location.Longitude})
	}
	if location.Schedule != nil {
		updates = append(updates, firestore.Update{Path: "schedule", Value: location.Schedule})
	}

	updates = append(updates, firestore.Update{Path: "is_active", Value: location.IsActive})
	updates = append(updates, firestore.Update{Path: "updated_at", Value: location.UpdatedAt})

	_, err := r.firebase.Collection("locations").Doc(id.String()).Update(ctx, updates)
	return err
}

// Delete soft deletes a location (sets is_active to false)
func (r *LocationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.firebase.Collection("locations").Doc(id.String()).Update(ctx, []firestore.Update{
		{Path: "is_active", Value: false},
		{Path: "updated_at", Value: time.Now()},
	})
	return err
}

// Count returns the total number of locations
func (r *LocationRepository) Count(ctx context.Context) (int, error) {
	iter := r.firebase.Collection("locations").Documents(ctx)
	defer iter.Stop()

	count := 0
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, err
		}
		count++
	}

	return count, nil
}
