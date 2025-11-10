package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/cheoscafe/backend/internal/database"
	"github.com/cheoscafe/backend/internal/models"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderRepository struct {
	firebase *database.FirebaseClient
}

func NewOrderRepository(firebase *database.FirebaseClient) *OrderRepository {
	return &OrderRepository{
		firebase: firebase,
	}
}

// Create crea una nueva orden
func (r *OrderRepository) Create(ctx context.Context, order *models.Order) error {
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}

	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now

	// Generate order number if not exists
	if order.OrderNumber == "" {
		order.OrderNumber = fmt.Sprintf("ORD-%s", order.ID.String()[:8])
	}

	_, err := r.firebase.Collection("orders").Doc(order.ID.String()).Set(ctx, order)
	if err != nil {
		return err
	}

	return nil
}

// CreateOrderItem crea un item de orden
func (r *OrderRepository) CreateOrderItem(ctx context.Context, item *models.OrderItem) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}

	_, err := r.firebase.Collection("order_items").Doc(item.ID.String()).Set(ctx, item)
	if err != nil {
		return err
	}

	return nil
}

// GetByID obtiene una orden por ID
func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	doc, err := r.firebase.Collection("orders").Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errors.New("orden no encontrada")
		}
		return nil, err
	}

	var order models.Order
	if err := doc.DataTo(&order); err != nil {
		return nil, err
	}

	order.ID = id
	return &order, nil
}

// GetByOrderNumber obtiene una orden por número de orden
func (r *OrderRepository) GetByOrderNumber(ctx context.Context, orderNumber string) (*models.Order, error) {
	iter := r.firebase.Collection("orders").
		Where("order_number", "==", orderNumber).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, errors.New("orden no encontrada")
	}
	if err != nil {
		return nil, err
	}

	var order models.Order
	if err := doc.DataTo(&order); err != nil {
		return nil, err
	}

	orderID, err := uuid.Parse(doc.Ref.ID)
	if err != nil {
		return nil, err
	}
	order.ID = orderID

	return &order, nil
}

// GetByUserID obtiene órdenes de un usuario
func (r *OrderRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Order, error) {
	query := r.firebase.Collection("orders").
		Where("user_id", "==", userID).
		Limit(limit).
		Offset(offset)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var orders []*models.Order
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var order models.Order
		if err := doc.DataTo(&order); err != nil {
			continue
		}

		orderID, err := uuid.Parse(doc.Ref.ID)
		if err != nil {
			continue
		}
		order.ID = orderID

		orders = append(orders, &order)
	}

	return orders, nil
}

// GetAll obtiene todas las órdenes con paginación
func (r *OrderRepository) GetAll(ctx context.Context, limit int, offset int) ([]*models.Order, error) {
	query := r.firebase.Collection("orders").
		OrderBy("created_at", firestore.Desc).
		Limit(limit).
		Offset(offset)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var orders []*models.Order
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var order models.Order
		if err := doc.DataTo(&order); err != nil {
			continue
		}

		orderID, err := uuid.Parse(doc.Ref.ID)
		if err != nil {
			continue
		}
		order.ID = orderID

		orders = append(orders, &order)
	}

	return orders, nil
}

// GetItemsByOrderID obtiene los items de una orden
func (r *OrderRepository) GetItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]*models.OrderItem, error) {
	iter := r.firebase.Collection("order_items").
		Where("order_id", "==", orderID).
		Documents(ctx)
	defer iter.Stop()

	var items []*models.OrderItem
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var item models.OrderItem
		if err := doc.DataTo(&item); err != nil {
			continue
		}

		itemID, err := uuid.Parse(doc.Ref.ID)
		if err != nil {
			continue
		}
		item.ID = itemID

		items = append(items, &item)
	}

	return items, nil
}

// Update actualiza una orden
func (r *OrderRepository) Update(ctx context.Context, order *models.Order) error {
	order.UpdatedAt = time.Now()

	_, err := r.firebase.Collection("orders").Doc(order.ID.String()).Set(ctx, order, firestore.MergeAll)
	if err != nil {
		return err
	}

	return nil
}

// UpdateStatus actualiza el estado de una orden
func (r *OrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.OrderStatus) error {
	updates := []firestore.Update{
		{Path: "status", Value: status},
		{Path: "updated_at", Value: time.Now()},
	}

	_, err := r.firebase.Collection("orders").Doc(id.String()).Update(ctx, updates)
	if err != nil {
		return err
	}

	return nil
}

// UpdatePaymentStatus actualiza el estado de pago
func (r *OrderRepository) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, paymentStatus models.PaymentStatus, mpPaymentID *string) error {
	updates := []firestore.Update{
		{Path: "payment_status", Value: paymentStatus},
		{Path: "updated_at", Value: time.Now()},
	}

	if mpPaymentID != nil {
		updates = append(updates, firestore.Update{Path: "mp_payment_id", Value: mpPaymentID})
	}

	_, err := r.firebase.Collection("orders").Doc(id.String()).Update(ctx, updates)
	if err != nil {
		return err
	}

	return nil
}

// CountOrders cuenta el total de órdenes
func (r *OrderRepository) CountOrders(ctx context.Context) (int64, error) {
	iter := r.firebase.Collection("orders").Documents(ctx)
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

// CountOrdersByUserID cuenta las órdenes de un usuario
func (r *OrderRepository) CountOrdersByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	iter := r.firebase.Collection("orders").
		Where("user_id", "==", userID).
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
