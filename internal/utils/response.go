package utils

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// SuccessResponse devuelve una respuesta exitosa
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse devuelve una respuesta de error
func ErrorResponse(c *gin.Context, statusCode int, message string, err interface{}) {
	c.JSON(statusCode, Response{
		Success: false,
		Message: message,
		Error:   err,
	})
}

// ValidationErrorResponse devuelve errores de validación
func ValidationErrorResponse(c *gin.Context, errors interface{}) {
	c.JSON(400, Response{
		Success: false,
		Message: "Errores de validación",
		Error:   errors,
	})
}
