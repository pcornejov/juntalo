// Package apperr defines the stable error codes that cross the domain/app/api
// boundary intact (Etapa 4 §1: el frontend traduce códigos, nunca parsea mensajes).
package apperr

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func New(code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Is reports whether err is an *Error with the given code.
func Is(err error, code string) bool {
	appErr, ok := err.(*Error)
	return ok && appErr.Code == code
}
