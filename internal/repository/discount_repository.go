package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/cheoscafe/backend/internal/database"
	"github.com/cheoscafe/backend/internal/models"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DiscountRepository struct {
	firebase *database.FirebaseClient
}

func NewDiscountRepository(firebase *database.FirebaseClient) *DiscountRepository {
	return &DiscountRepository{
		firebase: firebase,
	}
}

// Create creates a new discount code
func (r *DiscountRepository) Create(ctx context.Context, discount *models.DiscountCode) error {
	if discount.ID == uuid.Nil {
		discount.ID = uuid.New()
	}

	now := time.Now()
	discount.CreatedAt = now
	discount.UpdatedAt = now
	discount.UsedCount = 0

	// Convert code to uppercase for consistency
	discount.Code = strings.ToUpper(discount.Code)

	_, err := r.firebase.Collection("discount_codes").Doc(discount.ID.String()).Set(ctx, discount)
	if err != nil {
		return err
	}

	return nil
}

// GetByID gets a discount code by ID
func (r *DiscountRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.DiscountCode, error) {
	doc, err := r.firebase.Collection("discount_codes").Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errors.New("discount code not found")
		}
		return nil, err
	}

	var discount models.DiscountCode
	if err := doc.DataTo(&discount); err != nil {
		return nil, err
	}

	discount.ID = id
	return &discount, nil
}

// GetByCode gets a discount code by its code string
func (r *DiscountRepository) GetByCode(ctx context.Context, code string) (*models.DiscountCode, error) {
	code = strings.ToUpper(code)

	iter := r.firebase.Collection("discount_codes").
		Where("code", "==", code).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, errors.New("discount code not found")
	}
	if err != nil {
		return nil, err
	}

	var discount models.DiscountCode
	if err := doc.DataTo(&discount); err != nil {
		return nil, err
	}

	discountID, err := uuid.Parse(doc.Ref.ID)
	if err != nil {
		return nil, err
	}
	discount.ID = discountID

	return &discount, nil
}

// GetAll gets all discount codes with pagination
func (r *DiscountRepository) GetAll(ctx context.Context, limit int, offset int) ([]*models.DiscountCode, error) {
	query := r.firebase.Collection("discount_codes").
		Limit(limit).
		Offset(offset)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var discounts []*models.DiscountCode
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var discount models.DiscountCode
		if err := doc.DataTo(&discount); err != nil {
			continue
		}

		discountID, err := uuid.Parse(doc.Ref.ID)
		if err != nil {
			continue
		}
		discount.ID = discountID

		discounts = append(discounts, &discount)
	}

	return discounts, nil
}

// GetActive gets active discount codes
func (r *DiscountRepository) GetActive(ctx context.Context, limit int) ([]*models.DiscountCode, error) {
	query := r.firebase.Collection("discount_codes").
		Where("is_active", "==", true).
		Limit(limit)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var discounts []*models.DiscountCode
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var discount models.DiscountCode
		if err := doc.DataTo(&discount); err != nil {
			continue
		}

		discountID, err := uuid.Parse(doc.Ref.ID)
		if err != nil {
			continue
		}
		discount.ID = discountID

		discounts = append(discounts, &discount)
	}

	return discounts, nil
}

// Update updates a discount code
func (r *DiscountRepository) Update(ctx context.Context, discount *models.DiscountCode) error {
	discount.UpdatedAt = time.Now()

	// Convert code to uppercase
	if discount.Code != "" {
		discount.Code = strings.ToUpper(discount.Code)
	}

	_, err := r.firebase.Collection("discount_codes").Doc(discount.ID.String()).Set(ctx, discount)
	if err != nil {
		return err
	}

	return nil
}

// Delete elimina un código de descuento físicamente (hard delete)
func (r *DiscountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.firebase.Collection("discount_codes").Doc(id.String()).Delete(ctx)
	if err != nil {
		return err
	}

	return nil
}

// IncrementUsedCount increments the used count
func (r *DiscountRepository) IncrementUsedCount(ctx context.Context, id uuid.UUID) error {
	return r.firebase.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		docRef := r.firebase.Collection("discount_codes").Doc(id.String())
		doc, err := tx.Get(docRef)
		if err != nil {
			return err
		}

		var discount models.DiscountCode
		if err := doc.DataTo(&discount); err != nil {
			return err
		}

		newUsedCount := discount.UsedCount + 1

		return tx.Update(docRef, []firestore.Update{
			{Path: "used_count", Value: newUsedCount},
			{Path: "updated_at", Value: time.Now()},
		})
	})
}

// CountDiscountCodes counts total discount codes
func (r *DiscountRepository) CountDiscountCodes(ctx context.Context) (int64, error) {
	iter := r.firebase.Collection("discount_codes").Documents(ctx)
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
