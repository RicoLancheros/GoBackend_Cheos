// internal/handlers/wompi_handler.go
// Cheos Café — Wompi Payment Handler

package handlers

import (
	"net/http"

	"github.com/cheoscafe/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type WompiHandler struct {
	wompiService *services.WompiService
}

func NewWompiHandler(wompiService *services.WompiService) *WompiHandler {
	return &WompiHandler{wompiService: wompiService}
}

// POST /api/v1/payments/wompi/signature
// Genera la firma de integridad requerida por Wompi
func (h *WompiHandler) GenerateSignature(c *gin.Context) {
	var req services.SignatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Datos inválidos: " + err.Error(),
		})
		return
	}

	resp, err := h.wompiService.GenerateIntegritySignature(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "signature_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GET /api/v1/payments/wompi/transaction/:id
// Consulta el estado de una transacción directamente en Wompi
func (h *WompiHandler) GetTransaction(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_id",
			"message": "ID de transacción requerido",
		})
		return
	}

	result, err := h.wompiService.GetTransaction(transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "transaction_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// POST /api/v1/payments/wompi/webhook
// Recibe notificaciones de eventos de Wompi (server-to-server)
func (h *WompiHandler) HandleWebhook(c *gin.Context) {
	var payload services.WompiWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid_payload",
		})
		return
	}

	// Verificar firma del webhook
	if err := h.wompiService.VerifyWebhookSignature(payload, c.GetHeader("X-Wompi-Signature")); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid_signature",
			"message": "Firma del webhook inválida",
		})
		return
	}

	// Procesar el evento
	if err := h.wompiService.ProcessWebhookEvent(payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "processing_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}
