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

// ageRange calcula el rango de edad a partir de birth_date ("YYYY-MM-DD")
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

// statusCategory mapea OrderStatus a una de 3 categorias de metrica
func statusCategory(s models.OrderStatus) string {
	switch s {
	case models.OrderDelivered:
		return "completed"
	case models.OrderCancelled:
		return "cancelled"
	default: // PENDING, CONFIRMED, PROCESSING, SHIPPED
		return "pending"
	}
}

// containsString verifica si un string esta en un slice
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// --- Funciones para inicializar documentos vacios ---

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

// sortAndSliceTopProducts ordena AllProducts y genera MostSold/LeastSold (top 10)
func sortAndSliceTopProducts(metrics *models.TopProductsMetrics) {
	// Convertir AllProducts map a slice de ProductStats
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

	// Ordenar por TotalQuantity descendente para MostSold
	sort.Slice(products, func(i, j int) bool {
		return products[i].TotalQuantity > products[j].TotalQuantity
	})

	// MostSold: top 10
	if len(products) > 10 {
		metrics.MostSold = products[:10]
	} else {
		metrics.MostSold = make([]models.ProductStats, len(products))
		copy(metrics.MostSold, products)
	}

	// Ordenar por TotalQuantity ascendente para LeastSold
	sort.Slice(products, func(i, j int) bool {
		return products[i].TotalQuantity < products[j].TotalQuantity
	})

	// LeastSold: top 10
	if len(products) > 10 {
		metrics.LeastSold = products[:10]
	} else {
		metrics.LeastSold = make([]models.ProductStats, len(products))
		copy(metrics.LeastSold, products)
	}
}

// recalculateAverageTicket calcula el average ticket: revenue / (total_orders - cancelled_orders)
func recalculateAverageTicket(totalRevenue float64, totalOrders, cancelledOrders int) float64 {
	activeOrders := totalOrders - cancelledOrders
	if activeOrders > 0 {
		return totalRevenue / float64(activeOrders)
	}
	return 0
}

// ============================================================
// Metodos Event-Driven
// ============================================================

// OnOrderCreated se llama despues de crear una orden exitosamente.
// Toda orden creada es venta aprobada (la pagina es un catalogo).
func (s *DashboardService) OnOrderCreated(ctx context.Context, order *models.Order, items []models.OrderItem) error {
	year := order.CreatedAt.Year()
	month := int(order.CreatedAt.Month())
	dayKey := order.CreatedAt.Format("2006-01-02")
	monthKey := fmt.Sprintf("%02d", month)

	// ---- (a) Ventas mensuales ----
	salesMonthly, err := s.dashboardRepo.GetSalesMonthly(ctx, year, month)
	if err != nil {
		return fmt.Errorf("error obteniendo sales_monthly: %w", err)
	}
	if salesMonthly == nil {
		salesMonthly = newSalesMonthlyMetrics(year, month)
	}

	salesMonthly.TotalRevenue += order.Total
	salesMonthly.TotalOrders++
	salesMonthly.PendingOrders++

	if order.Discount > 0 {
		salesMonthly.TotalDiscountGiven += order.Discount
		salesMonthly.OrdersWithDiscount++
	}

	// Payment methods
	pmKey := string(order.PaymentMethod)
	pm := salesMonthly.PaymentMethods[pmKey]
	pm.Count++
	pm.Total += order.Total
	salesMonthly.PaymentMethods[pmKey] = pm

	// Daily breakdown
	daily := salesMonthly.DailyBreakdown[dayKey]
	daily.Revenue += order.Total
	daily.Orders++
	salesMonthly.DailyBreakdown[dayKey] = daily

	// Recalcular average ticket
	salesMonthly.AverageTicket = recalculateAverageTicket(salesMonthly.TotalRevenue, salesMonthly.TotalOrders, salesMonthly.CancelledOrders)
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

	salesYearly.TotalRevenue += order.Total
	salesYearly.TotalOrders++

	// Monthly breakdown
	mb := salesYearly.MonthlyBreakdown[monthKey]
	mb.Revenue += order.Total
	mb.Orders++
	salesYearly.MonthlyBreakdown[monthKey] = mb

	salesYearly.AverageTicket = recalculateAverageTicket(salesYearly.TotalRevenue, salesYearly.TotalOrders, salesYearly.CancelledOrders)
	salesYearly.UpdatedAt = time.Now()

	if err := s.dashboardRepo.SaveSalesYearly(ctx, year, salesYearly); err != nil {
		return fmt.Errorf("error guardando sales_yearly: %w", err)
	}

	// ---- (c) Compradores mensuales ----
	buyersMonthly, err := s.dashboardRepo.GetBuyersMonthly(ctx, year, month)
	if err != nil {
		return fmt.Errorf("error obteniendo buyers_monthly: %w", err)
	}
	if buyersMonthly == nil {
		buyersMonthly = newBuyersMonthlyMetrics(year, month)
	}

	// Determinar buyer ID
	var buyerID string
	if order.UserID != nil {
		buyerID = order.UserID.String()
	} else {
		buyerID = order.CustomerEmail
	}

	// Solo procesar si es comprador nuevo del mes
	if !containsString(buyersMonthly.BuyerIDs, buyerID) {
		buyersMonthly.BuyerIDs = append(buyersMonthly.BuyerIDs, buyerID)
		buyersMonthly.TotalBuyers++

		if order.UserID != nil {
			// Usuario registrado
			buyersMonthly.RegisteredBuyers++

			user, userErr := s.userRepo.GetByID(ctx, *order.UserID)
			if userErr != nil {
				log.Printf("Warning: no se pudo obtener usuario %s para metricas: %v", order.UserID.String(), userErr)
				// Usar valores desconocidos
				buyersMonthly.GenderBreakdown["UNKNOWN"]++
				buyersMonthly.AgeBreakdown["unknown"]++
				buyersMonthly.CityBreakdown["unknown"]++
			} else {
				// Gender
				if user.Gender != nil {
					buyersMonthly.GenderBreakdown[string(*user.Gender)]++
				} else {
					buyersMonthly.GenderBreakdown["UNKNOWN"]++
				}

				// Age
				if user.BirthDate != nil {
					buyersMonthly.AgeBreakdown[ageRange(*user.BirthDate)]++
				} else {
					buyersMonthly.AgeBreakdown["unknown"]++
				}

				// City
				if user.City != nil {
					buyersMonthly.CityBreakdown[*user.City]++
				} else {
					buyersMonthly.CityBreakdown["unknown"]++
				}

				// Verificar si se registro este mismo mes
				if user.CreatedAt.Year() == year && int(user.CreatedAt.Month()) == month {
					buyersMonthly.NewRegisteredThisMonth++
				}
			}
		} else {
			// Invitado
			buyersMonthly.GuestBuyers++
			buyersMonthly.GenderBreakdown["UNKNOWN"]++
			buyersMonthly.AgeBreakdown["unknown"]++
			buyersMonthly.CityBreakdown["unknown"]++
		}

		// Recalcular returning buyers
		buyersMonthly.ReturningBuyers = buyersMonthly.TotalBuyers - buyersMonthly.NewRegisteredThisMonth - buyersMonthly.GuestBuyers
		if buyersMonthly.ReturningBuyers < 0 {
			buyersMonthly.ReturningBuyers = 0
		}
	}

	buyersMonthly.UpdatedAt = time.Now()
	if err := s.dashboardRepo.SaveBuyersMonthly(ctx, year, month, buyersMonthly); err != nil {
		return fmt.Errorf("error guardando buyers_monthly: %w", err)
	}

	// ---- (d) Compradores anuales ----
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

			// Reusar el user ya obtenido si es posible, sino obtener de nuevo
			user, userErr := s.userRepo.GetByID(ctx, *order.UserID)
			if userErr != nil {
				log.Printf("Warning: no se pudo obtener usuario %s para metricas anuales: %v", order.UserID.String(), userErr)
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

	// ---- (e) Top productos mensuales ----
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

	// ---- (f) Top productos anuales ----
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

	return nil
}

// OnOrderStatusChanged se llama cuando un admin cambia el estado de una orden.
// Solo actualiza documentos de ventas (no afecta compradores ni productos).
func (s *DashboardService) OnOrderStatusChanged(ctx context.Context, order *models.Order, oldStatus, newStatus models.OrderStatus) error {
	year := order.CreatedAt.Year()
	month := int(order.CreatedAt.Month())
	dayKey := order.CreatedAt.Format("2006-01-02")
	monthKey := fmt.Sprintf("%02d", month)

	oldCategory := statusCategory(oldStatus)
	newCategory := statusCategory(newStatus)

	// Si no cambia de categoria, solo actualizar timestamp
	if oldCategory == newCategory {
		return nil
	}

	// ---- (a) Ventas mensuales ----
	salesMonthly, err := s.dashboardRepo.GetSalesMonthly(ctx, year, month)
	if err != nil {
		return fmt.Errorf("error obteniendo sales_monthly: %w", err)
	}
	if salesMonthly == nil {
		// No deberia pasar, pero por seguridad
		return nil
	}

	// Decrementar categoria anterior
	switch oldCategory {
	case "pending":
		salesMonthly.PendingOrders--
	case "completed":
		salesMonthly.CompletedOrders--
	case "cancelled":
		salesMonthly.CancelledOrders--
	}

	// Incrementar categoria nueva
	switch newCategory {
	case "pending":
		salesMonthly.PendingOrders++
	case "completed":
		salesMonthly.CompletedOrders++
	case "cancelled":
		salesMonthly.CancelledOrders++
	}

	// Si se cancela: restar revenue y contadores
	if newCategory == "cancelled" {
		salesMonthly.TotalRevenue -= order.Total
		salesMonthly.TotalOrders--

		if order.Discount > 0 {
			salesMonthly.TotalDiscountGiven -= order.Discount
			salesMonthly.OrdersWithDiscount--
		}

		// Actualizar payment methods
		pmKey := string(order.PaymentMethod)
		if pm, ok := salesMonthly.PaymentMethods[pmKey]; ok {
			pm.Count--
			pm.Total -= order.Total
			salesMonthly.PaymentMethods[pmKey] = pm
		}

		// Actualizar daily breakdown
		if daily, ok := salesMonthly.DailyBreakdown[dayKey]; ok {
			daily.Revenue -= order.Total
			daily.Orders--
			salesMonthly.DailyBreakdown[dayKey] = daily
		}
	}

	// Si se revierte una cancelacion (defensivo, no deberia pasar por validacion de transiciones)
	if oldCategory == "cancelled" && newCategory != "cancelled" {
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

	salesMonthly.AverageTicket = recalculateAverageTicket(salesMonthly.TotalRevenue, salesMonthly.TotalOrders, salesMonthly.CancelledOrders)
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
		return nil
	}

	// Mover contadores de estado (el struct anual tambien tiene CompletedOrders y CancelledOrders)
	switch oldCategory {
	case "completed":
		salesYearly.CompletedOrders--
	case "cancelled":
		salesYearly.CancelledOrders--
	}

	switch newCategory {
	case "completed":
		salesYearly.CompletedOrders++
	case "cancelled":
		salesYearly.CancelledOrders++
	}

	// Si se cancela: restar revenue
	if newCategory == "cancelled" {
		salesYearly.TotalRevenue -= order.Total
		salesYearly.TotalOrders--

		if mb, ok := salesYearly.MonthlyBreakdown[monthKey]; ok {
			mb.Revenue -= order.Total
			mb.Orders--
			salesYearly.MonthlyBreakdown[monthKey] = mb
		}
	}

	// Si se revierte cancelacion (defensivo)
	if oldCategory == "cancelled" && newCategory != "cancelled" {
		salesYearly.TotalRevenue += order.Total
		salesYearly.TotalOrders++

		mb := salesYearly.MonthlyBreakdown[monthKey]
		mb.Revenue += order.Total
		mb.Orders++
		salesYearly.MonthlyBreakdown[monthKey] = mb
	}

	salesYearly.AverageTicket = recalculateAverageTicket(salesYearly.TotalRevenue, salesYearly.TotalOrders, salesYearly.CancelledOrders)
	salesYearly.UpdatedAt = time.Now()

	if err := s.dashboardRepo.SaveSalesYearly(ctx, year, salesYearly); err != nil {
		return fmt.Errorf("error guardando sales_yearly: %w", err)
	}

	return nil
}

// ============================================================
// Metodos de Lectura (para endpoints GET)
// ============================================================

// GetSalesMonthly obtiene las metricas de ventas mensuales. Si no existe retorna documento vacio.
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

// GetSalesYearly obtiene las metricas de ventas anuales. Si no existe retorna documento vacio.
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

// GetBuyersMonthly obtiene las estadisticas de compradores mensuales. Si no existe retorna documento vacio.
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

// GetBuyersYearly obtiene las estadisticas de compradores anuales. Si no existe retorna documento vacio.
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

// GetTopProductsMonthly obtiene las metricas de top productos mensuales. Si no existe retorna documento vacio.
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

// GetTopProductsYearly obtiene las metricas de top productos anuales. Si no existe retorna documento vacio.
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

// GetSummary obtiene el resumen consolidado del dashboard para el mes y ano actual.
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

	// Convertir MostSold a DashboardSummaryTopProduct
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
// Recalculo Manual
// ============================================================

// RecalculateMonth reconstruye desde cero los documentos de metricas del mes y ano completo.
func (s *DashboardService) RecalculateMonth(ctx context.Context, year, month int) error {
	// Calcular rangos de fecha
	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	// Obtener todas las ordenes del mes
	monthlyOrders, err := s.orderRepo.GetOrdersByDateRange(ctx, startOfMonth, endOfMonth)
	if err != nil {
		return fmt.Errorf("error obteniendo ordenes del mes: %w", err)
	}

	// Obtener todas las ordenes del ano
	yearlyOrders, err := s.orderRepo.GetOrdersByDateRange(ctx, startOfYear, endOfYear)
	if err != nil {
		return fmt.Errorf("error obteniendo ordenes del ano: %w", err)
	}

	// Inicializar documentos vacios
	salesMonthly := newSalesMonthlyMetrics(year, month)
	buyersMonthly := newBuyersMonthlyMetrics(year, month)
	topProductsMonthly := newTopProductsMetrics(year, month)

	salesYearly := newSalesYearlyMetrics(year)
	buyersYearly := newBuyersYearlyMetrics(year)
	topProductsYearly := newTopProductsMetrics(year, 0)

	// --- Procesar ordenes mensuales ---
	for _, order := range monthlyOrders {
		items, itemsErr := s.orderRepo.GetItemsByOrderID(ctx, order.ID)
		if itemsErr != nil {
			log.Printf("Warning: no se pudieron obtener items de orden %s: %v", order.ID.String(), itemsErr)
			continue
		}

		s.recalcProcessOrder(ctx, order, items, salesMonthly, buyersMonthly, topProductsMonthly)
	}

	// --- Procesar ordenes anuales ---
	for _, order := range yearlyOrders {
		items, itemsErr := s.orderRepo.GetItemsByOrderID(ctx, order.ID)
		if itemsErr != nil {
			log.Printf("Warning: no se pudieron obtener items de orden %s: %v", order.ID.String(), itemsErr)
			continue
		}

		s.recalcProcessOrderYearly(ctx, order, items, salesYearly, buyersYearly, topProductsYearly)
	}

	// Recalcular average tickets
	salesMonthly.AverageTicket = recalculateAverageTicket(salesMonthly.TotalRevenue, salesMonthly.TotalOrders, salesMonthly.CancelledOrders)
	salesYearly.AverageTicket = recalculateAverageTicket(salesYearly.TotalRevenue, salesYearly.TotalOrders, salesYearly.CancelledOrders)

	// Recalcular returning buyers mensual
	buyersMonthly.ReturningBuyers = buyersMonthly.TotalBuyers - buyersMonthly.NewRegisteredThisMonth - buyersMonthly.GuestBuyers
	if buyersMonthly.ReturningBuyers < 0 {
		buyersMonthly.ReturningBuyers = 0
	}

	// Recalcular top products
	sortAndSliceTopProducts(topProductsMonthly)
	sortAndSliceTopProducts(topProductsYearly)

	// Actualizar timestamps
	now := time.Now()
	salesMonthly.UpdatedAt = now
	salesYearly.UpdatedAt = now
	buyersMonthly.UpdatedAt = now
	buyersYearly.UpdatedAt = now
	topProductsMonthly.UpdatedAt = now
	topProductsYearly.UpdatedAt = now

	// Guardar todos los documentos
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

// recalcProcessOrder procesa una orden individual para el recalculo mensual.
func (s *DashboardService) recalcProcessOrder(
	ctx context.Context,
	order *models.Order,
	items []*models.OrderItem,
	sales *models.SalesMonthlyMetrics,
	buyers *models.BuyersMonthlyMetrics,
	topProducts *models.TopProductsMetrics,
) {
	dayKey := order.CreatedAt.Format("2006-01-02")
	isCancelled := order.Status == models.OrderCancelled

	// --- Sales ---
	if !isCancelled {
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
	}

	// Contadores de estado
	switch statusCategory(order.Status) {
	case "pending":
		sales.PendingOrders++
	case "completed":
		sales.CompletedOrders++
	case "cancelled":
		sales.CancelledOrders++
	}

	// --- Buyers ---
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

	// --- Top Products (solo ordenes no canceladas) ---
	if !isCancelled {
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
}

// recalcProcessOrderYearly procesa una orden individual para el recalculo anual.
func (s *DashboardService) recalcProcessOrderYearly(
	ctx context.Context,
	order *models.Order,
	items []*models.OrderItem,
	sales *models.SalesYearlyMetrics,
	buyers *models.BuyersYearlyMetrics,
	topProducts *models.TopProductsMetrics,
) {
	monthKey := fmt.Sprintf("%02d", int(order.CreatedAt.Month()))
	isCancelled := order.Status == models.OrderCancelled

	// --- Sales ---
	if !isCancelled {
		sales.TotalRevenue += order.Total
		sales.TotalOrders++

		mb := sales.MonthlyBreakdown[monthKey]
		mb.Revenue += order.Total
		mb.Orders++
		sales.MonthlyBreakdown[monthKey] = mb
	}

	switch statusCategory(order.Status) {
	case "completed":
		sales.CompletedOrders++
	case "cancelled":
		sales.CancelledOrders++
	}

	// --- Buyers ---
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

	// --- Top Products (solo ordenes no canceladas) ---
	if !isCancelled {
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
}
