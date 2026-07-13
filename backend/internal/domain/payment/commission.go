package payment

import (
	"math"

	"github.com/pcornejov/juntalo/backend/internal/domain/money"
)

// ComputeCommission applies the organization's commission rate to a gross
// amount, rounding to the nearest peso (CLP has no decimals). The caller is
// responsible for persisting both the rate and the resulting amounts as a
// snapshot (Etapa 3 §5, riesgo 3) — never recomputed from a live rate later.
func ComputeCommission(gross money.CLP, rate float64) (commission, net money.CLP) {
	c := money.CLP(math.Round(float64(gross) * rate))
	return c, gross - c
}
