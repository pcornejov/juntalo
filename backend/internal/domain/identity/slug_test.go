package identity

import "testing"

func TestSlugifyOrgName(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"Organización de María José", "organizacion-de-maria-jose"},
		{"Junta de Vecinos Nº 12!!", "junta-de-vecinos-n-12"},
		{"   ", "organizacion"},
		{"---", "organizacion"},
	}
	for _, tc := range cases {
		if got := SlugifyOrgName(tc.name); got != tc.want {
			t.Errorf("SlugifyOrgName(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}
}
