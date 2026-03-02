// internal/services/wompi_service.go
// Cheos Café — Wompi Service (corregido)

package services

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ─── Structs ──────────────────────────────────────────────────────────────────

type SignatureRequest struct {
	Reference    string `json:"reference" binding:"required"`
	AmountCents  int64  `json:"amount_in_cents" binding:"required,min=1"`
	Currency     string `json:"currency" binding:"required"`
	ExpirationAt string `json:"expiration_time,omitempty"`
}

type SignatureResponse struct {
	Signature    string `json:"signature"`
	Reference    string `json:"reference"`
	AmountCents  int64  `json:"amount_in_cents"`
	Currency     string `json:"currency"`
	PublicKey    string `json:"public_key"`
	ExpirationAt string `json:"expiration_time,omitempty"`
}

type WompiWebhookPayload struct {
	Event     string                 `json:"event"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"sent_at"`
	Signature struct {
		Properties []string `json:"properties"`
		Checksum   string   `json:"checksum"`
	} `json:"signature"`
}

// ─── Service ──────────────────────────────────────────────────────────────────

type WompiService struct {
	privateKey   string
	publicKey    string
	integrityKey string
	eventsKey    string
	baseURL      string
	environment  string
}

func NewWompiService() *WompiService {
	env := os.Getenv("WOMPI_ENV")
	if env == "" {
		env = "sandbox"
	}

	baseURL := "https://sandbox.wompi.co/v1"
	if env == "production" {
		baseURL = "https://production.wompi.co/v1"
	}

	svc := &WompiService{
		privateKey:   os.Getenv("WOMPI_PRIVATE_KEY"),
		publicKey:    os.Getenv("WOMPI_PUBLIC_KEY"),
		integrityKey: os.Getenv("WOMPI_INTEGRITY_KEY"),
		eventsKey:    os.Getenv("WOMPI_EVENTS_SECRET"), // ✅ FIX: era WOMPI_EVENTS_KEY
		baseURL:      baseURL,
		environment:  env,
	}

	if svc.privateKey == "" {
		fmt.Println("⚠️  WOMPI_PRIVATE_KEY no configurada")
	}
	if svc.integrityKey == "" {
		fmt.Println("⚠️  WOMPI_INTEGRITY_KEY no configurada")
	}
	if svc.eventsKey == "" {
		fmt.Println("⚠️  WOMPI_EVENTS_SECRET no configurada")
	}

	return svc
}

// GenerateIntegritySignature genera el hash SHA256 requerido por Wompi
// Fórmula: SHA256(reference + amount_in_cents + currency + integrity_key)
func (s *WompiService) GenerateIntegritySignature(req SignatureRequest) (*SignatureResponse, error) {
	if s.integrityKey == "" {
		return nil, fmt.Errorf("WOMPI_INTEGRITY_KEY no configurada")
	}

	toSign := fmt.Sprintf("%s%d%s%s",
		req.Reference,
		req.AmountCents,
		req.Currency,
		s.integrityKey,
	)

	if req.ExpirationAt != "" {
		toSign = fmt.Sprintf("%s%d%s%s%s",
			req.Reference,
			req.AmountCents,
			req.Currency,
			req.ExpirationAt,
			s.integrityKey,
		)
	}

	hash := sha256.Sum256([]byte(toSign))
	signature := fmt.Sprintf("%x", hash)

	return &SignatureResponse{
		Signature:    signature,
		Reference:    req.Reference,
		AmountCents:  req.AmountCents,
		Currency:     req.Currency,
		PublicKey:    s.publicKey,
		ExpirationAt: req.ExpirationAt,
	}, nil
}

// GetTransaction consulta el estado de una transacción en Wompi (server-to-server)
func (s *WompiService) GetTransaction(transactionID string) (map[string]interface{}, error) {
	// ✅ FIX: validar que la private key esté configurada antes de llamar
	if s.privateKey == "" {
		return nil, fmt.Errorf("WOMPI_PRIVATE_KEY no configurada")
	}

	url := fmt.Sprintf("%s/transactions/%s", s.baseURL, transactionID)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.privateKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error consultando Wompi: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// ✅ FIX: validar status code antes de parsear
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("wompi error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parseando respuesta de Wompi: %w", err)
	}

	return result, nil
}

// VerifyWebhookSignature verifica que el webhook realmente proviene de Wompi
func (s *WompiService) VerifyWebhookSignature(payload WompiWebhookPayload, headerSignature string) error {
	if s.eventsKey == "" {
		// En sandbox sin llave configurada, omitir verificación
		if s.environment == "sandbox" {
			fmt.Println("⚠️  WOMPI_EVENTS_SECRET no configurada — omitiendo verificación en sandbox")
			return nil
		}
		return fmt.Errorf("WOMPI_EVENTS_SECRET no configurada")
	}

	var parts []string
	for _, prop := range payload.Signature.Properties {
		parts = append(parts, getNestedValue(payload.Data, prop))
	}
	parts = append(parts, fmt.Sprintf("%d", payload.Timestamp))
	parts = append(parts, s.eventsKey)

	toSign := strings.Join(parts, "")
	hash := sha256.Sum256([]byte(toSign))
	computed := fmt.Sprintf("%x", hash)

	if computed != payload.Signature.Checksum {
		return fmt.Errorf("firma del webhook inválida")
	}

	return nil
}

// ProcessWebhookEvent procesa los eventos de Wompi
func (s *WompiService) ProcessWebhookEvent(payload WompiWebhookPayload) error {
	switch payload.Event {
	case "transaction.updated":
		transaction, ok := payload.Data["transaction"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("payload de transacción inválido")
		}

		status, _ := transaction["status"].(string)
		reference, _ := transaction["reference"].(string)
		transactionID, _ := transaction["id"].(string)

		fmt.Printf("📦 Webhook Wompi — Referencia: %s | Estado: %s | TxID: %s\n",
			reference, status, transactionID)

		// TODO: conectar con orderService para actualizar el pedido
		// switch status {
		// case "APPROVED":
		//     orderService.UpdatePaymentStatus(ctx, reference, "APPROVED")
		// case "DECLINED", "VOIDED":
		//     orderService.UpdatePaymentStatus(ctx, reference, "REJECTED")
		// }
	}

	return nil
}

// ─── Helper ───────────────────────────────────────────────────────────────────

func getNestedValue(data map[string]interface{}, path string) string {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			if val, ok := current[part]; ok {
				return fmt.Sprintf("%v", val)
			}
			return ""
		}
		if nested, ok := current[part].(map[string]interface{}); ok {
			current = nested
		} else {
			return ""
		}
	}
	return ""
}
