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

type UserRepository struct {
	firebase *database.FirebaseClient
}

func NewUserRepository(firebase *database.FirebaseClient) *UserRepository {
	return &UserRepository{
		firebase: firebase,
	}
}

// Create crea un nuevo usuario en Firestore
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	// Establecer timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Crear documento en Firestore
	_, err := r.firebase.Collection("users").Doc(user.ID.String()).Set(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

// GetByEmail obtiene un usuario por email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	// Buscar usuario por email
	iter := r.firebase.Collection("users").Where("email", "==", email).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, errors.New("usuario no encontrado")
	}
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}

	// Asignar el ID del documento
	userID, err := uuid.Parse(doc.Ref.ID)
	if err != nil {
		return nil, err
	}
	user.ID = userID

	return &user, nil
}

// GetByID obtiene un usuario por ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	doc, err := r.firebase.Collection("users").Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errors.New("usuario no encontrado")
		}
		return nil, err
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}

	user.ID = id
	return &user, nil
}

// Update actualiza un usuario
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	// Actualizar timestamp
	user.UpdatedAt = time.Now()

	// Actualizar solo campos específicos
	updates := []firestore.Update{
		{Path: "name", Value: user.Name},
		{Path: "phone", Value: user.Phone},
		{Path: "updated_at", Value: user.UpdatedAt},
	}

	_, err := r.firebase.Collection("users").Doc(user.ID.String()).Update(ctx, updates)
	if err != nil {
		return err
	}

	return nil
}

// EmailExists verifica si un email ya existe
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	iter := r.firebase.Collection("users").Where("email", "==", email).Limit(1).Documents(ctx)
	defer iter.Stop()

	_, err := iter.Next()
	if err == iterator.Done {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}