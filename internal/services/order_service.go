package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/repository"
	"github.com/google/uuid"
)

type OrderService struct {
	orderRepo   *repository.OrderRepository
	productRepo *repository.ProductRepository
}

func NewOrderService(orderRepo *repository.OrderRepository, productRepo *repository.ProductRepository) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

// CreateOrder crea una nueva orden
func (s *OrderService) CreateOrder(ctx context.Context, req *models.CreateOrderRequest, userID *uuid.UUID) (*models.OrderWithItems, error) {
	// Validar que haya items
	if len(req.Items) == 0 {
		return nil, errors.New("la orden debe tener al menos un producto")
	}

	// Calcular subtotal y validar productos
	var subtotal float64
	var orderItems []models.OrderItem

	for _, itemReq := range req.Items {
		// Obtener producto
		product, err := s.productRepo.GetByID(ctx, itemReq.ProductID)
		if err != nil {
			return nil, fmt.Errorf("producto %s no encontrado", itemReq.ProductID)
		}

		// Validar stock
		if product.Stock < itemReq.Quantity {
			return nil, fmt.Errorf("stock insuficiente para el producto %s", product.Name)
		}

		// Validar que el producto esté activo
		if !product.IsActive {
			return nil, fmt.Errorf("el producto %s no está disponible", product.Name)
		}

		// Calcular subtotal del item
		itemSubtotal := product.Price * float64(itemReq.Quantity)
		subtotal += itemSubtotal

		// Crear item de orden (sin ID aún, se creará después)
		orderItem := models.OrderItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    itemReq.Quantity,
			Price:       product.Price,
			Subtotal:    itemSubtotal,
		}

		orderItems = append(orderItems, orderItem)
	}

	// Aplicar descuento si existe (por ahora 0, se implementará con discount codes)
	discount := 0.0

	// Calcular total
	total := subtotal - discount

	// Crear orden
	order := &models.Order{
		UserID:          userID,
		CustomerName:    req.CustomerName,
		CustomerEmail:   req.CustomerEmail,
		CustomerPhone:   req.CustomerPhone,
		Subtotal:        subtotal,
		Discount:        discount,
		Total:           total,
		PaymentMethod:   req.PaymentMethod,
		PaymentStatus:   models.PaymentPending,
		Status:          models.OrderPending,
		ShippingAddress: &req.ShippingAddress,
		UTMSource:       req.UTMSource,
		UTMMedium:       req.UTMMedium,
		UTMCampaign:     req.UTMCampaign,
	}

	// Guardar orden
	err := s.orderRepo.Create(ctx, order)
	if err != nil {
		return nil, err
	}

	// Guardar items de la orden
	var savedItems []models.OrderItem
	for _, item := range orderItems {
		item.OrderID = order.ID
		err := s.orderRepo.CreateOrderItem(ctx, &item)
		if err != nil {
			return nil, err
		}
		savedItems = append(savedItems, item)

		// Reducir stock del producto
		err = s.productRepo.UpdateStock(ctx, item.ProductID, -item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("error al actualizar stock: %v", err)
		}
	}

	return &models.OrderWithItems{
		Order: *order,
		Items: savedItems,
	}, nil
}

// GetOrderByID obtiene una orden por ID con sus items
func (s *OrderService) GetOrderByID(ctx context.Context, id uuid.UUID) (*models.OrderWithItems, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	items, err := s.orderRepo.GetItemsByOrderID(ctx, id)
	if err != nil {
		return nil, err
	}

	var itemsList []models.OrderItem
	for _, item := range items {
		itemsList = append(itemsList, *item)
	}

	return &models.OrderWithItems{
		Order: *order,
		Items: itemsList,
	}, nil
}

// GetOrderByOrderNumber obtiene una orden por número de orden
func (s *OrderService) GetOrderByOrderNumber(ctx context.Context, orderNumber string) (*models.OrderWithItems, error) {
	order, err := s.orderRepo.GetByOrderNumber(ctx, orderNumber)
	if err != nil {
		return nil, err
	}

	items, err := s.orderRepo.GetItemsByOrderID(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	var itemsList []models.OrderItem
	for _, item := range items {
		itemsList = append(itemsList, *item)
	}

	return &models.OrderWithItems{
		Order: *order,
		Items: itemsList,
	}, nil
}

// GetUserOrders obtiene las órdenes de un usuario
func (s *OrderService) GetUserOrders(ctx context.Context, userID uuid.UUID, page int, pageSize int) (*models.OrderListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	orders, err := s.orderRepo.GetByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	total, err := s.orderRepo.CountOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	var ordersList []models.Order
	for _, order := range orders {
		ordersList = append(ordersList, *order)
	}

	return &models.OrderListResponse{
		Orders:     ordersList,
		Total:      int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetAllOrders obtiene todas las órdenes (solo admin)
func (s *OrderService) GetAllOrders(ctx context.Context, page int, pageSize int) (*models.OrderListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	orders, err := s.orderRepo.GetAll(ctx, pageSize, offset)
	if err != nil {
		return nil, err
	}

	total, err := s.orderRepo.CountOrders(ctx)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	var ordersList []models.Order
	for _, order := range orders {
		ordersList = append(ordersList, *order)
	}

	return &models.OrderListResponse{
		Orders:     ordersList,
		Total:      int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateOrderStatus actualiza el estado de una orden
func (s *OrderService) UpdateOrderStatus(ctx context.Context, id uuid.UUID, req *models.UpdateOrderStatusRequest) (*models.Order, error) {
	// Verificar que la orden existe
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validar transiciones de estado
	if !s.isValidStatusTransition(order.Status, req.Status) {
		return nil, fmt.Errorf("transición de estado inválida de %s a %s", order.Status, req.Status)
	}

	// Actualizar estado
	err = s.orderRepo.UpdateStatus(ctx, id, req.Status)
	if err != nil {
		return nil, err
	}

	// Si el estado es cancelado, devolver stock
	if req.Status == models.OrderCancelled {
		items, err := s.orderRepo.GetItemsByOrderID(ctx, id)
		if err != nil {
			return nil, err
		}

		for _, item := range items {
			err = s.productRepo.UpdateStock(ctx, item.ProductID, item.Quantity)
			if err != nil {
				return nil, fmt.Errorf("error al devolver stock: %v", err)
			}
		}
	}

	// Obtener orden actualizada
	return s.orderRepo.GetByID(ctx, id)
}

// UpdatePaymentStatus actualiza el estado de pago
func (s *OrderService) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, req *models.UpdatePaymentStatusRequest) (*models.Order, error) {
	// Verificar que la orden existe
	_, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Actualizar estado de pago
	err = s.orderRepo.UpdatePaymentStatus(ctx, id, req.PaymentStatus, nil)
	if err != nil {
		return nil, err
	}

	// Si el pago fue aprobado, actualizar estado de la orden a confirmado
	if req.PaymentStatus == models.PaymentApproved {
		err = s.orderRepo.UpdateStatus(ctx, id, models.OrderConfirmed)
		if err != nil {
			return nil, err
		}
	}

	// Obtener orden actualizada
	return s.orderRepo.GetByID(ctx, id)
}

// isValidStatusTransition valida si la transición de estado es válida
func (s *OrderService) isValidStatusTransition(from models.OrderStatus, to models.OrderStatus) bool {
	validTransitions := map[models.OrderStatus][]models.OrderStatus{
		models.OrderPending:    {models.OrderConfirmed, models.OrderCancelled},
		models.OrderConfirmed:  {models.OrderProcessing, models.OrderCancelled},
		models.OrderProcessing: {models.OrderShipped, models.OrderCancelled},
		models.OrderShipped:    {models.OrderDelivered},
		models.OrderDelivered:  {},
		models.OrderCancelled:  {},
	}

	allowedTransitions, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == to {
			return true
		}
	}

	return false
}
