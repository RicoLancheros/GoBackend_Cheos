package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateOrderNumber genera un número de orden único
func GenerateOrderNumber() string {
	// Formato: CHEOS-YYYYMMDD-XXXXXX
	// Donde XXXXXX es un número aleatorio de 6 dígitos
	now := time.Now()
	dateStr := now.Format("20060102")
	rand.Seed(now.UnixNano())
	randomNum := rand.Intn(999999)
	return fmt.Sprintf("CHEOS-%s-%06d", dateStr, randomNum)
}
