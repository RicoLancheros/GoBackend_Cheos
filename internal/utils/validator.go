package utils

import (
	"fmt"

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

// FormatValidationErrors formatea los errores de validación con mensajes específicos por campo
func FormatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			// Intentar mensaje específico por campo, si no existe usar genérico
			msg := getFieldMessage(field, e.Tag(), e.Param())
			errors[field] = msg
		}
	}

	return errors
}

// getFieldMessage devuelve un mensaje de error específico según el campo y el tipo de validación
func getFieldMessage(field string, tag string, param string) string {
	// Mensajes específicos por campo
	switch field {
	case "Email":
		switch tag {
		case "required":
			return "El correo electrónico es obligatorio"
		case "email":
			return "El correo electrónico no es válido. Debe tener formato ejemplo@dominio.com"
		}
	case "Password":
		switch tag {
		case "required":
			return "La contraseña es obligatoria"
		case "min":
			return fmt.Sprintf("La contraseña debe tener al menos %s caracteres", param)
		}
	case "Name":
		switch tag {
		case "required":
			return "El nombre es obligatorio"
		case "min":
			return fmt.Sprintf("El nombre debe tener al menos %s caracteres", param)
		case "excludesall":
			return "El nombre no debe contener números"
		}
	case "Phone":
		switch tag {
		case "required":
			return "El número de teléfono es obligatorio"
		}
	case "Gender":
		switch tag {
		case "oneof":
			return "El género debe ser uno de: MALE, FEMALE, OTHER"
		}
	}

	// Mensajes genéricos por tag (fallback para otros modelos)
	switch tag {
	case "required":
		return "Este campo es obligatorio"
	case "email":
		return "Debe ser un correo electrónico válido"
	case "min":
		return fmt.Sprintf("Debe tener al menos %s caracteres", param)
	case "max":
		return fmt.Sprintf("No debe exceder %s caracteres", param)
	case "gt":
		return fmt.Sprintf("Debe ser mayor que %s", param)
	case "gte":
		return fmt.Sprintf("Debe ser mayor o igual a %s", param)
	case "lt":
		return fmt.Sprintf("Debe ser menor que %s", param)
	case "lte":
		return fmt.Sprintf("Debe ser menor o igual a %s", param)
	case "url":
		return "Debe ser una URL válida"
	case "oneof":
		return fmt.Sprintf("Debe ser uno de: %s", param)
	case "excludesall":
		return "Contiene caracteres no permitidos"
	default:
		return "Valor inválido"
	}
}
