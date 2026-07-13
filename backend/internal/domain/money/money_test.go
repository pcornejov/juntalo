package money

import "testing"

func TestCLP_IsPositive(t *testing.T) {
	cases := []struct {
		amount CLP
		want   bool
	}{
		{0, false},
		{-1, false},
		{1, true},
		{1_000_000, true},
	}
	for _, tc := range cases {
		if got := tc.amount.IsPositive(); got != tc.want {
			t.Errorf("CLP(%d).IsPositive() = %v, want %v", tc.amount, got, tc.want)
		}
	}
}

func TestCLP_AddSub(t *testing.T) {
	a := CLP(1000)
	b := CLP(300)
	if got := a.Add(b); got != CLP(1300) {
		t.Errorf("Add: got %d, want 1300", got)
	}
	if got := a.Sub(b); got != CLP(700) {
		t.Errorf("Sub: got %d, want 700", got)
	}
}
