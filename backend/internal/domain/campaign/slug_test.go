package campaign

import "testing"

func TestSlugify(t *testing.T) {
	cases := []struct {
		title string
		want  string
	}{
		{"Ayuda para el viaje de estudios", "ayuda-para-el-viaje-de-estudios"},
		{"Rifa Solidaria 2026!!", "rifa-solidaria-2026"},
		{"Año Nuevo Niño Feliz", "ano-nuevo-nino-feliz"},
		{"   ", "campana"},
		{"---", "campana"},
	}
	for _, tc := range cases {
		if got := Slugify(tc.title); got != tc.want {
			t.Errorf("Slugify(%q) = %q, want %q", tc.title, got, tc.want)
		}
	}
}
