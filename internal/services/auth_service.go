package services

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/cheoscafe/backend/internal/config"
	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/repository"
	"github.com/cheoscafe/backend/internal/utils"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo             *repository.UserRepository
	cfg                  *config.Config
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewAuthService(userRepo *repository.UserRepository, cfg *config.Config) *AuthService {
	// Pre-parsear duraciones una sola vez al startup en lugar de en cada login.
	accessDur, err := utils.ParseDuration(cfg.JWTExpiresIn)
	if err != nil {
		accessDur = 15 * time.Minute
	}
	refreshDur, err := utils.ParseDuration(cfg.JWTRefreshExpiresIn)
	if err != nil {
		refreshDur = 168 * time.Hour
	}

	return &AuthService{
		userRepo:             userRepo,
		cfg:                  cfg,
		accessTokenDuration:  accessDur,
		refreshTokenDuration: refreshDur,
	}
}

// Register registra un nuevo usuario
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	exists, err := s.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("el email ya está registrado")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	role := models.RoleCustomer
	if req.Role != "" {
		role = req.Role
	}

	user := &models.User{
		Email:        req.Email,
		Password:     hashedPassword,
		Name:         req.Name,
		Phone:        req.Phone,
		City:         req.City,
		Municipality: req.Municipality,
		Neighborhood: req.Neighborhood,
		Gender:       req.Gender,
		BirthDate:    req.BirthDate,
		Role:         role,
		IsActive:     true,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}

// Login autentica un usuario.
// Instrumentado con timing logs para identificar cuellos de botella bajo carga.
// Pasos: 1) Firestore GetByEmail  2) IsActive check  3) bcrypt compare  4) JWT x2
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	start := time.Now()
	log.Printf("[LOGIN] start email=%s", req.Email)

	// Paso 1: lectura desde Firestore (1 doc via Where+Limit(1))
	tStep := time.Now()
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	firestoreElapsed := time.Since(tStep)
	if err != nil {
		log.Printf("[LOGIN] firestore_fail email=%s firestore=%v total=%v", req.Email, firestoreElapsed, time.Since(start))
		return nil, errors.New("credenciales inválidas")
	}
	log.Printf("[LOGIN] firestore_ok email=%s firestore=%v", req.Email, firestoreElapsed)

	if !user.IsActive {
		log.Printf("[LOGIN] inactive email=%s total=%v", req.Email, time.Since(start))
		return nil, errors.New("usuario inactivo")
	}

	// Paso 2: bcrypt compare — históricamente el bottleneck principal en CPU compartida.
	tStep = time.Now()
	if !utils.CheckPassword(req.Password, user.Password) {
		log.Printf("[LOGIN] bcrypt_fail email=%s bcrypt=%v total=%v", req.Email, time.Since(tStep), time.Since(start))
		return nil, errors.New("credenciales inválidas")
	}
	bcryptElapsed := time.Since(tStep)
	log.Printf("[LOGIN] bcrypt_ok email=%s bcrypt=%v", req.Email, bcryptElapsed)

	// Paso 3: generación de tokens (HS256 — ~µs, no es bottleneck)
	tStep = time.Now()
	accessToken, err := utils.GenerateToken(
		user.ID, user.Email, string(user.Role),
		s.cfg.JWTSecret, s.accessTokenDuration,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.GenerateToken(
		user.ID, user.Email, string(user.Role),
		s.cfg.JWTRefreshSecret, s.refreshTokenDuration,
	)
	if err != nil {
		return nil, err
	}
	jwtElapsed := time.Since(tStep)

	totalElapsed := time.Since(start)
	log.Printf("[LOGIN] done email=%s total=%v firestore=%v bcrypt=%v jwt=%v",
		req.Email, totalElapsed, firestoreElapsed, bcryptElapsed, jwtElapsed)

	user.Password = ""
	return &models.LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken renueva el token de acceso
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := utils.ValidateToken(refreshToken, s.cfg.JWTRefreshSecret)
	if err != nil {
		return "", errors.New("refresh token inválido o expirado")
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return "", errors.New("usuario no encontrado")
	}

	if !user.IsActive {
		return "", errors.New("usuario inactivo")
	}

	return utils.GenerateToken(
		user.ID, user.Email, string(user.Role),
		s.cfg.JWTSecret, s.accessTokenDuration,
	)
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

	// Campos no-nullable
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	// Campos nullable: asignar SIEMPRE sin guard
	user.City = req.City
	user.Municipality = req.Municipality
	user.Neighborhood = req.Neighborhood
	user.BirthDate = req.BirthDate
	user.Gender = req.Gender

	log.Printf("[DEBUG UpdateProfile] user antes de guardar: city=%v municipality=%v neighborhood=%v gender=%v birth_date=%v",
		user.City, user.Municipality, user.Neighborhood, user.Gender, user.BirthDate)

	if err = s.userRepo.Update(ctx, user); err != nil {
		log.Printf("[DEBUG UpdateProfile] ERROR en Update: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG UpdateProfile] Update exitoso, leyendo de Firestore...")

	// Leer de Firestore para devolver estado real
	saved, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG UpdateProfile] saved desde Firestore: city=%v municipality=%v neighborhood=%v gender=%v birth_date=%v",
		saved.City, saved.Municipality, saved.Neighborhood, saved.Gender, saved.BirthDate)

	saved.Password = ""
	return saved, nil
}

// GetAllUsers obtiene todos los usuarios (solo admin)
func (s *AuthService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	users, err := s.userRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

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

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Email != nil && *req.Email != user.Email {
		exists, err := s.userRepo.EmailExists(ctx, *req.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("el email ya está registrado")
		}
		user.Email = *req.Email
	}
	if req.City != nil {
		user.City = req.City
	}
	if req.Municipality != nil {
		user.Municipality = req.Municipality
	}
	if req.Neighborhood != nil {
		user.Neighborhood = req.Neighborhood
	}
	if req.Gender != nil {
		user.Gender = req.Gender
	}
	if req.BirthDate != nil {
		user.BirthDate = req.BirthDate
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.Password != nil && *req.Password != "" {
		hashed, err := utils.HashPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		user.Password = hashed
	}

	if err = s.userRepo.UpdateByID(ctx, id, user); err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}

// DeleteUser elimina un usuario por ID (solo admin)
func (s *AuthService) DeleteUser(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("ID de usuario inválido")
	}

	_, err = s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return s.userRepo.Delete(ctx, id)
}
