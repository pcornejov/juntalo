package payment

type Status string

const (
	StatusPending           Status = "pending"
	StatusConfirmed         Status = "confirmed"
	StatusFailed            Status = "failed"
	StatusRefunded          Status = "refunded"
	StatusPartiallyRefunded Status = "partially_refunded"
)

// allowedTransitions is the payment state machine (Etapa 3 §5):
// pending -> confirmed -> {partially_refunded -> refunded, refunded}
// pending -> failed
// failed y refunded son terminales.
var allowedTransitions = map[Status][]Status{
	StatusPending:           {StatusConfirmed, StatusFailed},
	StatusConfirmed:         {StatusPartiallyRefunded, StatusRefunded},
	StatusPartiallyRefunded: {StatusRefunded},
	StatusFailed:            {},
	StatusRefunded:          {},
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
