package campaign

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)
	trimDashes      = regexp.MustCompile(`^-+|-+$`)
)

// Slugify normalizes a title into a URL-safe slug base. It does not guarantee
// uniqueness — callers must append a suffix on collision (Etapa 4 §3: el slug
// lo genera el servidor).
func Slugify(title string) string {
	s := strings.ToLower(title)
	s = stripDiacritics(s)
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	s = trimDashes.ReplaceAllString(s, "")
	if s == "" {
		s = "campana"
	}
	if len(s) > 80 {
		s = s[:80]
	}
	return s
}

// WithSuffix appends a short disambiguating suffix, e.g. slug-4f2a.
func WithSuffix(base, suffix string) string {
	return fmt.Sprintf("%s-%s", base, suffix)
}

var diacriticsReplacer = strings.NewReplacer(
	"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ñ", "n", "ü", "u",
)

func stripDiacritics(s string) string {
	return diacriticsReplacer.Replace(s)
}
