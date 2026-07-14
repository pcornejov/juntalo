// Package apperr defines the stable error codes that cross the domain/app/api
// boundary intact (Etapa 4 §1: el frontend traduce códigos, nunca parsea mensajes).
package apperr

type Error struct {
	Code    string
	Message string
	// Details: solo lo usa la validación de forma en el borde de la API
	// (dto.Validate) para indicar qué campo falló y por qué — las reglas de
	// negocio del dominio siguen comunicándose solo por Code (Etapa 4 §1).
	Details map[string]any
}

func (e *Error) Error() string {
	return e.Message
}

func New(code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// WithDetails attaches field-level details to an existing error (fluent,
// para no tener que agregar un parámetro a cada llamada de New).
func (e *Error) WithDetails(details map[string]any) *Error {
	e.Details = details
	return e
}

// Is reports whether err is an *Error with the given code.
func Is(err error, code string) bool {
	appErr, ok := err.(*Error)
	return ok && appErr.Code == code
}
