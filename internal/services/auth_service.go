package services

import (
	"context"
	"errors"
	"time"

	"github.com/cheoscafe/backend/internal/config"
	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/repository"
	"github.com/cheoscafe/backend/internal/utils"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// Register registra un nuevo usuario
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	// Verificar si el email ya existe
	exists, err := s.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("el email ya está registrado")
	}

	// Hashear la contraseña
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Determinar rol (por defecto CUSTOMER)
	role := models.RoleCustomer
	if req.Role != "" {
		role = req.Role
	}

	// Crear usuario
	user := &models.User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
		Phone:    req.Phone,
		Role:     role,
		IsActive: true,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	// Limpiar contraseña antes de devolver
	user.Password = ""

	return user, nil
}

// Login autentica un usuario
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// Buscar usuario por email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("credenciales inválidas")
	}

	// Verificar que el usuario esté activo
	if !user.IsActive {
		return nil, errors.New("usuario inactivo")
	}

	// Verificar contraseña
	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, errors.New("credenciales inválidas")
	}

	// Generar tokens
	accessTokenDuration, err := utils.ParseDuration(s.cfg.JWTExpiresIn)
	if err != nil {
		accessTokenDuration = 15 * time.Minute
	}

	refreshTokenDuration, err := utils.ParseDuration(s.cfg.JWTRefreshExpiresIn)
	if err != nil {
		refreshTokenDuration = 168 * time.Hour // 7 días
	}

	accessToken, err := utils.GenerateToken(
		user.ID,
		user.Email,
		string(user.Role),
		s.cfg.JWTSecret,
		accessTokenDuration,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.GenerateToken(
		user.ID,
		user.Email,
		string(user.Role),
		s.cfg.JWTRefreshSecret,
		refreshTokenDuration,
	)
	if err != nil {
		return nil, err
	}

	// Limpiar contraseña antes de devolver
	user.Password = ""

	return &models.LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken renueva el token de acceso
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// Validar refresh token
	claims, err := utils.ValidateToken(refreshToken, s.cfg.JWTRefreshSecret)
	if err != nil {
		return "", errors.New("refresh token inválido o expirado")
	}

	// Verificar que el usuario existe y está activo
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return "", errors.New("usuario no encontrado")
	}

	if !user.IsActive {
		return "", errors.New("usuario inactivo")
	}

	// Generar nuevo access token
	accessTokenDuration, err := utils.ParseDuration(s.cfg.JWTExpiresIn)
	if err != nil {
		accessTokenDuration = 15 * time.Minute
	}

	accessToken, err := utils.GenerateToken(
		user.ID,
		user.Email,
		string(user.Role),
		s.cfg.JWTSecret,
		accessTokenDuration,
	)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// GetProfile obtiene el perfil del usuario autenticado
func (s *AuthService) GetProfile(ctx context.Context, userID string) (*models.User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("ID de usuario inválido")
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Limpiar contraseña
	user.Password = ""

	return user, nil
}

// UpdateProfile actualiza el perfil del usuario
func (s *AuthService) UpdateProfile(ctx context.Context, userID string, req *models.UpdateProfileRequest) (*models.User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("ID de usuario inválido")
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Actualizar campos si se proporcionan
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	// Limpiar contraseña
	user.Password = ""

	return user, nil
}

// GetAllUsers obtiene todos los usuarios (solo admin)
func (s *AuthService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	users, err := s.userRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Limpiar contraseñas
	for _, user := range users {
		user.Password = ""
	}

	return users, nil
}

// UpdateUserByID actualiza cualquier usuario por ID (solo admin)
func (s *AuthService) UpdateUserByID(ctx context.Context, userID string, req *models.UpdateUserByIDRequest) (*models.User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("ID de usuario inválido")
	}

	// Obtener usuario existente
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Actualizar campos si se proporcionan
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Email != nil {
		// Verificar que el nuevo email no exista
		if *req.Email != user.Email {
			exists, err := s.userRepo.EmailExists(ctx, *req.Email)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, errors.New("el email ya está registrado")
			}
			user.Email = *req.Email
		}
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.Password != nil && *req.Password != "" {
		// Hashear nueva contraseña
		hashedPassword, err := utils.HashPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		user.Password = hashedPassword
	}

	// Actualizar usuario
	err = s.userRepo.UpdateByID(ctx, id, user)
	if err != nil {
		return nil, err
	}

	// Limpiar contraseña antes de devolver
	user.Password = ""

	return user, nil
}

// DeleteUser elimina un usuario por ID (soft delete, solo admin)
func (s *AuthService) DeleteUser(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("ID de usuario inválido")
	}

	// Verificar que el usuario existe
	_, err = s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Eliminar usuario (soft delete)
	return s.userRepo.Delete(ctx, id)
}
