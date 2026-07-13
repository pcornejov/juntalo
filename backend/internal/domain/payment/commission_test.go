package payment

import (
	"testing"

	"github.com/pcornejov/juntalo/backend/internal/domain/money"
)

func TestComputeCommission(t *testing.T) {
	cases := []struct {
		gross          money.CLP
		rate           float64
		wantCommission money.CLP
		wantNet        money.CLP
	}{
		{100_000, 0.05, 5_000, 95_000},
		{1_000, 0.05, 50, 950},
		{1, 0.05, 0, 1},   // redondea a 0.05 -> 0
		{33, 0.10, 3, 30}, // 3.3 redondea a 3
		{35, 0.10, 4, 31}, // 3.5 redondea a 4 (round-half-away-from-zero de math.Round)
		{0, 0.05, 0, 0},
	}
	for _, tc := range cases {
		commission, net := ComputeCommission(tc.gross, tc.rate)
		if commission != tc.wantCommission {
			t.Errorf("ComputeCommission(%d, %.2f) commission = %d, want %d", tc.gross, tc.rate, commission, tc.wantCommission)
		}
		if net != tc.wantNet {
			t.Errorf("ComputeCommission(%d, %.2f) net = %d, want %d", tc.gross, tc.rate, net, tc.wantNet)
		}
		if commission+net != tc.gross {
			t.Errorf("commission + net must equal gross: %d + %d != %d", commission, net, tc.gross)
		}
	}
}
