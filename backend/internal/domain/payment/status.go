package payment

type Status string

const (
	StatusPending   Status = "pending"
	StatusConfirmed Status = "confirmed"
	StatusFailed    Status = "failed"
)

// allowedTransitions is the payment state machine: pending -> {confirmed,
// failed}. confirmed y failed son terminales — Juntalo es solo para aportar
// a causas, no hay reembolso self-service que reabra un pago confirmado.
var allowedTransitions = map[Status][]Status{
	StatusPending:   {StatusConfirmed, StatusFailed},
	StatusConfirmed: {},
	StatusFailed:    {},
}

// CanTransition reports whether moving from 'from' to 'to' is legal.
func CanTransition(from, to Status) bool {
	for _, s := range allowedTransitions[from] {
		if s == to {
			return true
		}
	}
	return false
}

// IsTerminal reports whether a payment in this status can no longer transition.
func IsTerminal(s Status) bool {
	return len(allowedTransitions[s]) == 0
}
