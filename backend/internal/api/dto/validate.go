package dto

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
)

var validate = validator.New()

func init() {
	// Usa el nombre del campo JSON (ej. "goal_amount") en vez del nombre Go
	// (ej. "GoalAmount") en los detalles de validación — así el frontend
	// puede mapear cada detalle directo a su input sin traducir nombres.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// Validate applies struct tags at the API edge (Etapa 2 §2.5). Business invariants
// are re-validated in domain — this is only the first line of defense. Cuando falla,
// Details trae un mensaje legible por campo (clave = nombre JSON) para que el
// frontend pueda marcar el input específico en vez de mostrar un error genérico
// (auditoría de QA: antes siempre era el mismo mensaje sin decir qué campo falló).
func Validate(v any) error {
	err := validate.Struct(v)
	if err == nil {
		return nil
	}

	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok || len(validationErrs) == 0 {
		return apperr.New("validation_failed", "Los datos enviados no son válidos")
	}

	details := make(map[string]any, len(validationErrs))
	for _, fe := range validationErrs {
		details[fe.Field()] = fieldErrorMessage(fe)
	}
	return apperr.New("validation_failed", "Los datos enviados no son válidos").WithDetails(details)
}

func fieldErrorMessage(fe validator.FieldError) string {
	isString := fe.Kind() == reflect.String
	switch fe.Tag() {
	case "required":
		return "Este campo es obligatorio"
	case "email":
		return "Debe ser un email válido"
	case "min":
		if isString {
			return "Debe tener al menos " + fe.Param() + " caracteres"
		}
		return "Debe ser mayor o igual a " + fe.Param()
	case "max":
		if isString {
			return "Debe tener como máximo " + fe.Param() + " caracteres"
		}
		return "Debe ser menor o igual a " + fe.Param()
	case "gt":
		return "Debe ser mayor a " + fe.Param()
	case "lt":
		return "Debe ser menor a " + fe.Param()
	default:
		return "Valor inválido"
	}
}
