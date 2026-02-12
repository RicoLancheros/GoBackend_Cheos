package repository

import (
	"context"
	"time"

	"github.com/cheoscafe/backend/internal/database"
	"github.com/cheoscafe/backend/internal/models"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CartRepository struct {
	firebase *database.FirebaseClient
}

func NewCartRepository(firebase *database.FirebaseClient) *CartRepository {
	return &CartRepository{firebase: firebase}
}

// GetByUserID obtiene el carrito de un usuario. Si no existe, retorna carrito vacío.
func (r *CartRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Cart, error) {
	doc, err := r.firebase.Collection("carts").Doc(userID.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return &models.Cart{
				UserID:    userID,
				Items:     []models.CartItem{},
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, err
	}

	var cart models.Cart
	if err := doc.DataTo(&cart); err != nil {
		return nil, err
	}

	cart.UserID = userID
	if cart.Items == nil {
		cart.Items = []models.CartItem{}
	}

	return &cart, nil
}

// Save guarda/sobreescribe el carrito completo
func (r *CartRepository) Save(ctx context.Context, cart *models.Cart) error {
	cart.UpdatedAt = time.Now()
	_, err := r.firebase.Collection("carts").Doc(cart.UserID.String()).Set(ctx, cart)
	return err
}

// Delete elimina el carrito de un usuario
func (r *CartRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	_, err := r.firebase.Collection("carts").Doc(userID.String()).Delete(ctx)
	return err
}
