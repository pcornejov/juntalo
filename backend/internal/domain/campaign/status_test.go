package campaign

import "testing"

func TestValidateTransition(t *testing.T) {
	cases := []struct {
		from, to Status
		wantErr  bool
	}{
		{StatusDraft, StatusActive, false},
		{StatusActive, StatusPaused, false},
		{StatusActive, StatusFinished, false},
		{StatusPaused, StatusActive, false},
		{StatusPaused, StatusFinished, false},
		{StatusDraft, StatusFinished, true},
		{StatusFinished, StatusActive, true},
		{StatusActive, StatusDraft, true},
		{StatusSuspended, StatusActive, true}, // solo plataforma puede salir de suspended
	}
	for _, tc := range cases {
		err := ValidateTransition(tc.from, tc.to)
		if (err != nil) != tc.wantErr {
			t.Errorf("ValidateTransition(%s, %s) error = %v, wantErr %v", tc.from, tc.to, err, tc.wantErr)
		}
	}
}
