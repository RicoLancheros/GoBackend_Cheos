package middleware

import (
	"net/http"
	"strings"

	"github.com/cheoscafe/backend/internal/config"
	"github.com/cheoscafe/backend/internal/models"
	"github.com/cheoscafe/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware verifica el token JWT
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtener el header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Token de autenticación requerido", nil)
			c.Abort()
			return
		}

		// Verificar que tenga el formato "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Formato de token inválido", nil)
			c.Abort()
			return
		}

		token := parts[1]

		// Validar el token
		claims, err := utils.ValidateToken(token, cfg.JWTSecret)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Token inválido o expirado", err.Error())
			c.Abort()
			return
		}

		// Guardar la información del usuario en el contexto
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// RequireAdmin verifica que el usuario sea admin
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "No autenticado", nil)
			c.Abort()
			return
		}

		if role != string(models.RoleAdmin) {
			utils.ErrorResponse(c, http.StatusForbidden, "Se requieren permisos de administrador", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth permite acceso con o sin token
func OptionalAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No hay token, continuar sin autenticación
			c.Next()
			return
		}

		// Si hay token, validarlo
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			token := parts[1]
			claims, err := utils.ValidateToken(token, cfg.JWTSecret)
			if err == nil {
				// Token válido, guardar información
				c.Set("user_id", claims.UserID)
				c.Set("user_email", claims.Email)
				c.Set("user_role", claims.Role)
			}
		}

		c.Next()
	}
}
