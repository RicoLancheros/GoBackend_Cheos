package models

import (
	"time"
)

// --- Ventas ---

// PaymentMethodStats estadisticas por metodo de pago
type PaymentMethodStats struct {
	Count int     `json:"count" firestore:"count"`
	Total float64 `json:"total" firestore:"total"`
}

// DailyBreakdown desglose diario de ventas
type DailyBreakdown struct {
	Revenue float64 `json:"revenue" firestore:"revenue"`
	Orders  int     `json:"orders" firestore:"orders"`
}

// MonthlyBreakdown desglose mensual de ventas
type MonthlyBreakdown struct {
	Revenue float64 `json:"revenue" firestore:"revenue"`
	Orders  int     `json:"orders" firestore:"orders"`
}

// SalesMonthlyMetrics metricas de ventas mensuales
type SalesMonthlyMetrics struct {
	Year               int                           `json:"year" firestore:"year"`
	Month              int                           `json:"month" firestore:"month"`
	TotalRevenue       float64                       `json:"total_revenue" firestore:"total_revenue"`
	TotalOrders        int                           `json:"total_orders" firestore:"total_orders"`
	CompletedOrders    int                           `json:"completed_orders" firestore:"completed_orders"`
	CancelledOrders    int                           `json:"cancelled_orders" firestore:"cancelled_orders"`
	PendingOrders      int                           `json:"pending_orders" firestore:"pending_orders"`
	AverageTicket      float64                       `json:"average_ticket" firestore:"average_ticket"`
	TotalDiscountGiven float64                       `json:"total_discount_given" firestore:"total_discount_given"`
	OrdersWithDiscount int                           `json:"orders_with_discount" firestore:"orders_with_discount"`
	PaymentMethods     map[string]PaymentMethodStats `json:"payment_methods" firestore:"payment_methods"`
	DailyBreakdown     map[string]DailyBreakdown     `json:"daily_breakdown" firestore:"daily_breakdown"`
	UpdatedAt          time.Time                     `json:"updated_at" firestore:"updated_at"`
}

// SalesYearlyMetrics metricas de ventas anuales
type SalesYearlyMetrics struct {
	Year             int                         `json:"year" firestore:"year"`
	TotalRevenue     float64                     `json:"total_revenue" firestore:"total_revenue"`
	TotalOrders      int                         `json:"total_orders" firestore:"total_orders"`
	CompletedOrders  int                         `json:"completed_orders" firestore:"completed_orders"`
	CancelledOrders  int                         `json:"cancelled_orders" firestore:"cancelled_orders"`
	AverageTicket    float64                     `json:"average_ticket" firestore:"average_ticket"`
	MonthlyBreakdown map[string]MonthlyBreakdown `json:"monthly_breakdown" firestore:"monthly_breakdown"`
	UpdatedAt        time.Time                   `json:"updated_at" firestore:"updated_at"`
}

// --- Compradores ---

// BuyersMonthlyMetrics estadisticas de compradores mensuales
type BuyersMonthlyMetrics struct {
	Year                   int            `json:"year" firestore:"year"`
	Month                  int            `json:"month" firestore:"month"`
	TotalBuyers            int            `json:"total_buyers" firestore:"total_buyers"`
	RegisteredBuyers       int            `json:"registered_buyers" firestore:"registered_buyers"`
	GuestBuyers            int            `json:"guest_buyers" firestore:"guest_buyers"`
	NewRegisteredThisMonth int            `json:"new_registered_this_month" firestore:"new_registered_this_month"`
	ReturningBuyers        int            `json:"returning_buyers" firestore:"returning_buyers"`
	GenderBreakdown        map[string]int `json:"gender_breakdown" firestore:"gender_breakdown"`
	AgeBreakdown           map[string]int `json:"age_breakdown" firestore:"age_breakdown"`
	CityBreakdown          map[string]int `json:"city_breakdown" firestore:"city_breakdown"`
	BuyerIDs               []string       `json:"buyer_ids" firestore:"buyer_ids"`
	UpdatedAt              time.Time      `json:"updated_at" firestore:"updated_at"`
}

// BuyersYearlyMetrics estadisticas de compradores anuales
type BuyersYearlyMetrics struct {
	Year              int            `json:"year" firestore:"year"`
	TotalUniqueBuyers int            `json:"total_unique_buyers" firestore:"total_unique_buyers"`
	RegisteredBuyers  int            `json:"registered_buyers" firestore:"registered_buyers"`
	GuestBuyers       int            `json:"guest_buyers" firestore:"guest_buyers"`
	GenderBreakdown   map[string]int `json:"gender_breakdown" firestore:"gender_breakdown"`
	AgeBreakdown      map[string]int `json:"age_breakdown" firestore:"age_breakdown"`
	BuyerIDs          []string       `json:"buyer_ids" firestore:"buyer_ids"`
	UpdatedAt         time.Time      `json:"updated_at" firestore:"updated_at"`
}

// --- Productos ---

// ProductStats estadisticas de un producto individual
type ProductStats struct {
	ProductID     string  `json:"product_id" firestore:"product_id"`
	ProductName   string  `json:"product_name" firestore:"product_name"`
	TotalQuantity int     `json:"total_quantity" firestore:"total_quantity"`
	TotalRevenue  float64 `json:"total_revenue" firestore:"total_revenue"`
	OrderCount    int     `json:"order_count" firestore:"order_count"`
}

// ProductStatsMap estadisticas de un producto en el mapa all_products
type ProductStatsMap struct {
	Name     string  `json:"name" firestore:"name"`
	Quantity int     `json:"quantity" firestore:"quantity"`
	Revenue  float64 `json:"revenue" firestore:"revenue"`
	Orders   int     `json:"orders" firestore:"orders"`
}

// TopProductsMetrics metricas de top productos (mensual y anual)
type TopProductsMetrics struct {
	Year        int                        `json:"year" firestore:"year"`
	Month       int                        `json:"month,omitempty" firestore:"month,omitempty"`
	MostSold    []ProductStats             `json:"most_sold" firestore:"most_sold"`
	LeastSold   []ProductStats             `json:"least_sold" firestore:"least_sold"`
	AllProducts map[string]ProductStatsMap `json:"all_products" firestore:"all_products"`
	UpdatedAt   time.Time                  `json:"updated_at" firestore:"updated_at"`
}

// --- DTO para /summary ---

// DashboardSummaryCurrentMonth resumen del mes actual
type DashboardSummaryCurrentMonth struct {
	Revenue       float64 `json:"revenue"`
	Orders        int     `json:"orders"`
	AverageTicket float64 `json:"average_ticket"`
	NewBuyers     int     `json:"new_buyers"`
}

// DashboardSummaryCurrentYear resumen del ano actual
type DashboardSummaryCurrentYear struct {
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

// DashboardSummaryTopProduct producto en el resumen del dashboard
type DashboardSummaryTopProduct struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Revenue  float64 `json:"revenue"`
}

// DashboardSummary resumen consolidado del dashboard
type DashboardSummary struct {
	CurrentMonth    DashboardSummaryCurrentMonth `json:"current_month"`
	CurrentYear     DashboardSummaryCurrentYear  `json:"current_year"`
	TopProducts     []DashboardSummaryTopProduct `json:"top_products"`
	GenderBreakdown map[string]int               `json:"gender_breakdown"`
	AgeBreakdown    map[string]int               `json:"age_breakdown"`
}
