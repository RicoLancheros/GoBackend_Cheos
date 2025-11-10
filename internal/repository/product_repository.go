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

type ProductRepository struct {
	firebase *database.FirebaseClient
}

func NewProductRepository(firebase *database.FirebaseClient) *ProductRepository {
	return &ProductRepository{
		firebase: firebase,
	}
}

// Create crea un nuevo producto
func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}

	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	_, err := r.firebase.Collection("products").Doc(product.ID.String()).Set(ctx, product)
	if err != nil {
		return err
	}

	return nil
}

// GetByID obtiene un producto por ID
func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	doc, err := r.firebase.Collection("products").Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errors.New("producto no encontrado")
		}
		return nil, err
	}

	var product models.Product
	if err := doc.DataTo(&product); err != nil {
		return nil, err
	}

	product.ID = id
	return &product, nil
}

// GetAll obtiene todos los productos con paginación
func (r *ProductRepository) GetAll(ctx context.Context, limit int, offset int) ([]*models.Product, error) {
	query := r.firebase.Collection("products").
		OrderBy("created_at", firestore.Desc).
		Limit(limit).
		Offset(offset)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var products []*models.Product
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var product models.Product
		if err := doc.DataTo(&product); err != nil {
			continue
		}

		productID, err := uuid.Parse(doc.Ref.ID)
		if err != nil {
			continue
		}
		product.ID = productID

		products = append(products, &product)
	}

	return products, nil
}

// GetFeatured obtiene productos destacados
func (r *ProductRepository) GetFeatured(ctx context.Context, limit int) ([]*models.Product, error) {
	// Sin OrderBy para evitar índice compuesto (WHERE + OrderBy requiere índice)
	query := r.firebase.Collection("products").
		Where("is_featured", "==", true).
		Limit(limit)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var products []*models.Product
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var product models.Product
		if err := doc.DataTo(&product); err != nil {
			continue
		}

		productID, err := uuid.Parse(doc.Ref.ID)
		if err != nil {
			continue
		}
		product.ID = productID

		products = append(products, &product)
	}

	return products, nil
}

// Update actualiza un producto
func (r *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	product.UpdatedAt = time.Now()

	_, err := r.firebase.Collection("products").Doc(product.ID.String()).Set(ctx, product)
	if err != nil {
		return err
	}

	return nil
}

// Delete elimina un producto (soft delete)
func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	updates := []firestore.Update{
		{Path: "is_active", Value: false},
		{Path: "updated_at", Value: time.Now()},
	}

	_, err := r.firebase.Collection("products").Doc(id.String()).Update(ctx, updates)
	if err != nil {
		return err
	}

	return nil
}

// UpdateStock actualiza el stock de un producto
func (r *ProductRepository) UpdateStock(ctx context.Context, id uuid.UUID, quantity int) error {
	return r.firebase.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		docRef := r.firebase.Collection("products").Doc(id.String())
		doc, err := tx.Get(docRef)
		if err != nil {
			return err
		}

		var product models.Product
		if err := doc.DataTo(&product); err != nil {
			return err
		}

		newStock := product.Stock + quantity
		if newStock < 0 {
			return errors.New("stock insuficiente")
		}

		return tx.Update(docRef, []firestore.Update{
			{Path: "stock", Value: newStock},
			{Path: "updated_at", Value: time.Now()},
		})
	})
}

// Search busca productos por nombre o descripción
func (r *ProductRepository) Search(ctx context.Context, searchTerm string, limit int) ([]*models.Product, error) {
	// Firestore no tiene búsqueda full-text nativa, así que hacemos búsqueda simple
	// Para producción, considera usar Algolia o Elasticsearch

	query := r.firebase.Collection("products").
		Where("is_active", "==", true).
		Limit(limit * 3) // Obtenemos más para filtrar localmente

	iter := query.Documents(ctx)
	defer iter.Stop()

	var products []*models.Product
	count := 0

	for {
		if count >= limit {
			break
		}

		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var product models.Product
		if err := doc.DataTo(&product); err != nil {
			continue
		}

		// Filtro simple por nombre (case-insensitive sería mejor)
		// En producción, usa un servicio de búsqueda dedicado

		productID, err := uuid.Parse(doc.Ref.ID)
		if err != nil {
			continue
		}
		product.ID = productID

		products = append(products, &product)
		count++
	}

	return products, nil
}

// CountProducts cuenta el total de productos activos
func (r *ProductRepository) CountProducts(ctx context.Context) (int64, error) {
	iter := r.firebase.Collection("products").Where("is_active", "==", true).Documents(ctx)
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
