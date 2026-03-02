package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cheoscafe/backend/internal/config"
	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/repository"
	"github.com/cheoscafe/backend/internal/utils"
	"github.com/google/uuid"
)

type PasswordResetService struct {
	resetRepo    *repository.PasswordResetRepository
	userRepo     *repository.UserRepository
	emailService *EmailService
	cfg          *config.Config
}

func NewPasswordResetService(
	resetRepo *repository.PasswordResetRepository,
	userRepo *repository.UserRepository,
	emailService *EmailService,
	cfg *config.Config,
) *PasswordResetService {
	return &PasswordResetService{
		resetRepo:    resetRepo,
		userRepo:     userRepo,
		emailService: emailService,
		cfg:          cfg,
	}
}

// ForgotPassword genera un token de reset y envia el email al usuario
func (s *PasswordResetService) ForgotPassword(ctx context.Context, email string) error {
	// Buscar usuario por email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Por seguridad, NO revelar si el email existe o no
		// Retornar nil para que el handler siempre responda 200
		return nil
	}

	// Eliminar tokens previos del usuario (ignorar error si no hay tokens)
	_ = s.resetRepo.DeleteByUserID(ctx, user.ID)

	// Generar token JWT con role "password_reset" para distinguirlo de tokens de sesion
	token, err := utils.GenerateToken(
		user.ID,
		user.Email,
		"password_reset",
		s.cfg.JWTSecret,
		15*time.Minute,
	)
	if err != nil {
		return fmt.Errorf("error generando token de reset: %w", err)
	}

	// Crear registro en Firestore
	now := time.Now()
	resetDoc := &models.PasswordReset{
		ID:        uuid.New(),
		UserID:    user.ID,
		Email:     user.Email,
		Token:     token,
		Used:      false,
		CreatedAt: now,
		ExpiresAt: now.Add(15 * time.Minute),
	}

	if err := s.resetRepo.Create(ctx, resetDoc); err != nil {
		return fmt.Errorf("error guardando token de reset: %w", err)
	}

	// Enviar email en goroutine (no bloquear la respuesta HTTP)
	// Si el email falla, el usuario puede reintentar pidiendo otro token
	go s.emailService.SendPasswordResetEmail(user.Email, user.Name, token)

	return nil
}

// ResetPassword valida el token y actualiza la contrasena del usuario
func (s *PasswordResetService) ResetPassword(ctx context.Context, token string, newPassword string) error {
	// 1. Validar el token JWT
	claims, err := utils.ValidateToken(token, s.cfg.JWTSecret)
	if err != nil {
		return errors.New("token invalido o expirado")
	}

	// 2. Buscar el token en Firestore
	resetDoc, err := s.resetRepo.GetByToken(ctx, token)
	if err != nil {
		return errors.New("token no valido o ya fue utilizado")
	}

	// 3. Verificar que no haya sido usado
	if resetDoc.Used {
		return errors.New("este token ya fue utilizado")
	}

	// 4. Verificar que no haya expirado (doble verificacion: JWT + campo ExpiresAt)
	if time.Now().After(resetDoc.ExpiresAt) {
		return errors.New("el token ha expirado")
	}

	// 5. Buscar el usuario por ID (del JWT claims)
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return errors.New("usuario no encontrado")
	}

	// 6. Hashear la nueva contrasena
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("error hasheando contrasena: %w", err)
	}

	// 7. Actualizar la contrasena del usuario
	// Se usa UpdateByID (Set completo) porque Update parcial no incluye el campo password
	user.Password = hashedPassword
	if err := s.userRepo.UpdateByID(ctx, user.ID, user); err != nil {
		return fmt.Errorf("error actualizando contrasena: %w", err)
	}

	// 8. Eliminar el token de Firestore (hard delete, un solo uso)
	if err := s.resetRepo.Delete(ctx, resetDoc.ID); err != nil {
		// Log del error pero no fallar - la contrasena ya fue actualizada
		fmt.Printf("advertencia: error eliminando token de reset %s: %v\n", resetDoc.ID, err)
	}

	return nil
}
