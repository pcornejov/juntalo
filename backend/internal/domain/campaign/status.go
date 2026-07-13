package campaign

import "github.com/pcornejov/juntalo/backend/internal/domain/apperr"

type Status string

const (
	StatusDraft     Status = "draft"
	StatusActive    Status = "active"
	StatusPaused    Status = "paused"
	StatusFinished  Status = "finished"
	StatusSuspended Status = "suspended"
)

// allowedTransitions encodes the state machine of Etapa 3 §3 (riesgo 13).
// suspended has no outgoing transition here: only platform/moderation can
// leave that state, never through the organizer-facing use cases.
var allowedTransitions = map[Status][]Status{
	StatusDraft:    {StatusActive},
	StatusActive:   {StatusPaused, StatusFinished},
	StatusPaused:   {StatusActive, StatusFinished},
	StatusFinished: {},
}

var errInvalidTransition = apperr.New("invalid_status_transition", "Esa transición de estado no está permitida")

// CanTransition reports whether moving from 'from' to 'to' is a legal transition.
func CanTransition(from, to Status) bool {
	for _, s := range allowedTransitions[from] {
		if s == to {
			return true
		}
	}
	return false
}

// ValidateTransition returns apperr "invalid_status_transition" when illegal
// (Etapa 4 §3: cada transición es un caso de uso explícito, no un PATCH genérico).
func ValidateTransition(from, to Status) error {
	if !CanTransition(from, to) {
		return errInvalidTransition
	}
	return nil
}
