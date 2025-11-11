package handlers

import (
	"net/http"

	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/services"
	"github.com/cheoscafe/backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register maneja el registro de usuarios
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Validar
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.FormatValidationErrors(err))
		return
	}

	// Registrar
	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Error al registrar usuario", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Usuario registrado exitosamente", user)
}

// Login maneja el login de usuarios
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Validar
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.FormatValidationErrors(err))
		return
	}

	// Login
	response, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Credenciales inválidas", err.Error())
		return
	}

	// Establecer cookies HttpOnly
	// Access Token - expira en 15 minutos (900 segundos)
	c.SetCookie(
		"access_token",              // nombre
		response.AccessToken,        // valor
		900,                         // maxAge en segundos (15 minutos)
		"/",                         // path
		"",                          // domain (vacío = dominio actual)
		false,                       // secure (false en desarrollo, true en producción)
		true,                        // httpOnly (JavaScript no puede acceder)
	)

	// Refresh Token - expira en 7 días (604800 segundos)
	c.SetCookie(
		"refresh_token",             // nombre
		response.RefreshToken,       // valor
		604800,                      // maxAge en segundos (7 días)
		"/",                         // path
		"",                          // domain
		false,                       // secure
		true,                        // httpOnly
	)

	utils.SuccessResponse(c, http.StatusOK, "Login exitoso", response)
}

// RefreshToken maneja la renovación de tokens
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Refresh token requerido", err.Error())
		return
	}

	// Renovar token
	accessToken, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Refresh token inválido", err.Error())
		return
	}

	// Establecer nueva cookie con el access token renovado
	c.SetCookie(
		"access_token",              // nombre
		accessToken,                 // valor
		900,                         // maxAge en segundos (15 minutos)
		"/",                         // path
		"",                          // domain
		false,                       // secure
		true,                        // httpOnly
	)

	utils.SuccessResponse(c, http.StatusOK, "Token renovado exitosamente", gin.H{
		"access_token": accessToken,
	})
}

// GetProfile obtiene el perfil del usuario autenticado
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "No autenticado", nil)
		return
	}

	user, err := h.authService.GetProfile(c.Request.Context(), userID.(uuid.UUID).String())
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Usuario no encontrado", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Perfil obtenido exitosamente", user)
}

// UpdateProfile actualiza el perfil del usuario
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "No autenticado", nil)
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Validar
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, utils.FormatValidationErrors(err))
		return
	}

	user, err := h.authService.UpdateProfile(c.Request.Context(), userID.(uuid.UUID).String(), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Error al actualizar perfil", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Perfil actualizado exitosamente", user)
}
