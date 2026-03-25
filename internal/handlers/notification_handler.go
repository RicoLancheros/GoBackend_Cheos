package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cheoscafe/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	service *services.NotificationService
	hub     *SSEHub
}

func NewNotificationHandler(service *services.NotificationService, hub *SSEHub) *NotificationHandler {
	return &NotificationHandler{service: service, hub: hub}
}

// GetNotifications — GET /api/v1/notifications
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, err := extractUserID(c)
	if err != nil {
		log.Printf("[NotifHandler] extractUserID error: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "no autorizado"})
		return
	}

	log.Printf("[NotifHandler] GetNotifications userID=%s", userID)

	result, err := h.service.GetUserNotifications(c.Request.Context(), userID)
	if err != nil {
		log.Printf("[NotifHandler] GetUserNotifications error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "error al obtener notificaciones"})
		return
	}

	log.Printf("[NotifHandler] OK — %d notificaciones, %d no leídas", result.Total, result.UnreadCount)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// MarkRead — PATCH /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID, err := extractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "no autorizado"})
		return
	}

	notifID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id inválido"})
		return
	}

	if err := h.service.MarkRead(c.Request.Context(), notifID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "error al marcar notificación"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// MarkAllRead — PATCH /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID, err := extractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "no autorizado"})
		return
	}

	if err := h.service.MarkAllRead(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "error al marcar notificaciones"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// StreamNotifications — GET /api/v1/notifications/stream
// Usa el mismo AuthMiddleware que el resto del grupo (JWT en header Authorization).
func (h *NotificationHandler) StreamNotifications(c *gin.Context) {
	userID, err := extractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "no autorizado"})
		return
	}

	// Cabeceras SSE — antes de cualquier escritura al body
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // desactiva buffering en Nginx/proxies

	userIDStr := userID.String()
	ch := h.hub.Subscribe(userIDStr)
	defer h.hub.Unsubscribe(userIDStr, ch)

	// Enviar estado actual al conectarse (evento "sync")
	result, err := h.service.GetUserNotifications(c.Request.Context(), userID)
	if err == nil {
		if b, err := json.Marshal(result); err == nil {
			fmt.Fprintf(c.Writer, "event: sync\ndata: %s\n\n", b)
			c.Writer.Flush()
		}
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return

		case payload, open := <-ch:
			if !open {
				return
			}
			fmt.Fprintf(c.Writer, "event: notification\ndata: %s\n\n", payload)
			c.Writer.Flush()

		case <-ticker.C:
			fmt.Fprintf(c.Writer, ": heartbeat\n\n")
			c.Writer.Flush()
		}
	}
}

// extractUserID saca el userID del contexto de Gin (puesto por AuthMiddleware como "user_id")
func extractUserID(c *gin.Context) (uuid.UUID, error) {
	raw, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, fmt.Errorf("user_id no encontrado en contexto")
	}
	switch v := raw.(type) {
	case uuid.UUID:
		return v, nil
	case string:
		return uuid.Parse(v)
	}
	return uuid.Nil, fmt.Errorf("tipo de user_id desconocido: %T", raw)
}
