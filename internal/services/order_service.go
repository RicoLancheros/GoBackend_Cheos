package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/repository"
	"github.com/google/uuid"
)

type OrderService struct {
	orderRepo        *repository.OrderRepository
	productRepo      *repository.ProductRepository
	cartRepo         *repository.CartRepository
	dashboardService *DashboardService
}

func NewOrderService(orderRepo *repository.OrderRepository, productRepo *repository.ProductRepository, cartRepo *repository.CartRepository, dashboardService *DashboardService) *OrderService {
	return &OrderService{
		orderRepo:        orderRepo,
		productRepo:      productRepo,
		cartRepo:         cartRepo,
		dashboardService: dashboardService,
	}
}

// CreateOrder crea una nueva orden
func (s *OrderService) CreateOrder(ctx context.Context, req *models.CreateOrderRequest, userID *uuid.UUID) (*models.OrderWithItems, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("la orden debe tener al menos un producto")
	}

	var subtotal float64
	var orderItems []models.OrderItem

	for _, itemReq := range req.Items {
		product, err := s.productRepo.GetByID(ctx, itemReq.ProductID)
		if err != nil {
			return nil, fmt.Errorf("producto %s no encontrado", itemReq.ProductID)
		}

		if product.Stock < itemReq.Quantity {
			return nil, fmt.Errorf("stock insuficiente para el producto %s", product.Name)
		}

		if !product.IsActive {
			return nil, fmt.Errorf("el producto %s no está disponible", product.Name)
		}

		itemSubtotal := product.Price * float64(itemReq.Quantity)
		subtotal += itemSubtotal

		orderItem := models.OrderItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    itemReq.Quantity,
			Price:       product.Price,
			Subtotal:    itemSubtotal,
		}

		orderItems = append(orderItems, orderItem)
	}

	discount := 0.0
	total := subtotal - discount

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

	err := s.orderRepo.Create(ctx, order)
	if err != nil {
		return nil, err
	}

	var savedItems []models.OrderItem
	for _, item := range orderItems {
		item.OrderID = order.ID
		err := s.orderRepo.CreateOrderItem(ctx, &item)
		if err != nil {
			return nil, err
		}
		savedItems = append(savedItems, item)

		err = s.productRepo.UpdateStock(ctx, item.ProductID, -item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("error al actualizar stock: %v", err)
		}
	}

	if userID != nil {
		_ = s.cartRepo.Delete(ctx, *userID)
	}

	if s.dashboardService != nil {
		if err := s.dashboardService.OnOrderCreated(ctx, order, savedItems); err != nil {
			log.Printf("Error actualizando metricas de dashboard: %v", err)
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
// statusGroup: "" = todas, "active" = activas, "completed" = DELIVERED o CANCELLED
func (s *OrderService) GetAllOrders(ctx context.Context, page int, pageSize int, statusGroup string) (*models.OrderListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 500 {
		pageSize = 10
	}

	// Traemos todo sin paginar y filtramos en memoria
	// (Firestore no soporta WHERE status NOT IN de forma simple)
	allOrders, err := s.orderRepo.GetAllUnpaginated(ctx)
	if err != nil {
		return nil, err
	}

	completedStatuses := map[models.OrderStatus]bool{
		models.OrderDelivered: true,
		models.OrderCancelled: true,
	}

	var filtered []models.Order
	for _, o := range allOrders {
		isCompleted := completedStatuses[o.Status]
		switch statusGroup {
		case "active":
			if !isCompleted {
				filtered = append(filtered, *o)
			}
		case "completed":
			if isCompleted {
				filtered = append(filtered, *o)
			}
		default:
			filtered = append(filtered, *o)
		}
	}

	total := len(filtered)

	// Paginar en memoria
	offset := (page - 1) * pageSize
	if offset > total {
		offset = total
	}
	end := offset + pageSize
	if end > total {
		end = total
	}
	paginated := filtered[offset:end]

	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}

	return &models.OrderListResponse{
		Orders:     paginated,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateOrderStatus actualiza el estado de una orden
func (s *OrderService) UpdateOrderStatus(ctx context.Context, id uuid.UUID, req *models.UpdateOrderStatusRequest) (*models.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !s.isValidStatusTransition(order.Status, req.Status) {
		return nil, fmt.Errorf("transición de estado inválida de %s a %s", order.Status, req.Status)
	}

	err = s.orderRepo.UpdateStatus(ctx, id, req.Status)
	if err != nil {
		return nil, err
	}

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

	if s.dashboardService != nil {
		if err := s.dashboardService.OnOrderStatusChanged(ctx, order, order.Status, req.Status); err != nil {
			log.Printf("Error actualizando metricas de dashboard: %v", err)
		}
	}

	return s.orderRepo.GetByID(ctx, id)
}

// UpdatePaymentStatus actualiza el estado de pago
func (s *OrderService) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, req *models.UpdatePaymentStatusRequest) (*models.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	err = s.orderRepo.UpdatePaymentStatus(ctx, id, req.PaymentStatus, nil)
	if err != nil {
		return nil, err
	}

	// Auto-avanzar a CONFIRMED solo si está en PENDING (flujo MercadoPago/Wompi)
	// Para contra entrega ya habrá avanzado, no se toca el status
	if req.PaymentStatus == models.PaymentApproved && order.Status == models.OrderPending {
		err = s.orderRepo.UpdateStatus(ctx, id, models.OrderConfirmed)
		if err != nil {
			return nil, err
		}
	}

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
