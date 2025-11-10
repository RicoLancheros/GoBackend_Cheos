package models

import (
	"time"

	"github.com/google/uuid"
)

type PaymentMethod string
type PaymentStatus string
type OrderStatus string

const (
	PaymentMercadoPago  PaymentMethod = "MERCADO_PAGO"
	PaymentContraEntrega PaymentMethod = "CONTRA_ENTREGA"

	PaymentPending  PaymentStatus = "PENDING"
	PaymentApproved PaymentStatus = "APPROVED"
	PaymentRejected PaymentStatus = "REJECTED"
	PaymentRefunded PaymentStatus = "REFUNDED"

	OrderPending   OrderStatus = "PENDING"
	OrderConfirmed OrderStatus = "CONFIRMED"
	OrderProcessing OrderStatus = "PROCESSING"
	OrderShipped   OrderStatus = "SHIPPED"
	OrderDelivered OrderStatus = "DELIVERED"
	OrderCancelled OrderStatus = "CANCELLED"
)

type ShippingAddress struct {
	Street     string `json:"street" firestore:"street"`
	Number     string `json:"number" firestore:"number"`
	City       string `json:"city" firestore:"city"`
	Department string `json:"department" firestore:"department"`
	ZipCode    string `json:"zip_code" firestore:"zip_code"`
	Details    string `json:"details" firestore:"details"`
}

type Order struct {
	ID              uuid.UUID        `json:"id" firestore:"id"`
	OrderNumber     string           `json:"order_number" firestore:"order_number"`
	UserID          *uuid.UUID       `json:"user_id" firestore:"user_id"`
	CustomerName    string           `json:"customer_name" firestore:"customer_name"`
	CustomerEmail   string           `json:"customer_email" firestore:"customer_email"`
	CustomerPhone   string           `json:"customer_phone" firestore:"customer_phone"`
	Subtotal        float64          `json:"subtotal" firestore:"subtotal"`
	Discount        float64          `json:"discount" firestore:"discount"`
	Total           float64          `json:"total" firestore:"total"`
	PaymentMethod   PaymentMethod    `json:"payment_method" firestore:"payment_method"`
	PaymentStatus   PaymentStatus    `json:"payment_status" firestore:"payment_status"`
	MPPaymentID     *string          `json:"mp_payment_id" firestore:"mp_payment_id"`
	Status          OrderStatus      `json:"status" firestore:"status"`
	ShippingAddress *ShippingAddress `json:"shipping_address" firestore:"shipping_address"`
	DiscountCodeID  *uuid.UUID       `json:"discount_code_id" firestore:"discount_code_id"`
	UTMSource       string           `json:"utm_source" firestore:"utm_source"`
	UTMMedium       string           `json:"utm_medium" firestore:"utm_medium"`
	UTMCampaign     string           `json:"utm_campaign" firestore:"utm_campaign"`
	CreatedAt       time.Time        `json:"created_at" firestore:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at" firestore:"updated_at"`
}

type OrderItem struct {
	ID          uuid.UUID `json:"id" firestore:"id"`
	OrderID     uuid.UUID `json:"order_id" firestore:"order_id"`
	ProductID   uuid.UUID `json:"product_id" firestore:"product_id"`
	ProductName string    `json:"product_name" firestore:"product_name"`
	Quantity    int       `json:"quantity" firestore:"quantity"`
	Price       float64   `json:"price" firestore:"price"`
	Subtotal    float64   `json:"subtotal" firestore:"subtotal"`
}

// DTOs

type CreateOrderItemRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,gt=0"`
}

type CreateOrderRequest struct {
	CustomerName    string                   `json:"customer_name" validate:"required,min=2"`
	CustomerEmail   string                   `json:"customer_email" validate:"required,email"`
	CustomerPhone   string                   `json:"customer_phone" validate:"required"`
	PaymentMethod   PaymentMethod            `json:"payment_method" validate:"required"`
	ShippingAddress ShippingAddress          `json:"shipping_address" validate:"required"`
	Items           []CreateOrderItemRequest `json:"items" validate:"required,min=1,dive"`
	DiscountCode    string                   `json:"discount_code"`
	UTMSource       string                   `json:"utm_source"`
	UTMMedium       string                   `json:"utm_medium"`
	UTMCampaign     string                   `json:"utm_campaign"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" validate:"required"`
}

type UpdatePaymentStatusRequest struct {
	PaymentStatus PaymentStatus `json:"payment_status" validate:"required"`
}

type OrderWithItems struct {
	Order Order       `json:"order"`
	Items []OrderItem `json:"items"`
}

type OrderListResponse struct {
	Orders     []Order `json:"orders"`
	Total      int     `json:"total"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	TotalPages int     `json:"total_pages"`
}
