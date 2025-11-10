package utils

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateStruct valida una estructura usando tags de validación
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// GetValidator retorna la instancia del validador
func GetValidator() *validator.Validate {
	return validate
}

// FormatValidationErrors formatea los errores de validación
func FormatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			switch e.Tag() {
			case "required":
				errors[field] = "Este campo es requerido"
			case "email":
				errors[field] = "Debe ser un email válido"
			case "min":
				errors[field] = "Debe tener al menos " + e.Param() + " caracteres"
			case "max":
				errors[field] = "No debe exceder " + e.Param() + " caracteres"
			case "gt":
				errors[field] = "Debe ser mayor que " + e.Param()
			case "gte":
				errors[field] = "Debe ser mayor o igual a " + e.Param()
			case "lt":
				errors[field] = "Debe ser menor que " + e.Param()
			case "lte":
				errors[field] = "Debe ser menor o igual a " + e.Param()
			case "url":
				errors[field] = "Debe ser una URL válida"
			default:
				errors[field] = "Valor inválido"
			}
		}
	}

	return errors
}
