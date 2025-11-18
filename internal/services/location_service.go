package services

import (
	"context"
	"errors"
	"math"

	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/repository"
	"github.com/google/uuid"
)

type LocationService struct {
	locationRepo *repository.LocationRepository
}

func NewLocationService(locationRepo *repository.LocationRepository) *LocationService {
	return &LocationService{
		locationRepo: locationRepo,
	}
}

// CreateLocation creates a new location
func (s *LocationService) CreateLocation(ctx context.Context, req *models.CreateLocationRequest) (*models.Location, error) {
	location := &models.Location{
		Name:       req.Name,
		Address:    req.Address,
		City:       req.City,
		Department: req.Department,
		Phone:      req.Phone,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		MapIframe:  req.MapIframe,
		Schedule:   req.Schedule,
		IsActive:   req.IsActive,
	}

	if err := s.locationRepo.Create(ctx, location); err != nil {
		return nil, err
	}

	return location, nil
}

// GetLocationByID gets a location by ID
func (s *LocationService) GetLocationByID(ctx context.Context, id uuid.UUID) (*models.Location, error) {
	location, err := s.locationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("location not found")
	}

	return location, nil
}

// GetAllLocations gets all locations with pagination
func (s *LocationService) GetAllLocations(ctx context.Context, page int, pageSize int) (*models.LocationListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	locations, err := s.locationRepo.GetAll(ctx, pageSize, offset)
	if err != nil {
		return nil, err
	}

	total, err := s.locationRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	// Convert pointers to values for response
	locationList := make([]models.Location, 0, len(locations))
	for _, loc := range locations {
		locationList = append(locationList, *loc)
	}

	return &models.LocationListResponse{
		Locations:  locationList,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetActiveLocations gets all active locations (for public display)
func (s *LocationService) GetActiveLocations(ctx context.Context) ([]models.Location, error) {
	locations, err := s.locationRepo.GetActive(ctx)
	if err != nil {
		return nil, err
	}

	// Convert pointers to values for response
	locationList := make([]models.Location, 0, len(locations))
	for _, loc := range locations {
		locationList = append(locationList, *loc)
	}

	return locationList, nil
}

// UpdateLocation updates a location
func (s *LocationService) UpdateLocation(ctx context.Context, id uuid.UUID, req *models.UpdateLocationRequest) (*models.Location, error) {
	existing, err := s.locationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("location not found")
	}

	// Update fields
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Address != "" {
		existing.Address = req.Address
	}
	if req.City != "" {
		existing.City = req.City
	}
	if req.Department != "" {
		existing.Department = req.Department
	}
	if req.Phone != "" {
		existing.Phone = req.Phone
	}
	if req.Latitude != 0 {
		existing.Latitude = req.Latitude
	}
	if req.Longitude != 0 {
		existing.Longitude = req.Longitude
	}
	if req.MapIframe != "" {
		existing.MapIframe = req.MapIframe
	}
	if req.Schedule != nil {
		existing.Schedule = req.Schedule
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := s.locationRepo.Update(ctx, id, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// DeleteLocation soft deletes a location
func (s *LocationService) DeleteLocation(ctx context.Context, id uuid.UUID) error {
	_, err := s.locationRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("location not found")
	}

	if err := s.locationRepo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}
