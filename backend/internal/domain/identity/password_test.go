package identity

import "testing"

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		name    string
		pw      string
		wantErr bool
	}{
		{"too short", "abc123", true},
		{"exactly minimum", "12345678", false},
		{"long", "a-very-long-password-123", false},
		{"empty", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePassword(tc.pw)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidatePassword(%q) error = %v, wantErr %v", tc.pw, err, tc.wantErr)
			}
		})
	}
}
