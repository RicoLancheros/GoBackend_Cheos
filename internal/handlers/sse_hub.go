package handlers

import "sync"

// SSEHub mantiene un canal por usuario conectado vía SSE.
type SSEHub struct {
	mu      sync.RWMutex
	clients map[string]chan []byte
}

func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[string]chan []byte),
	}
}

func (h *SSEHub) Subscribe(userID string) chan []byte {
	ch := make(chan []byte, 16)
	h.mu.Lock()
	// Si ya había una conexión previa (tab duplicada), la cerramos limpiamente
	if old, ok := h.clients[userID]; ok {
		close(old)
	}
	h.clients[userID] = ch
	h.mu.Unlock()
	return ch
}

func (h *SSEHub) Unsubscribe(userID string, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// Solo borramos si el canal sigue siendo el actual
	if current, ok := h.clients[userID]; ok && current == ch {
		close(ch)
		delete(h.clients, userID)
	}
}

// Publish envía payload al usuario si está conectado. No-op silencioso si no lo está.
func (h *SSEHub) Publish(userID string, payload []byte) {
	h.mu.RLock()
	ch, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return
	}
	select {
	case ch <- payload:
	default:
		// Canal lleno (cliente lento) — descartamos para no bloquear
	}
}
