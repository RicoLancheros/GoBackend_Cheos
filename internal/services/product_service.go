package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/repository"
	"github.com/cheoscafe/backend/internal/utils"
	"github.com/google/uuid"
)

type ProductService struct {
	productRepo *repository.ProductRepository
}

func NewProductService(productRepo *repository.ProductRepository) *ProductService {
	return &ProductService{
		productRepo: productRepo,
	}
}

// CreateProduct crea un nuevo producto (solo admin)
func (s *ProductService) CreateProduct(ctx context.Context, req *models.CreateProductRequest) (*models.Product, error) {
	// Validar request
	if err := utils.ValidateStruct(req); err != nil {
		return nil, err
	}

	// Validar precio
	if req.Price <= 0 {
		return nil, errors.New("el precio debe ser mayor a 0")
	}

	// Validar stock
	if req.Stock < 0 {
		return nil, errors.New("el stock no puede ser negativo")
	}

	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
		Weight:      req.Weight,
		Images:      req.Images,
		IsFeatured:  req.IsFeatured,
		IsActive:    true,
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("error al crear producto: %w", err)
	}

	return product, nil
}

// GetProduct obtiene un producto por ID
func (s *ProductService) GetProduct(ctx context.Context, productID string) (*models.Product, error) {
	id, err := uuid.Parse(productID)
	if err != nil {
		return nil, errors.New("ID de producto inválido")
	}

	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return product, nil
}

// GetAllProducts obtiene todos los productos con paginación
func (s *ProductService) GetAllProducts(ctx context.Context, page, pageSize int) (*models.PaginatedProductsResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	products, err := s.productRepo.GetAll(ctx, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("error al obtener productos: %w", err)
	}

	total, err := s.productRepo.CountProducts(ctx)
	if err != nil {
		total = int64(len(products)) // Fallback
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &models.PaginatedProductsResponse{
		Products:    products,
		Total:       int(total),
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}, nil
}

// GetFeaturedProducts obtiene productos destacados
func (s *ProductService) GetFeaturedProducts(ctx context.Context) ([]*models.Product, error) {
	products, err := s.productRepo.GetFeatured(ctx, 8)
	if err != nil {
		return nil, fmt.Errorf("error al obtener productos destacados: %w", err)
	}

	return products, nil
}

// UpdateProduct actualiza un producto (solo admin)
func (s *ProductService) UpdateProduct(ctx context.Context, productID string, req *models.UpdateProductRequest) (*models.Product, error) {
	id, err := uuid.Parse(productID)
	if err != nil {
		return nil, errors.New("ID de producto inválido")
	}

	// Validar request
	if err := utils.ValidateStruct(req); err != nil {
		return nil, err
	}

	// Obtener producto existente
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Actualizar campos
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		if *req.Price <= 0 {
			return nil, errors.New("el precio debe ser mayor a 0")
		}
		product.Price = *req.Price
	}
	if req.Stock != nil {
		if *req.Stock < 0 {
			return nil, errors.New("el stock no puede ser negativo")
		}
		product.Stock = *req.Stock
	}
	if req.Category != nil {
		product.Category = *req.Category
	}
	if req.Weight != nil {
		product.Weight = *req.Weight
	}
	if req.Images != nil {
		product.Images = req.Images
	}
	if req.IsFeatured != nil {
		product.IsFeatured = *req.IsFeatured
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("error al actualizar producto: %w", err)
	}

	return product, nil
}

// DeleteProduct elimina un producto (soft delete, solo admin)
func (s *ProductService) DeleteProduct(ctx context.Context, productID string) error {
	id, err := uuid.Parse(productID)
	if err != nil {
		return errors.New("ID de producto inválido")
	}

	if err := s.productRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("error al eliminar producto: %w", err)
	}

	return nil
}

// UpdateStock actualiza el stock de un producto
func (s *ProductService) UpdateStock(ctx context.Context, productID string, quantity int) error {
	id, err := uuid.Parse(productID)
	if err != nil {
		return errors.New("ID de producto inválido")
	}

	if err := s.productRepo.UpdateStock(ctx, id, quantity); err != nil {
		return fmt.Errorf("error al actualizar stock: %w", err)
	}

	return nil
}

// SearchProducts busca productos
func (s *ProductService) SearchProducts(ctx context.Context, searchTerm string, limit int) ([]*models.Product, error) {
	if searchTerm == "" {
		return nil, errors.New("término de búsqueda requerido")
	}

	if limit < 1 || limit > 100 {
		limit = 10
	}

	products, err := s.productRepo.Search(ctx, searchTerm, limit)
	if err != nil {
		return nil, fmt.Errorf("error al buscar productos: %w", err)
	}

	return products, nil
}
