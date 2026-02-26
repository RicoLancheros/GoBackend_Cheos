package repository

import (
	"context"
	"fmt"

	"github.com/cheoscafe/backend/internal/database"
	"github.com/cheoscafe/backend/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const dashboardMetricsCollection = "dashboard_metrics"

type DashboardRepository struct {
	firebase *database.FirebaseClient
}

func NewDashboardRepository(firebase *database.FirebaseClient) *DashboardRepository {
	return &DashboardRepository{
		firebase: firebase,
	}
}

// --- Helpers para construir document IDs ---

func salesMonthlyDocID(year, month int) string {
	return fmt.Sprintf("sales_monthly_%04d-%02d", year, month)
}

func salesYearlyDocID(year int) string {
	return fmt.Sprintf("sales_yearly_%04d", year)
}

func buyersMonthlyDocID(year, month int) string {
	return fmt.Sprintf("buyers_monthly_%04d-%02d", year, month)
}

func buyersYearlyDocID(year int) string {
	return fmt.Sprintf("buyers_yearly_%04d", year)
}

func topProductsMonthlyDocID(year, month int) string {
	return fmt.Sprintf("top_products_monthly_%04d-%02d", year, month)
}

func topProductsYearlyDocID(year int) string {
	return fmt.Sprintf("top_products_yearly_%04d", year)
}

// --- Ventas Mensuales ---

// GetSalesMonthly obtiene las metricas de ventas mensuales. Retorna nil, nil si no existe.
func (r *DashboardRepository) GetSalesMonthly(ctx context.Context, year, month int) (*models.SalesMonthlyMetrics, error) {
	docID := salesMonthlyDocID(year, month)
	doc, err := r.firebase.Collection(dashboardMetricsCollection).Doc(docID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	var metrics models.SalesMonthlyMetrics
	if err := doc.DataTo(&metrics); err != nil {
		return nil, err
	}

	return &metrics, nil
}

// SaveSalesMonthly guarda/sobreescribe las metricas de ventas mensuales.
func (r *DashboardRepository) SaveSalesMonthly(ctx context.Context, year, month int, metrics *models.SalesMonthlyMetrics) error {
	docID := salesMonthlyDocID(year, month)
	_, err := r.firebase.Collection(dashboardMetricsCollection).Doc(docID).Set(ctx, metrics)
	return err
}

// --- Ventas Anuales ---

// GetSalesYearly obtiene las metricas de ventas anuales. Retorna nil, nil si no existe.
func (r *DashboardRepository) GetSalesYearly(ctx context.Context, year int) (*models.SalesYearlyMetrics, error) {
	docID := salesYearlyDocID(year)
	doc, err := r.firebase.Collection(dashboardMetricsCollection).Doc(docID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	var metrics models.SalesYearlyMetrics
	if err := doc.DataTo(&metrics); err != nil {
		return nil, err
	}

	return &metrics, nil
}

// SaveSalesYearly guarda/sobreescribe las metricas de ventas anuales.
func (r *DashboardRepository) SaveSalesYearly(ctx context.Context, year int, metrics *models.SalesYearlyMetrics) error {
	docID := salesYearlyDocID(year)
	_, err := r.firebase.Collection(dashboardMetricsCollection).Doc(docID).Set(ctx, metrics)
	return err
}

// --- Compradores Mensuales ---

// GetBuyersMonthly obtiene las estadisticas de compradores mensuales. Retorna nil, nil si no existe.
func (r *DashboardRepository) GetBuyersMonthly(ctx context.Context, year, month int) (*models.BuyersMonthlyMetrics, error) {
	docID := buyersMonthlyDocID(year, month)
	doc, err := r.firebase.Collection(dashboardMetricsCollection).Doc(docID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	var metrics models.BuyersMonthlyMetrics
	if err := doc.DataTo(&metrics); err != nil {
		return nil, err
	}

	return &metrics, nil
}

// SaveBuyersMonthly guarda/sobreescribe las estadisticas de compradores mensuales.
func (r *DashboardRepository) SaveBuyersMonthly(ctx context.Context, year, month int, metrics *models.BuyersMonthlyMetrics) error {
	docID := buyersMonthlyDocID(year, month)
	_, err := r.firebase.Collection(dashboardMetricsCollection).Doc(docID).Set(ctx, metrics)
	return err
}

// --- Compradores Anuales ---

// GetBuyersYearly obtiene las estadisticas de compradores anuales. Retorna nil, nil si no existe.
func (r *DashboardRepository) GetBuyersYearly(ctx context.Context, year int) (*models.BuyersYearlyMetrics, error) {
	docID := buyersYearlyDocID(year)
	doc, err := r.firebase.Collection(dashboardMetricsCollection).Doc(docID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	var metrics models.BuyersYearlyMetrics
	if err := doc.DataTo(&metrics); err != nil {
		return nil, err
	}

	return &metrics, nil
}

// SaveBuyersYearly guarda/sobreescribe las estadisticas de compradores anuales.
func (r *DashboardRepository) SaveBuyersYearly(ctx context.Context, year int, metrics *models.BuyersYearlyMetrics) error {
	docID := buyersYearlyDocID(year)
	_, err := r.firebase.Collection(dashboardMetricsCollection).Doc(docID).Set(ctx, metrics)
	return err
}

// --- Top Productos Mensuales ---

// GetTopProductsMonthly obtiene las metricas de top productos mensuales. Retorna nil, nil si no existe.
func (r *DashboardRepository) GetTopProductsMonthly(ctx context.Context, year, month int) (*models.TopProductsMetrics, error) {
	docID := topProductsMonthlyDocID(year, month)
	doc, err := r.firebase.Collection(dashboardMetricsCollection).Doc(docID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	var metrics models.TopProductsMetrics
	if err := doc.DataTo(&metrics); err != nil {
		return nil, err
	}

	return &metrics, nil
}

// SaveTopProductsMonthly guarda/sobreescribe las metricas de top productos mensuales.
func (r *DashboardRepository) SaveTopProductsMonthly(ctx context.Context, year, month int, metrics *models.TopProductsMetrics) error {
	docID := topProductsMonthlyDocID(year, month)
	_, err := r.firebase.Collection(dashboardMetricsCollection).Doc(docID).Set(ctx, metrics)
	return err
}

// --- Top Productos Anuales ---

// GetTopProductsYearly obtiene las metricas de top productos anuales. Retorna nil, nil si no existe.
func (r *DashboardRepository) GetTopProductsYearly(ctx context.Context, year int) (*models.TopProductsMetrics, error) {
	docID := topProductsYearlyDocID(year)
	doc, err := r.firebase.Collection(dashboardMetricsCollection).Doc(docID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	var metrics models.TopProductsMetrics
	if err := doc.DataTo(&metrics); err != nil {
		return nil, err
	}

	return &metrics, nil
}

// SaveTopProductsYearly guarda/sobreescribe las metricas de top productos anuales.
func (r *DashboardRepository) SaveTopProductsYearly(ctx context.Context, year int, metrics *models.TopProductsMetrics) error {
	docID := topProductsYearlyDocID(year)
	_, err := r.firebase.Collection(dashboardMetricsCollection).Doc(docID).Set(ctx, metrics)
	return err
}
