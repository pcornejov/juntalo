// Package money models CLP amounts as integer cents-free pesos, per Etapa 3 (§1: "Dinero").
package money

import "fmt"

// CLP is an amount in Chilean pesos: always an integer, no decimals.
type CLP int64

func (c CLP) Add(other CLP) CLP {
	return c + other
}

func (c CLP) Sub(other CLP) CLP {
	return c - other
}

// IsPositive reports whether the amount is strictly greater than zero.
// Money amounts in Juntalo (contributions, payments) must never be zero or negative.
func (c CLP) IsPositive() bool {
	return c > 0
}

func (c CLP) String() string {
	return fmt.Sprintf("$%d", int64(c))
}
