package contribution

import (
	"testing"

	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
)

func TestValidateAmount(t *testing.T) {
	cases := []struct {
		amount  money.CLP
		wantErr bool
	}{
		{0, true},
		{-100, true},
		{1, false},
		{500_000, false},
		{50_000_000, false},
		{50_000_001, true},
	}
	for _, tc := range cases {
		err := ValidateAmount(tc.amount)
		if (err != nil) != tc.wantErr {
			t.Errorf("ValidateAmount(%d) error = %v, wantErr %v", tc.amount, err, tc.wantErr)
		}
		if err != nil && !apperr.Is(err, "amount_out_of_range") {
			t.Errorf("expected amount_out_of_range, got %v", err)
		}
	}
}
