package payment

import "testing"

func TestCanTransition(t *testing.T) {
	cases := []struct {
		from, to Status
		want     bool
	}{
		{StatusPending, StatusConfirmed, true},
		{StatusPending, StatusFailed, true},
		{StatusConfirmed, StatusRefunded, true},
		{StatusConfirmed, StatusPartiallyRefunded, true},
		{StatusPartiallyRefunded, StatusRefunded, true},
		{StatusPending, StatusRefunded, false},
		{StatusFailed, StatusConfirmed, false},
		{StatusRefunded, StatusConfirmed, false},
		{StatusConfirmed, StatusPending, false},
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
	if !IsTerminal(StatusRefunded) {
		t.Error("refunded should be terminal")
	}
	if IsTerminal(StatusPending) {
		t.Error("pending should not be terminal")
	}
	if IsTerminal(StatusConfirmed) {
		t.Error("confirmed should not be terminal (can be refunded)")
	}
}
