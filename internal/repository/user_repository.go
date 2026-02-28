package repository

import (
	"context"
	"errors"
	"log"
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
	return &UserRepository{firebase: firebase}
}

func userFromDoc(doc *firestore.DocumentSnapshot, id uuid.UUID) (*models.User, error) {
	data := doc.Data()

	user := &models.User{}
	user.ID = id

	if v, ok := data["email"].(string); ok {
		user.Email = v
	}
	if v, ok := data["password"].(string); ok {
		user.Password = v
	}
	if v, ok := data["name"].(string); ok {
		user.Name = v
	}
	if v, ok := data["phone"].(string); ok {
		user.Phone = v
	}
	if v, ok := data["role"].(string); ok {
		user.Role = models.UserRole(v)
	}
	if v, ok := data["is_active"].(bool); ok {
		user.IsActive = v
	}
	if v, ok := data["created_at"]; ok {
		if t, ok := v.(time.Time); ok {
			user.CreatedAt = t
		}
	}
	if v, ok := data["updated_at"]; ok {
		if t, ok := v.(time.Time); ok {
			user.UpdatedAt = t
		}
	}

	if v, ok := data["city"].(string); ok && v != "" {
		user.City = &v
	}
	if v, ok := data["municipality"].(string); ok && v != "" {
		user.Municipality = &v
	}
	if v, ok := data["neighborhood"].(string); ok && v != "" {
		user.Neighborhood = &v
	}
	if v, ok := data["birth_date"].(string); ok && v != "" {
		user.BirthDate = &v
	}
	if v, ok := data["gender"].(string); ok && v != "" {
		g := models.Gender(v)
		user.Gender = &g
	}

	return user, nil
}

// userToMap convierte User a map con tipos explícitos para evitar
// que Firestore serialice uuid.UUID como array de bytes.
func userToMap(user *models.User) map[string]interface{} {
	m := map[string]interface{}{
		"id":         user.ID.String(),
		"email":      user.Email,
		"password":   user.Password,
		"name":       user.Name,
		"phone":      user.Phone,
		"role":       string(user.Role),
		"is_active":  user.IsActive,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
		// Inicializar nullable como nil siempre
		"city":         nil,
		"municipality": nil,
		"neighborhood": nil,
		"birth_date":   nil,
		"gender":       nil,
	}

	if user.City != nil {
		m["city"] = *user.City
	}
	if user.Municipality != nil {
		m["municipality"] = *user.Municipality
	}
	if user.Neighborhood != nil {
		m["neighborhood"] = *user.Neighborhood
	}
	if user.BirthDate != nil {
		m["birth_date"] = *user.BirthDate
	}
	if user.Gender != nil {
		m["gender"] = string(*user.Gender)
	}

	return m
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	m := userToMap(user)
	log.Printf("[DEBUG Create] guardando usuario %s con map: %v", user.ID.String(), m)

	_, err := r.firebase.Collection("users").Doc(user.ID.String()).Set(ctx, m)
	if err != nil {
		log.Printf("[DEBUG Create] ERROR: %v", err)
	}
	return err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	iter := r.firebase.Collection("users").Where("email", "==", email).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, errors.New("usuario no encontrado")
	}
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(doc.Ref.ID)
	if err != nil {
		return nil, err
	}

	return userFromDoc(doc, userID)
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	doc, err := r.firebase.Collection("users").Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errors.New("usuario no encontrado")
		}
		return nil, err
	}

	return userFromDoc(doc, id)
}

// Update guarda los campos de perfil usando Set + MergeAll.
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	profileFields := map[string]interface{}{
		"name":         user.Name,
		"phone":        user.Phone,
		"updated_at":   user.UpdatedAt,
		"city":         nil,
		"municipality": nil,
		"neighborhood": nil,
		"birth_date":   nil,
		"gender":       nil,
	}

	if user.City != nil {
		profileFields["city"] = *user.City
	}
	if user.Municipality != nil {
		profileFields["municipality"] = *user.Municipality
	}
	if user.Neighborhood != nil {
		profileFields["neighborhood"] = *user.Neighborhood
	}
	if user.BirthDate != nil {
		profileFields["birth_date"] = *user.BirthDate
	}
	if user.Gender != nil {
		profileFields["gender"] = string(*user.Gender)
	}

	// ── DEBUG ─────────────────────────────────────────────────────────────────
	log.Printf("[DEBUG Update] doc=%s profileFields=%v", user.ID.String(), profileFields)
	// ── FIN DEBUG ──────────────────────────────────────────────────────────────

	_, err := r.firebase.Collection("users").Doc(user.ID.String()).Set(ctx, profileFields, firestore.MergeAll)
	if err != nil {
		log.Printf("[DEBUG Update] ERROR Firestore: %v", err)
	} else {
		log.Printf("[DEBUG Update] Firestore OK para doc=%s", user.ID.String())
	}
	return err
}

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

func (r *UserRepository) GetAll(ctx context.Context) ([]*models.User, error) {
	iter := r.firebase.Collection("users").OrderBy("created_at", firestore.Desc).Documents(ctx)
	defer iter.Stop()

	var users []*models.User
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		userID, err := uuid.Parse(doc.Ref.ID)
		if err != nil {
			continue
		}

		user, err := userFromDoc(doc, userID)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepository) UpdateByID(ctx context.Context, id uuid.UUID, user *models.User) error {
	user.UpdatedAt = time.Now()
	user.ID = id
	m := userToMap(user)
	_, err := r.firebase.Collection("users").Doc(id.String()).Set(ctx, m)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.firebase.Collection("users").Doc(id.String()).Delete(ctx)
	return err
}

func (r *UserRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.firebase.Collection("users").Doc(id.String()).Delete(ctx)
	return err
}
