package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/repository"
	"github.com/google/uuid"
)

type CartService struct {
	cartRepo    *repository.CartRepository
	productRepo *repository.ProductRepository
}

func NewCartService(cartRepo *repository.CartRepository, productRepo *repository.ProductRepository) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

// GetCart obtiene el carrito de un usuario
func (s *CartService) GetCart(ctx context.Context, userID uuid.UUID) (*models.Cart, error) {
	return s.cartRepo.GetByUserID(ctx, userID)
}

// AddItem agrega un producto al carrito o suma cantidad si ya existe.
// Solo valida que el producto exista y que haya stock suficiente.
// is_active no se verifica aquí — un producto puede estar "inactivo"
// visualmente pero aún tener stock comprable.
func (s *CartService) AddItem(ctx context.Context, userID uuid.UUID, req *models.AddToCartRequest) (*models.Cart, error) {
	product, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, errors.New("producto no encontrado")
	}

	cart, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Calcular cantidad total resultante para validar contra stock
	newQuantity := req.Quantity
	for _, item := range cart.Items {
		if item.ProductID == req.ProductID {
			newQuantity += item.Quantity
			break
		}
	}

	if product.Stock < newQuantity {
		return nil, fmt.Errorf("solo hay %d unidad(es) disponible(s) de %s", product.Stock, product.Name)
	}

	// Buscar si el producto ya está en el carrito
	found := false
	for i, item := range cart.Items {
		if item.ProductID == req.ProductID {
			cart.Items[i].Quantity += req.Quantity
			found = true
			break
		}
	}

	if !found {
		productImage := ""
		if len(product.Images) > 0 {
			productImage = product.Images[0]
		}

		cart.Items = append(cart.Items, models.CartItem{
			ProductID:    product.ID,
			ProductName:  product.Name,
			ProductPrice: product.Price,
			ProductImage: productImage,
			Quantity:     req.Quantity,
		})
	}

	if err := s.cartRepo.Save(ctx, cart); err != nil {
		return nil, err
	}

	return cart, nil
}

// UpdateItemQuantity actualiza la cantidad de un item en el carrito.
// Valida que la nueva cantidad no supere el stock disponible.
func (s *CartService) UpdateItemQuantity(ctx context.Context, userID uuid.UUID, productID uuid.UUID, req *models.UpdateCartItemRequest) (*models.Cart, error) {
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, errors.New("producto no encontrado")
	}

	if product.Stock < req.Quantity {
		return nil, fmt.Errorf("solo hay %d unidad(es) disponible(s) de %s", product.Stock, product.Name)
	}

	cart, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	found := false
	for i, item := range cart.Items {
		if item.ProductID == productID {
			cart.Items[i].Quantity = req.Quantity
			found = true
			break
		}
	}

	if !found {
		return nil, errors.New("producto no encontrado en el carrito")
	}

	if err := s.cartRepo.Save(ctx, cart); err != nil {
		return nil, err
	}

	return cart, nil
}

// RemoveItem elimina un producto del carrito
func (s *CartService) RemoveItem(ctx context.Context, userID uuid.UUID, productID uuid.UUID) (*models.Cart, error) {
	cart, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	newItems := make([]models.CartItem, 0)
	for _, item := range cart.Items {
		if item.ProductID != productID {
			newItems = append(newItems, item)
		}
	}
	cart.Items = newItems

	if err := s.cartRepo.Save(ctx, cart); err != nil {
		return nil, err
	}

	return cart, nil
}

// ClearCart vacía el carrito de un usuario
func (s *CartService) ClearCart(ctx context.Context, userID uuid.UUID) error {
	return s.cartRepo.Delete(ctx, userID)
}

// SyncCart fusiona el carrito local (invitado) con el guardado en Firebase.
// Ignora productos que no existen. Respeta el stock — si la cantidad
// fusionada supera el stock, se clampea al máximo disponible.
func (s *CartService) SyncCart(ctx context.Context, userID uuid.UUID, req *models.SyncCartRequest) (*models.Cart, error) {
	cart, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, reqItem := range req.Items {
		product, err := s.productRepo.GetByID(ctx, reqItem.ProductID)
		if err != nil {
			continue // Producto no existe, ignorar
		}

		// Buscar si ya existe en el carrito guardado
		found := false
		for i, item := range cart.Items {
			if item.ProductID == reqItem.ProductID {
				merged := item.Quantity + reqItem.Quantity
				// Clampear al stock disponible
				if merged > product.Stock {
					merged = product.Stock
				}
				cart.Items[i].Quantity = merged
				// Actualizar datos del producto por si cambiaron
				cart.Items[i].ProductName = product.Name
				cart.Items[i].ProductPrice = product.Price
				if len(product.Images) > 0 {
					cart.Items[i].ProductImage = product.Images[0]
				}
				found = true
				break
			}
		}

		if !found {
			// Solo agregar si hay stock
			if product.Stock <= 0 {
				continue
			}

			qty := reqItem.Quantity
			if qty > product.Stock {
				qty = product.Stock
			}

			productImage := ""
			if len(product.Images) > 0 {
				productImage = product.Images[0]
			}

			cart.Items = append(cart.Items, models.CartItem{
				ProductID:    product.ID,
				ProductName:  product.Name,
				ProductPrice: product.Price,
				ProductImage: productImage,
				Quantity:     qty,
			})
		}
	}

	if err := s.cartRepo.Save(ctx, cart); err != nil {
		return nil, err
	}

	return cart, nil
}
