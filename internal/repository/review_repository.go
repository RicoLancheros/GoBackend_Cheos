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

type ReviewRepository struct {
	firebase *database.FirebaseClient
}

func NewReviewRepository(firebase *database.FirebaseClient) *ReviewRepository {
	return &ReviewRepository{
		firebase: firebase,
	}
}

// Create creates a new review
func (r *ReviewRepository) Create(ctx context.Context, review *models.Review) error {
	if review.ID == uuid.Nil {
		review.ID = uuid.New()
	}

	now := time.Now()
	review.CreatedAt = now
	review.UpdatedAt = now
	review.IsApproved = false // New reviews require approval

	_, err := r.firebase.Collection("reviews").Doc(review.ID.String()).Set(ctx, review)
	if err != nil {
		return err
	}

	return nil
}

// GetByID gets a review by ID
func (r *ReviewRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Review, error) {
	doc, err := r.firebase.Collection("reviews").Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errors.New("review not found")
		}
		return nil, err
	}

	var review models.Review
	if err := doc.DataTo(&review); err != nil {
		return nil, err
	}

	review.ID = id
	return &review, nil
}

// GetAll gets all reviews with pagination
func (r *ReviewRepository) GetAll(ctx context.Context, limit int, offset int) ([]*models.Review, error) {
	query := r.firebase.Collection("reviews").
		OrderBy("created_at", firestore.Desc).
		Limit(limit).
		Offset(offset)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var reviews []*models.Review
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var review models.Review
		if err := doc.DataTo(&review); err != nil {
			continue
		}

		reviewID, err := uuid.Parse(doc.Ref.ID)
		if err != nil {
			continue
		}
		review.ID = reviewID

		reviews = append(reviews, &review)
	}

	return reviews, nil
}

// GetByProductID gets all reviews for a product
func (r *ReviewRepository) GetByProductID(ctx context.Context, productID uuid.UUID, limit int, offset int) ([]*models.Review, error) {
	query := r.firebase.Collection("reviews").
		Where("product_id", "==", productID).
		Where("is_approved", "==", true). // Only return approved reviews to public
		OrderBy("created_at", firestore.Desc).
		Limit(limit).
		Offset(offset)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var reviews []*models.Review
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var review models.Review
		if err := doc.DataTo(&review); err != nil {
			continue
		}

		reviewID, err := uuid.Parse(doc.Ref.ID)
		if err != nil {
			continue
		}
		review.ID = reviewID

		reviews = append(reviews, &review)
	}

	return reviews, nil
}

// GetAllByProductID gets ALL reviews for a product (including unapproved) - Admin only
func (r *ReviewRepository) GetAllByProductID(ctx context.Context, productID uuid.UUID) ([]*models.Review, error) {
	query := r.firebase.Collection("reviews").
		Where("product_id", "==", productID).
		OrderBy("created_at", firestore.Desc)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var reviews []*models.Review
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var review models.Review
		if err := doc.DataTo(&review); err != nil {
			continue
		}

		reviewID, err := uuid.Parse(doc.Ref.ID)
		if err != nil {
			continue
		}
		review.ID = reviewID

		reviews = append(reviews, &review)
	}

	return reviews, nil
}

// Update updates a review
func (r *ReviewRepository) Update(ctx context.Context, review *models.Review) error {
	review.UpdatedAt = time.Now()

	_, err := r.firebase.Collection("reviews").Doc(review.ID.String()).Set(ctx, review)
	if err != nil {
		return err
	}

	return nil
}

// Delete deletes a review
func (r *ReviewRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.firebase.Collection("reviews").Doc(id.String()).Delete(ctx)
	if err != nil {
		return err
	}

	return nil
}

// CountReviews counts total reviews
func (r *ReviewRepository) CountReviews(ctx context.Context) (int64, error) {
	iter := r.firebase.Collection("reviews").Documents(ctx)
	defer iter.Stop()

	var count int64
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

// CountReviewsByProductID counts reviews for a specific product
func (r *ReviewRepository) CountReviewsByProductID(ctx context.Context, productID uuid.UUID) (int64, error) {
	iter := r.firebase.Collection("reviews").
		Where("product_id", "==", productID).
		Where("is_approved", "==", true).
		Documents(ctx)
	defer iter.Stop()

	var count int64
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
