package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/cheoscafe/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type visitor struct {
	lastSeen time.Time
	count    int
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

// RateLimiter limita el número de requests por IP
func RateLimiter(maxRequests int, duration time.Duration) gin.HandlerFunc {
	// Limpiar visitantes viejos cada hora
	go cleanupVisitors(duration)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		v, exists := visitors[ip]
		if !exists {
			visitors[ip] = &visitor{
				lastSeen: time.Now(),
				count:    1,
			}
			mu.Unlock()
			c.Next()
			return
		}

		// Verificar si ha pasado el tiempo de la ventana
		if time.Since(v.lastSeen) > duration {
			v.lastSeen = time.Now()
			v.count = 1
			mu.Unlock()
			c.Next()
			return
		}

		// Incrementar contador
		v.count++
		count := v.count
		mu.Unlock()

		// Verificar límite
		if count > maxRequests {
			utils.ErrorResponse(c, http.StatusTooManyRequests, "Demasiadas solicitudes. Intenta de nuevo más tarde.", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoginRateLimiter límite específico para intentos de login
func LoginRateLimiter() gin.HandlerFunc {
	return RateLimiter(20, 15*time.Minute)
}

// cleanupVisitors limpia visitantes antiguos
func cleanupVisitors(duration time.Duration) {
	for {
		time.Sleep(duration)
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > duration {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}
