package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// bcryptCost: rondas = 2^cost. Cost 10 es el default de golang.org/x/crypto/bcrypt
// y el mínimo recomendado por OWASP. Cost 12 era ~4x más caro y resultaba
// inviable en CPU compartida (Render Free 0.1 vCPU → ~2-3s por compare).
const bcryptCost = 10

// HashPassword hashea una contraseña usando bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compara una contraseña con su hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
