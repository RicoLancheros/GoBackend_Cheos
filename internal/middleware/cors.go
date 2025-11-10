package middleware

import (
	"strings"

	"github.com/cheoscafe/backend/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS configura el middleware de CORS
func CORS(cfg *config.Config) gin.HandlerFunc {
	// Parsear orígenes permitidos desde la configuración
	allowedOrigins := strings.Split(cfg.CORSAllowedOrigins, ",")

	// Limpiar espacios en blanco
	for i, origin := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(origin)
	}

	config := cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}

	return cors.New(config)
}
