package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/repository"
	"github.com/google/uuid"
)

// SSEPublisher es la interfaz mínima que necesita el service.
// Vive aquí para evitar import circular (handlers → services).
// SSEHub la implementa.
type SSEPublisher interface {
	Publish(userID string, payload []byte)
}

type statusMeta struct {
	title string
	body  func(orderNum string) string
	icon  string
}

var orderStatusMessages = map[models.OrderStatus]statusMeta{
	models.OrderConfirmed: {
		title: "Pedido confirmado",
		body:  func(n string) string { return fmt.Sprintf("Tu pedido %s fue confirmado y está en preparación.", n) },
		icon:  "confirmed",
	},
	models.OrderProcessing: {
		title: "En preparación",
		body:  func(n string) string { return fmt.Sprintf("Tu pedido %s ya está siendo preparado con cariño.", n) },
		icon:  "processing",
	},
	models.OrderShipped: {
		title: "¡Tu pedido va en camino!",
		body:  func(n string) string { return fmt.Sprintf("El pedido %s fue enviado y llegará pronto.", n) },
		icon:  "shipped",
	},
	models.OrderDelivered: {
		title: "Pedido entregado",
		body:  func(n string) string { return fmt.Sprintf("El pedido %s fue entregado. ¡Gracias por tu compra!", n) },
		icon:  "delivered",
	},
	models.OrderCancelled: {
		title: "Pedido cancelado",
		body: func(n string) string {
			return fmt.Sprintf("El pedido %s fue cancelado. Contáctanos si necesitas ayuda.", n)
		},
		icon: "cancelled",
	},
}

type NotificationService struct {
	repo *repository.NotificationRepository
	hub  SSEPublisher
}

func NewNotificationService(repo *repository.NotificationRepository, hub SSEPublisher) *NotificationService {
	return &NotificationService{repo: repo, hub: hub}
}

// CreateOrderStatusNotification crea una notificación cuando cambia el estado de una orden
// y la empuja en tiempo real al cliente si está conectado vía SSE.
// Debe llamarse DESPUÉS de que el repo ya persistió el nuevo estado.
func (s *NotificationService) CreateOrderStatusNotification(
	ctx context.Context,
	order *models.Order,
	newStatus models.OrderStatus,
) {
	log.Printf("[Notif] llamado — orderID=%s orderNumber=%q userID=%v newStatus=%s",
		order.ID, order.OrderNumber, order.UserID, newStatus)

	if order.UserID == nil {
		log.Printf("[Notif] SKIP — orden %q no tiene userID (pedido invitado)", order.OrderNumber)
		return
	}

	meta, ok := orderStatusMessages[newStatus]
	if !ok {
		log.Printf("[Notif] SKIP — no hay mensaje configurado para estado %q", newStatus)
		return
	}

	n := &models.Notification{
		UserID:      *order.UserID,
		Type:        models.NotificationOrderStatus,
		OrderID:     order.ID,
		OrderNum:    order.OrderNumber,
		Title:       meta.title,
		Body:        meta.body(order.OrderNumber),
		Icon:        meta.icon,
		OrderStatus: string(newStatus),
	}

	if err := s.repo.Create(ctx, n); err != nil {
		log.Printf("[Notif] ERROR guardando notificación para orden %q: %v", order.OrderNumber, err)
		return
	}

	// Publicar en tiempo real — no-op silencioso si el usuario no está conectado
	if b, err := json.Marshal(n); err == nil {
		s.hub.Publish(n.UserID.String(), b)
	}

	log.Printf("[Notif] OK — notificación creada y publicada: user=%s orden=%q estado=%s",
		order.UserID, order.OrderNumber, newStatus)
}

// GetUserNotifications retorna todas las notificaciones de un usuario con conteo de no leídas
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID) (*models.NotificationListResponse, error) {
	notifications, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	unread := 0
	list := make([]models.Notification, 0, len(notifications))
	for _, n := range notifications {
		if n.Status == models.NotificationUnread {
			unread++
		}
		list = append(list, *n)
	}

	return &models.NotificationListResponse{
		Notifications: list,
		Total:         len(list),
		UnreadCount:   unread,
	}, nil
}

// MarkRead marca una notificación específica como leída
func (s *NotificationService) MarkRead(ctx context.Context, notifID uuid.UUID, userID uuid.UUID) error {
	return s.repo.MarkRead(ctx, notifID, userID)
}

// MarkAllRead marca todas las notificaciones de un usuario como leídas
func (s *NotificationService) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return s.repo.MarkAllRead(ctx, userID)
}
