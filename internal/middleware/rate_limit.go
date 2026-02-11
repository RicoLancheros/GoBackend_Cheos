package middleware

import (
	"fmt"
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

// Rate limiters separados para evitar interferencias
var (
	// Rate limiter global
	globalVisitors   = make(map[string]*visitor)
	globalMu         sync.Mutex
	globalCleanupRun bool

	// Rate limiter específico para login
	loginVisitors   = make(map[string]*visitor)
	loginMu         sync.Mutex
	loginCleanupRun bool
)

// RateLimiter limita el número de requests por IP (uso global)
func RateLimiter(maxRequests int, duration time.Duration) gin.HandlerFunc {
	// Iniciar cleanup solo una vez
	globalMu.Lock()
	if !globalCleanupRun {
		globalCleanupRun = true
		go cleanupMap(&globalMu, globalVisitors, duration)
	}
	globalMu.Unlock()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		globalMu.Lock()
		v, exists := globalVisitors[ip]
		if !exists {
			globalVisitors[ip] = &visitor{
				lastSeen: time.Now(),
				count:    1,
			}
			globalMu.Unlock()
			c.Next()
			return
		}

		// Verificar si ha pasado el tiempo de la ventana
		if time.Since(v.lastSeen) > duration {
			v.lastSeen = time.Now()
			v.count = 1
			globalMu.Unlock()
			c.Next()
			return
		}

		// Incrementar contador
		v.count++
		count := v.count
		globalMu.Unlock()

		// Verificar límite
		if count > maxRequests {
			utils.ErrorResponse(c, http.StatusTooManyRequests, "Demasiadas solicitudes. Intenta de nuevo más tarde.", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoginRateLimiter límite específico para intentos de login (mapa separado)
// Permite 5 intentos fallidos cada 15 minutos por IP
func LoginRateLimiter() gin.HandlerFunc {
	maxAttempts := 5
	duration := 15 * time.Minute

	// Iniciar cleanup solo una vez
	loginMu.Lock()
	if !loginCleanupRun {
		loginCleanupRun = true
		go cleanupMap(&loginMu, loginVisitors, duration)
	}
	loginMu.Unlock()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		loginMu.Lock()
		v, exists := loginVisitors[ip]
		if !exists {
			loginVisitors[ip] = &visitor{
				lastSeen: time.Now(),
				count:    1,
			}
			loginMu.Unlock()
			c.Next()
			return
		}

		// Verificar si ha pasado el tiempo de la ventana
		if time.Since(v.lastSeen) > duration {
			v.lastSeen = time.Now()
			v.count = 1
			loginMu.Unlock()
			c.Next()
			return
		}

		// Incrementar contador
		v.count++
		count := v.count
		loginMu.Unlock()

		// Verificar límite
		if count > maxAttempts {
			remaining := duration - time.Since(v.lastSeen)
			minutes := int(remaining.Minutes()) + 1
			utils.ErrorResponse(c, http.StatusTooManyRequests,
				fmt.Sprintf("Demasiados intentos de inicio de sesión. Intenta de nuevo en %d minutos.", minutes),
				map[string]interface{}{
					"retry_after_seconds": int(remaining.Seconds()),
					"retry_after_minutes": minutes,
				})
			c.Abort()
			return
		}

		c.Next()
	}
}

// cleanupMap limpia visitantes antiguos de un mapa específico
func cleanupMap(mu *sync.Mutex, visitors map[string]*visitor, duration time.Duration) {
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
