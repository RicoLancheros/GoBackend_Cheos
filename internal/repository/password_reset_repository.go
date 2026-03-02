package repository

import (
	"context"
	"errors"

	"github.com/cheoscafe/backend/internal/database"
	"github.com/cheoscafe/backend/internal/models"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

type PasswordResetRepository struct {
	firebase *database.FirebaseClient
}

func NewPasswordResetRepository(firebase *database.FirebaseClient) *PasswordResetRepository {
	return &PasswordResetRepository{
		firebase: firebase,
	}
}

// Create guarda un nuevo token de reset en Firestore
func (r *PasswordResetRepository) Create(ctx context.Context, reset *models.PasswordReset) error {
	if reset.ID == uuid.Nil {
		reset.ID = uuid.New()
	}

	_, err := r.firebase.Collection("password_resets").Doc(reset.ID.String()).Set(ctx, reset)
	if err != nil {
		return err
	}

	return nil
}

// GetByToken busca un token de reset por el valor del token JWT
func (r *PasswordResetRepository) GetByToken(ctx context.Context, token string) (*models.PasswordReset, error) {
	iter := r.firebase.Collection("password_resets").Where("token", "==", token).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, errors.New("token no encontrado")
	}
	if err != nil {
		return nil, err
	}

	var reset models.PasswordReset
	if err := doc.DataTo(&reset); err != nil {
		return nil, err
	}

	// Asignar el ID del documento
	resetID, err := uuid.Parse(doc.Ref.ID)
	if err != nil {
		return nil, err
	}
	reset.ID = resetID

	return &reset, nil
}

// DeleteByUserID elimina todos los tokens de reset previos de un usuario
func (r *PasswordResetRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	iter := r.firebase.Collection("password_resets").Where("user_id", "==", userID).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		_, err = doc.Ref.Delete(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete elimina un token de reset especifico por ID (hard delete)
func (r *PasswordResetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.firebase.Collection("password_resets").Doc(id.String()).Delete(ctx)
	if err != nil {
		return err
	}

	return nil
}
