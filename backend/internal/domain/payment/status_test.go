package payment

import "testing"

func TestCanTransition(t *testing.T) {
	cases := []struct {
		from, to Status
		want     bool
	}{
		{StatusPending, StatusConfirmed, true},
		{StatusPending, StatusFailed, true},
		{StatusFailed, StatusConfirmed, false},
		{StatusConfirmed, StatusPending, false},
		{StatusConfirmed, StatusFailed, false},
	}
	for _, tc := range cases {
		if got := CanTransition(tc.from, tc.to); got != tc.want {
			t.Errorf("CanTransition(%s, %s) = %v, want %v", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestIsTerminal(t *testing.T) {
	if !IsTerminal(StatusFailed) {
		t.Error("failed should be terminal")
	}
	if !IsTerminal(StatusConfirmed) {
		t.Error("confirmed should be terminal")
	}
	if IsTerminal(StatusPending) {
		t.Error("pending should not be terminal")
	}
}
