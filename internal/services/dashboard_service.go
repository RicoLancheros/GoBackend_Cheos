package services

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/repository"
)

type DashboardService struct {
	dashboardRepo *repository.DashboardRepository
	orderRepo     *repository.OrderRepository
	userRepo      *repository.UserRepository
}

func NewDashboardService(
	dashboardRepo *repository.DashboardRepository,
	orderRepo *repository.OrderRepository,
	userRepo *repository.UserRepository,
) *DashboardService {
	return &DashboardService{
		dashboardRepo: dashboardRepo,
		orderRepo:     orderRepo,
		userRepo:      userRepo,
	}
}

// ============================================================
// Helpers privados
// ============================================================

func ageRange(birthDate string) string {
	parsed, err := time.Parse("2006-01-02", birthDate)
	if err != nil {
		return "unknown"
	}
	now := time.Now()
	age := now.Year() - parsed.Year()
	if now.YearDay() < parsed.YearDay() {
		age--
	}
	switch {
	case age < 18:
		return "unknown"
	case age <= 24:
		return "18-24"
	case age <= 34:
		return "25-34"
	case age <= 44:
		return "35-44"
	case age <= 54:
		return "45-54"
	default:
		return "55+"
	}
}

func statusCategory(s models.OrderStatus) string {
	switch s {
	case models.OrderDelivered:
		return "completed"
	case models.OrderCancelled:
		return "cancelled"
	default:
		return "pending"
	}
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func newSalesMonthlyMetrics(year, month int) *models.SalesMonthlyMetrics {
	return &models.SalesMonthlyMetrics{
		Year:           year,
		Month:          month,
		PaymentMethods: make(map[string]models.PaymentMethodStats),
		DailyBreakdown: make(map[string]models.DailyBreakdown),
	}
}

func newSalesYearlyMetrics(year int) *models.SalesYearlyMetrics {
	return &models.SalesYearlyMetrics{
		Year:             year,
		MonthlyBreakdown: make(map[string]models.MonthlyBreakdown),
	}
}

func newBuyersMonthlyMetrics(year, month int) *models.BuyersMonthlyMetrics {
	return &models.BuyersMonthlyMetrics{
		Year:            year,
		Month:           month,
		GenderBreakdown: make(map[string]int),
		AgeBreakdown:    make(map[string]int),
		CityBreakdown:   make(map[string]int),
		BuyerIDs:        []string{},
	}
}

func newBuyersYearlyMetrics(year int) *models.BuyersYearlyMetrics {
	return &models.BuyersYearlyMetrics{
		Year:            year,
		GenderBreakdown: make(map[string]int),
		AgeBreakdown:    make(map[string]int),
		BuyerIDs:        []string{},
	}
}

func newTopProductsMetrics(year, month int) *models.TopProductsMetrics {
	return &models.TopProductsMetrics{
		Year:        year,
		Month:       month,
		MostSold:    []models.ProductStats{},
		LeastSold:   []models.ProductStats{},
		AllProducts: make(map[string]models.ProductStatsMap),
	}
}

func sortAndSliceTopProducts(metrics *models.TopProductsMetrics) {
	var products []models.ProductStats
	for pid, p := range metrics.AllProducts {
		products = append(products, models.ProductStats{
			ProductID:     pid,
			ProductName:   p.Name,
			TotalQuantity: p.Quantity,
			TotalRevenue:  p.Revenue,
			OrderCount:    p.Orders,
		})
	}

	sort.Slice(products, func(i, j int) bool {
		return products[i].TotalQuantity > products[j].TotalQuantity
	})
	if len(products) > 10 {
		metrics.MostSold = products[:10]
	} else {
		metrics.MostSold = make([]models.ProductStats, len(products))
		copy(metrics.MostSold, products)
	}

	sort.Slice(products, func(i, j int) bool {
		return products[i].TotalQuantity < products[j].TotalQuantity
	})
	if len(products) > 10 {
		metrics.LeastSold = products[:10]
	} else {
		metrics.LeastSold = make([]models.ProductStats, len(products))
		copy(metrics.LeastSold, products)
	}
}

func recalculateAverageTicket(totalRevenue float64, completedOrders int) float64 {
	if completedOrders > 0 {
		return totalRevenue / float64(completedOrders)
	}
	return 0
}

// ============================================================
// Métodos Event-Driven
// ============================================================

// OnOrderCreated se llama al crear una orden.
// SOLO incrementa PendingOrders y registra el comprador.
// El revenue y productos se cuentan en OnOrderStatusChanged cuando llega a DELIVERED.
func (s *DashboardService) OnOrderCreated(ctx context.Context, order *models.Order, items []models.OrderItem) error {
	year := order.CreatedAt.Year()
	month := int(order.CreatedAt.Month())

	// ---- (a) Ventas mensuales: solo pending ----
	salesMonthly, err := s.dashboardRepo.GetSalesMonthly(ctx, year, month)
	if err != nil {
		return fmt.Errorf("error obteniendo sales_monthly: %w", err)
	}
	if salesMonthly == nil {
		salesMonthly = newSalesMonthlyMetrics(year, month)
	}

	salesMonthly.PendingOrders++
	salesMonthly.UpdatedAt = time.Now()

	if err := s.dashboardRepo.SaveSalesMonthly(ctx, year, month, salesMonthly); err != nil {
		return fmt.Errorf("error guardando sales_monthly: %w", err)
	}

	// ---- (b) Compradores mensuales ----
	buyersMonthly, err := s.dashboardRepo.GetBuyersMonthly(ctx, year, month)
	if err != nil {
		return fmt.Errorf("error obteniendo buyers_monthly: %w", err)
	}
	if buyersMonthly == nil {
		buyersMonthly = newBuyersMonthlyMetrics(year, month)
	}

	var buyerID string
	if order.UserID != nil {
		buyerID = order.UserID.String()
	} else {
		buyerID = order.CustomerEmail
	}

	if !containsString(buyersMonthly.BuyerIDs, buyerID) {
		buyersMonthly.BuyerIDs = append(buyersMonthly.BuyerIDs, buyerID)
		buyersMonthly.TotalBuyers++

		if order.UserID != nil {
			buyersMonthly.RegisteredBuyers++
			user, userErr := s.userRepo.GetByID(ctx, *order.UserID)
			if userErr != nil {
				log.Printf("Warning: no se pudo obtener usuario %s: %v", order.UserID.String(), userErr)
				buyersMonthly.GenderBreakdown["UNKNOWN"]++
				buyersMonthly.AgeBreakdown["unknown"]++
				buyersMonthly.CityBreakdown["unknown"]++
			} else {
				if user.Gender != nil {
					buyersMonthly.GenderBreakdown[string(*user.Gender)]++
				} else {
					buyersMonthly.GenderBreakdown["UNKNOWN"]++
				}
				if user.BirthDate != nil {
					buyersMonthly.AgeBreakdown[ageRange(*user.BirthDate)]++
				} else {
					buyersMonthly.AgeBreakdown["unknown"]++
				}
				if user.City != nil {
					buyersMonthly.CityBreakdown[*user.City]++
				} else {
					buyersMonthly.CityBreakdown["unknown"]++
				}
				if user.CreatedAt.Year() == year && int(user.CreatedAt.Month()) == month {
					buyersMonthly.NewRegisteredThisMonth++
				}
			}
		} else {
			buyersMonthly.GuestBuyers++
			buyersMonthly.GenderBreakdown["UNKNOWN"]++
			buyersMonthly.AgeBreakdown["unknown"]++
			buyersMonthly.CityBreakdown["unknown"]++
		}

		buyersMonthly.ReturningBuyers = buyersMonthly.TotalBuyers - buyersMonthly.NewRegisteredThisMonth - buyersMonthly.GuestBuyers
		if buyersMonthly.ReturningBuyers < 0 {
			buyersMonthly.ReturningBuyers = 0
		}
	}

	buyersMonthly.UpdatedAt = time.Now()
	if err := s.dashboardRepo.SaveBuyersMonthly(ctx, year, month, buyersMonthly); err != nil {
		return fmt.Errorf("error guardando buyers_monthly: %w", err)
	}

	// ---- (c) Compradores anuales ----
	buyersYearly, err := s.dashboardRepo.GetBuyersYearly(ctx, year)
	if err != nil {
		return fmt.Errorf("error obteniendo buyers_yearly: %w", err)
	}
	if buyersYearly == nil {
		buyersYearly = newBuyersYearlyMetrics(year)
	}

	if !containsString(buyersYearly.BuyerIDs, buyerID) {
		buyersYearly.BuyerIDs = append(buyersYearly.BuyerIDs, buyerID)
		buyersYearly.TotalUniqueBuyers++

		if order.UserID != nil {
			buyersYearly.RegisteredBuyers++
			user, userErr := s.userRepo.GetByID(ctx, *order.UserID)
			if userErr != nil {
				buyersYearly.GenderBreakdown["UNKNOWN"]++
				buyersYearly.AgeBreakdown["unknown"]++
			} else {
				if user.Gender != nil {
					buyersYearly.GenderBreakdown[string(*user.Gender)]++
				} else {
					buyersYearly.GenderBreakdown["UNKNOWN"]++
				}
				if user.BirthDate != nil {
					buyersYearly.AgeBreakdown[ageRange(*user.BirthDate)]++
				} else {
					buyersYearly.AgeBreakdown["unknown"]++
				}
			}
		} else {
			buyersYearly.GuestBuyers++
			buyersYearly.GenderBreakdown["UNKNOWN"]++
			buyersYearly.AgeBreakdown["unknown"]++
		}
	}

	buyersYearly.UpdatedAt = time.Now()
	if err := s.dashboardRepo.SaveBuyersYearly(ctx, year, buyersYearly); err != nil {
		return fmt.Errorf("error guardando buyers_yearly: %w", err)
	}

	return nil
}

// OnOrderStatusChanged se llama cuando un admin cambia el estado de una orden.
//
// Lógica de revenue:
//   - DELIVERED  → sumar revenue, productos y completedOrders
//   - CANCELLED  → restar pending. Si venía de DELIVERED, restar revenue también
//   - Otros cambios dentro de "pending" → solo mover contadores de estado
func (s *DashboardService) OnOrderStatusChanged(ctx context.Context, order *models.Order, oldStatus, newStatus models.OrderStatus) error {
	year := order.CreatedAt.Year()
	month := int(order.CreatedAt.Month())
	dayKey := order.CreatedAt.Format("2006-01-02")
	monthKey := fmt.Sprintf("%02d", month)

	oldCat := statusCategory(oldStatus)
	newCat := statusCategory(newStatus)

	// Sin cambio de categoría — nada que hacer
	if oldCat == newCat {
		return nil
	}

	// ---- (a) Ventas mensuales ----
	salesMonthly, err := s.dashboardRepo.GetSalesMonthly(ctx, year, month)
	if err != nil {
		return fmt.Errorf("error obteniendo sales_monthly: %w", err)
	}
	if salesMonthly == nil {
		salesMonthly = newSalesMonthlyMetrics(year, month)
	}

	// FIX: Guard para evitar contadores negativos al decrementar
	switch oldCat {
	case "pending":
		if salesMonthly.PendingOrders > 0 {
			salesMonthly.PendingOrders--
		}
	case "completed":
		if salesMonthly.CompletedOrders > 0 {
			salesMonthly.CompletedOrders--
		}
	case "cancelled":
		if salesMonthly.CancelledOrders > 0 {
			salesMonthly.CancelledOrders--
		}
	}
	switch newCat {
	case "pending":
		salesMonthly.PendingOrders++
	case "completed":
		salesMonthly.CompletedOrders++
	case "cancelled":
		salesMonthly.CancelledOrders++
	}

	// DELIVERED → sumar revenue y pedidos completados
	if newStatus == models.OrderDelivered {
		salesMonthly.TotalRevenue += order.Total
		salesMonthly.TotalOrders++

		if order.Discount > 0 {
			salesMonthly.TotalDiscountGiven += order.Discount
			salesMonthly.OrdersWithDiscount++
		}

		pmKey := string(order.PaymentMethod)
		pm := salesMonthly.PaymentMethods[pmKey]
		pm.Count++
		pm.Total += order.Total
		salesMonthly.PaymentMethods[pmKey] = pm

		daily := salesMonthly.DailyBreakdown[dayKey]
		daily.Revenue += order.Total
		daily.Orders++
		salesMonthly.DailyBreakdown[dayKey] = daily
	}

	// CANCELLED desde DELIVERED → restar revenue (devolución)
	if newStatus == models.OrderCancelled && oldStatus == models.OrderDelivered {
		salesMonthly.TotalRevenue -= order.Total
		if salesMonthly.TotalOrders > 0 {
			salesMonthly.TotalOrders--
		}

		if order.Discount > 0 {
			salesMonthly.TotalDiscountGiven -= order.Discount
			if salesMonthly.OrdersWithDiscount > 0 {
				salesMonthly.OrdersWithDiscount--
			}
		}

		pmKey := string(order.PaymentMethod)
		if pm, ok := salesMonthly.PaymentMethods[pmKey]; ok {
			pm.Count--
			pm.Total -= order.Total
			salesMonthly.PaymentMethods[pmKey] = pm
		}

		if daily, ok := salesMonthly.DailyBreakdown[dayKey]; ok {
			daily.Revenue -= order.Total
			if daily.Orders > 0 {
				daily.Orders--
			}
			salesMonthly.DailyBreakdown[dayKey] = daily
		}
	}

	salesMonthly.AverageTicket = recalculateAverageTicket(salesMonthly.TotalRevenue, salesMonthly.CompletedOrders)
	salesMonthly.UpdatedAt = time.Now()

	if err := s.dashboardRepo.SaveSalesMonthly(ctx, year, month, salesMonthly); err != nil {
		return fmt.Errorf("error guardando sales_monthly: %w", err)
	}

	// ---- (b) Ventas anuales ----
	salesYearly, err := s.dashboardRepo.GetSalesYearly(ctx, year)
	if err != nil {
		return fmt.Errorf("error obteniendo sales_yearly: %w", err)
	}
	if salesYearly == nil {
		salesYearly = newSalesYearlyMetrics(year)
	}

	switch oldCat {
	case "completed":
		if salesYearly.CompletedOrders > 0 {
			salesYearly.CompletedOrders--
		}
	case "cancelled":
		if salesYearly.CancelledOrders > 0 {
			salesYearly.CancelledOrders--
		}
	}
	switch newCat {
	case "completed":
		salesYearly.CompletedOrders++
	case "cancelled":
		salesYearly.CancelledOrders++
	}

	if newStatus == models.OrderDelivered {
		salesYearly.TotalRevenue += order.Total
		salesYearly.TotalOrders++

		mb := salesYearly.MonthlyBreakdown[monthKey]
		mb.Revenue += order.Total
		mb.Orders++
		salesYearly.MonthlyBreakdown[monthKey] = mb
	}

	if newStatus == models.OrderCancelled && oldStatus == models.OrderDelivered {
		salesYearly.TotalRevenue -= order.Total
		if salesYearly.TotalOrders > 0 {
			salesYearly.TotalOrders--
		}

		if mb, ok := salesYearly.MonthlyBreakdown[monthKey]; ok {
			mb.Revenue -= order.Total
			if mb.Orders > 0 {
				mb.Orders--
			}
			salesYearly.MonthlyBreakdown[monthKey] = mb
		}
	}

	salesYearly.AverageTicket = recalculateAverageTicket(salesYearly.TotalRevenue, salesYearly.CompletedOrders)
	salesYearly.UpdatedAt = time.Now()

	if err := s.dashboardRepo.SaveSalesYearly(ctx, year, salesYearly); err != nil {
		return fmt.Errorf("error guardando sales_yearly: %w", err)
	}

	// ---- (c) Top productos — solo al DELIVERED ----
	if newStatus == models.OrderDelivered {
		items, err := s.orderRepo.GetItemsByOrderID(ctx, order.ID)
		if err != nil {
			log.Printf("Warning: no se pudieron obtener items para top_products orden %s: %v", order.ID.String(), err)
		} else {
			// Mensual
			topMonthly, err := s.dashboardRepo.GetTopProductsMonthly(ctx, year, month)
			if err != nil {
				return fmt.Errorf("error obteniendo top_products_monthly: %w", err)
			}
			if topMonthly == nil {
				topMonthly = newTopProductsMetrics(year, month)
			}
			for _, item := range items {
				pid := item.ProductID.String()
				p := topMonthly.AllProducts[pid]
				p.Name = item.ProductName
				p.Quantity += item.Quantity
				p.Revenue += item.Subtotal
				p.Orders++
				topMonthly.AllProducts[pid] = p
			}
			sortAndSliceTopProducts(topMonthly)
			topMonthly.UpdatedAt = time.Now()
			if err := s.dashboardRepo.SaveTopProductsMonthly(ctx, year, month, topMonthly); err != nil {
				return fmt.Errorf("error guardando top_products_monthly: %w", err)
			}

			// Anual
			topYearly, err := s.dashboardRepo.GetTopProductsYearly(ctx, year)
			if err != nil {
				return fmt.Errorf("error obteniendo top_products_yearly: %w", err)
			}
			if topYearly == nil {
				topYearly = newTopProductsMetrics(year, 0)
			}
			for _, item := range items {
				pid := item.ProductID.String()
				p := topYearly.AllProducts[pid]
				p.Name = item.ProductName
				p.Quantity += item.Quantity
				p.Revenue += item.Subtotal
				p.Orders++
				topYearly.AllProducts[pid] = p
			}
			sortAndSliceTopProducts(topYearly)
			topYearly.UpdatedAt = time.Now()
			if err := s.dashboardRepo.SaveTopProductsYearly(ctx, year, topYearly); err != nil {
				return fmt.Errorf("error guardando top_products_yearly: %w", err)
			}
		}
	}

	return nil
}

// ============================================================
// Métodos de Lectura
// ============================================================

func (s *DashboardService) GetSalesMonthly(ctx context.Context, year, month int) (*models.SalesMonthlyMetrics, error) {
	metrics, err := s.dashboardRepo.GetSalesMonthly(ctx, year, month)
	if err != nil {
		return nil, err
	}
	if metrics == nil {
		return newSalesMonthlyMetrics(year, month), nil
	}
	return metrics, nil
}

func (s *DashboardService) GetSalesYearly(ctx context.Context, year int) (*models.SalesYearlyMetrics, error) {
	metrics, err := s.dashboardRepo.GetSalesYearly(ctx, year)
	if err != nil {
		return nil, err
	}
	if metrics == nil {
		return newSalesYearlyMetrics(year), nil
	}
	return metrics, nil
}

func (s *DashboardService) GetBuyersMonthly(ctx context.Context, year, month int) (*models.BuyersMonthlyMetrics, error) {
	metrics, err := s.dashboardRepo.GetBuyersMonthly(ctx, year, month)
	if err != nil {
		return nil, err
	}
	if metrics == nil {
		return newBuyersMonthlyMetrics(year, month), nil
	}
	return metrics, nil
}

func (s *DashboardService) GetBuyersYearly(ctx context.Context, year int) (*models.BuyersYearlyMetrics, error) {
	metrics, err := s.dashboardRepo.GetBuyersYearly(ctx, year)
	if err != nil {
		return nil, err
	}
	if metrics == nil {
		return newBuyersYearlyMetrics(year), nil
	}
	return metrics, nil
}

func (s *DashboardService) GetTopProductsMonthly(ctx context.Context, year, month int) (*models.TopProductsMetrics, error) {
	metrics, err := s.dashboardRepo.GetTopProductsMonthly(ctx, year, month)
	if err != nil {
		return nil, err
	}
	if metrics == nil {
		return newTopProductsMetrics(year, month), nil
	}
	return metrics, nil
}

func (s *DashboardService) GetTopProductsYearly(ctx context.Context, year int) (*models.TopProductsMetrics, error) {
	metrics, err := s.dashboardRepo.GetTopProductsYearly(ctx, year)
	if err != nil {
		return nil, err
	}
	if metrics == nil {
		return newTopProductsMetrics(year, 0), nil
	}
	return metrics, nil
}

func (s *DashboardService) GetSummary(ctx context.Context) (*models.DashboardSummary, error) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	salesMonthly, err := s.GetSalesMonthly(ctx, year, month)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo sales_monthly para summary: %w", err)
	}

	salesYearly, err := s.GetSalesYearly(ctx, year)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo sales_yearly para summary: %w", err)
	}

	buyersMonthly, err := s.GetBuyersMonthly(ctx, year, month)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo buyers_monthly para summary: %w", err)
	}

	topProductsMonthly, err := s.GetTopProductsMonthly(ctx, year, month)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo top_products_monthly para summary: %w", err)
	}

	var topProducts []models.DashboardSummaryTopProduct
	for _, p := range topProductsMonthly.MostSold {
		topProducts = append(topProducts, models.DashboardSummaryTopProduct{
			Name:     p.ProductName,
			Quantity: p.TotalQuantity,
			Revenue:  p.TotalRevenue,
		})
	}

	summary := &models.DashboardSummary{
		CurrentMonth: models.DashboardSummaryCurrentMonth{
			Revenue:       salesMonthly.TotalRevenue,
			Orders:        salesMonthly.TotalOrders,
			AverageTicket: salesMonthly.AverageTicket,
			NewBuyers:     buyersMonthly.NewRegisteredThisMonth + buyersMonthly.GuestBuyers,
		},
		CurrentYear: models.DashboardSummaryCurrentYear{
			Revenue: salesYearly.TotalRevenue,
			Orders:  salesYearly.TotalOrders,
		},
		TopProducts:     topProducts,
		GenderBreakdown: buyersMonthly.GenderBreakdown,
		AgeBreakdown:    buyersMonthly.AgeBreakdown,
	}

	return summary, nil
}

// ============================================================
// Recálculo Manual
// ============================================================

func (s *DashboardService) RecalculateMonth(ctx context.Context, year, month int) error {
	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)
	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	monthlyOrders, err := s.orderRepo.GetOrdersByDateRange(ctx, startOfMonth, endOfMonth)
	if err != nil {
		return fmt.Errorf("error obteniendo ordenes del mes: %w", err)
	}
	yearlyOrders, err := s.orderRepo.GetOrdersByDateRange(ctx, startOfYear, endOfYear)
	if err != nil {
		return fmt.Errorf("error obteniendo ordenes del ano: %w", err)
	}

	salesMonthly := newSalesMonthlyMetrics(year, month)
	buyersMonthly := newBuyersMonthlyMetrics(year, month)
	topProductsMonthly := newTopProductsMetrics(year, month)

	salesYearly := newSalesYearlyMetrics(year)
	buyersYearly := newBuyersYearlyMetrics(year)
	topProductsYearly := newTopProductsMetrics(year, 0)

	for _, order := range monthlyOrders {
		items, itemsErr := s.orderRepo.GetItemsByOrderID(ctx, order.ID)
		if itemsErr != nil {
			log.Printf("Warning: no se pudieron obtener items de orden %s: %v", order.ID.String(), itemsErr)
			continue
		}
		s.recalcProcessOrder(ctx, order, items, salesMonthly, buyersMonthly, topProductsMonthly)
	}

	for _, order := range yearlyOrders {
		items, itemsErr := s.orderRepo.GetItemsByOrderID(ctx, order.ID)
		if itemsErr != nil {
			log.Printf("Warning: no se pudieron obtener items de orden %s: %v", order.ID.String(), itemsErr)
			continue
		}
		s.recalcProcessOrderYearly(ctx, order, items, salesYearly, buyersYearly, topProductsYearly)
	}

	salesMonthly.AverageTicket = recalculateAverageTicket(salesMonthly.TotalRevenue, salesMonthly.CompletedOrders)
	salesYearly.AverageTicket = recalculateAverageTicket(salesYearly.TotalRevenue, salesYearly.CompletedOrders)

	buyersMonthly.ReturningBuyers = buyersMonthly.TotalBuyers - buyersMonthly.NewRegisteredThisMonth - buyersMonthly.GuestBuyers
	if buyersMonthly.ReturningBuyers < 0 {
		buyersMonthly.ReturningBuyers = 0
	}

	sortAndSliceTopProducts(topProductsMonthly)
	sortAndSliceTopProducts(topProductsYearly)

	now := time.Now()
	salesMonthly.UpdatedAt = now
	salesYearly.UpdatedAt = now
	buyersMonthly.UpdatedAt = now
	buyersYearly.UpdatedAt = now
	topProductsMonthly.UpdatedAt = now
	topProductsYearly.UpdatedAt = now

	if err := s.dashboardRepo.SaveSalesMonthly(ctx, year, month, salesMonthly); err != nil {
		return fmt.Errorf("error guardando sales_monthly recalculado: %w", err)
	}
	if err := s.dashboardRepo.SaveSalesYearly(ctx, year, salesYearly); err != nil {
		return fmt.Errorf("error guardando sales_yearly recalculado: %w", err)
	}
	if err := s.dashboardRepo.SaveBuyersMonthly(ctx, year, month, buyersMonthly); err != nil {
		return fmt.Errorf("error guardando buyers_monthly recalculado: %w", err)
	}
	if err := s.dashboardRepo.SaveBuyersYearly(ctx, year, buyersYearly); err != nil {
		return fmt.Errorf("error guardando buyers_yearly recalculado: %w", err)
	}
	if err := s.dashboardRepo.SaveTopProductsMonthly(ctx, year, month, topProductsMonthly); err != nil {
		return fmt.Errorf("error guardando top_products_monthly recalculado: %w", err)
	}
	if err := s.dashboardRepo.SaveTopProductsYearly(ctx, year, topProductsYearly); err != nil {
		return fmt.Errorf("error guardando top_products_yearly recalculado: %w", err)
	}

	return nil
}

func (s *DashboardService) recalcProcessOrder(
	ctx context.Context,
	order *models.Order,
	items []*models.OrderItem,
	sales *models.SalesMonthlyMetrics,
	buyers *models.BuyersMonthlyMetrics,
	topProducts *models.TopProductsMetrics,
) {
	dayKey := order.CreatedAt.Format("2006-01-02")
	isDelivered := order.Status == models.OrderDelivered

	switch statusCategory(order.Status) {
	case "pending":
		sales.PendingOrders++
	case "completed":
		sales.CompletedOrders++
	case "cancelled":
		sales.CancelledOrders++
	}

	if isDelivered {
		sales.TotalRevenue += order.Total
		sales.TotalOrders++

		if order.Discount > 0 {
			sales.TotalDiscountGiven += order.Discount
			sales.OrdersWithDiscount++
		}

		pmKey := string(order.PaymentMethod)
		pm := sales.PaymentMethods[pmKey]
		pm.Count++
		pm.Total += order.Total
		sales.PaymentMethods[pmKey] = pm

		daily := sales.DailyBreakdown[dayKey]
		daily.Revenue += order.Total
		daily.Orders++
		sales.DailyBreakdown[dayKey] = daily

		for _, item := range items {
			pid := item.ProductID.String()
			p := topProducts.AllProducts[pid]
			p.Name = item.ProductName
			p.Quantity += item.Quantity
			p.Revenue += item.Subtotal
			p.Orders++
			topProducts.AllProducts[pid] = p
		}
	}

	var buyerID string
	if order.UserID != nil {
		buyerID = order.UserID.String()
	} else {
		buyerID = order.CustomerEmail
	}

	if !containsString(buyers.BuyerIDs, buyerID) {
		buyers.BuyerIDs = append(buyers.BuyerIDs, buyerID)
		buyers.TotalBuyers++

		if order.UserID != nil {
			buyers.RegisteredBuyers++
			user, userErr := s.userRepo.GetByID(ctx, *order.UserID)
			if userErr != nil {
				buyers.GenderBreakdown["UNKNOWN"]++
				buyers.AgeBreakdown["unknown"]++
				buyers.CityBreakdown["unknown"]++
			} else {
				if user.Gender != nil {
					buyers.GenderBreakdown[string(*user.Gender)]++
				} else {
					buyers.GenderBreakdown["UNKNOWN"]++
				}
				if user.BirthDate != nil {
					buyers.AgeBreakdown[ageRange(*user.BirthDate)]++
				} else {
					buyers.AgeBreakdown["unknown"]++
				}
				if user.City != nil {
					buyers.CityBreakdown[*user.City]++
				} else {
					buyers.CityBreakdown["unknown"]++
				}
				if user.CreatedAt.Year() == sales.Year && int(user.CreatedAt.Month()) == sales.Month {
					buyers.NewRegisteredThisMonth++
				}
			}
		} else {
			buyers.GuestBuyers++
			buyers.GenderBreakdown["UNKNOWN"]++
			buyers.AgeBreakdown["unknown"]++
			buyers.CityBreakdown["unknown"]++
		}
	}
}

func (s *DashboardService) recalcProcessOrderYearly(
	ctx context.Context,
	order *models.Order,
	items []*models.OrderItem,
	sales *models.SalesYearlyMetrics,
	buyers *models.BuyersYearlyMetrics,
	topProducts *models.TopProductsMetrics,
) {
	monthKey := fmt.Sprintf("%02d", int(order.CreatedAt.Month()))
	isDelivered := order.Status == models.OrderDelivered

	switch statusCategory(order.Status) {
	case "completed":
		sales.CompletedOrders++
	case "cancelled":
		sales.CancelledOrders++
	}

	if isDelivered {
		sales.TotalRevenue += order.Total
		sales.TotalOrders++

		mb := sales.MonthlyBreakdown[monthKey]
		mb.Revenue += order.Total
		mb.Orders++
		sales.MonthlyBreakdown[monthKey] = mb

		for _, item := range items {
			pid := item.ProductID.String()
			p := topProducts.AllProducts[pid]
			p.Name = item.ProductName
			p.Quantity += item.Quantity
			p.Revenue += item.Subtotal
			p.Orders++
			topProducts.AllProducts[pid] = p
		}
	}

	var buyerID string
	if order.UserID != nil {
		buyerID = order.UserID.String()
	} else {
		buyerID = order.CustomerEmail
	}

	if !containsString(buyers.BuyerIDs, buyerID) {
		buyers.BuyerIDs = append(buyers.BuyerIDs, buyerID)
		buyers.TotalUniqueBuyers++

		if order.UserID != nil {
			buyers.RegisteredBuyers++
			user, userErr := s.userRepo.GetByID(ctx, *order.UserID)
			if userErr != nil {
				buyers.GenderBreakdown["UNKNOWN"]++
				buyers.AgeBreakdown["unknown"]++
			} else {
				if user.Gender != nil {
					buyers.GenderBreakdown[string(*user.Gender)]++
				} else {
					buyers.GenderBreakdown["UNKNOWN"]++
				}
				if user.BirthDate != nil {
					buyers.AgeBreakdown[ageRange(*user.BirthDate)]++
				} else {
					buyers.AgeBreakdown["unknown"]++
				}
			}
		} else {
			buyers.GuestBuyers++
			buyers.GenderBreakdown["UNKNOWN"]++
			buyers.AgeBreakdown["unknown"]++
		}
	}
}
