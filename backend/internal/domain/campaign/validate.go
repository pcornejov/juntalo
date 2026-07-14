package campaign

import (
	"strings"
	"time"

	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
)

const maxTitleLength = 120

var (
	errUnknownType   = apperr.New("unknown_campaign_type", "Tipo de campaña desconocido")
	errTypeDisabled  = apperr.New("campaign_type_disabled", "Este tipo de campaña no está disponible todavía")
	errTitleRequired = apperr.New("validation_failed", "El título es obligatorio")
	errTitleTooLong  = apperr.New("validation_failed", "El título es demasiado largo")
)

// ValidateType checks that the given type key is known and enabled for creation.
func ValidateType(key TypeKey) error {
	def, ok := Registry[key]
	if !ok {
		return errUnknownType
	}
	if !def.Enabled {
		return errTypeDisabled
	}
	return nil
}

func ValidateTitle(title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return errTitleRequired
	}
	if len(title) > maxTitleLength {
		return errTitleTooLong
	}
	return nil
}

var errGoalBelowRaised = apperr.New("goal_below_raised", "La meta no puede ser menor a lo ya recaudado")

// ValidateGoalUpdate enforces that a campaign's goal never drops below what's
// already been raised (Etapa 4 §3).
func ValidateGoalUpdate(newGoal *money.CLP, raised money.CLP) error {
	if newGoal != nil && *newGoal < raised {
		return errGoalBelowRaised
	}
	return nil
}

var (
	errPublishAtPast     = apperr.New("validation_failed", "La fecha de publicación debe ser futura")
	errPublishAtNotDraft = apperr.New("campaign_not_active", "Solo se puede programar la publicación de una campaña en borrador")
)

// ValidatePublishAt enforces that a scheduled auto-publish date is only set
// on a draft campaign and lies in the future — un draft ya publicado o una
// fecha pasada no tiene sentido para el scheduler en background.
func ValidatePublishAt(publishAt *time.Time, status Status) error {
	if publishAt == nil {
		return nil
	}
	if status != StatusDraft {
		return errPublishAtNotDraft
	}
	if !publishAt.After(time.Now()) {
		return errPublishAtPast
	}
	return nil
}
