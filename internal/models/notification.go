package models

import (
	"time"

	"github.com/google/uuid"
)

type NotificationStatus string
type NotificationType string

const (
	NotificationUnread NotificationStatus = "UNREAD"
	NotificationRead   NotificationStatus = "READ"

	NotificationOrderStatus NotificationType = "ORDER_STATUS"
)

type Notification struct {
	ID          uuid.UUID          `json:"id" firestore:"id"`
	UserID      uuid.UUID          `json:"user_id" firestore:"user_id"`
	Type        NotificationType   `json:"type" firestore:"type"`
	Status      NotificationStatus `json:"status" firestore:"status"`
	OrderID     uuid.UUID          `json:"order_id" firestore:"order_id"`
	OrderNum    string             `json:"order_number" firestore:"order_number"`
	Title       string             `json:"title" firestore:"title"`
	Body        string             `json:"body" firestore:"body"`
	Icon        string             `json:"icon" firestore:"icon"`
	OrderStatus string             `json:"order_status" firestore:"order_status"`
	CreatedAt   time.Time          `json:"created_at" firestore:"created_at"`
	ReadAt      *time.Time         `json:"read_at,omitempty" firestore:"read_at,omitempty"`
}

type NotificationListResponse struct {
	Notifications []Notification `json:"notifications"`
	Total         int            `json:"total"`
	UnreadCount   int            `json:"unread_count"`
}
