package identity

import (
	"regexp"
	"strings"
)

var (
	nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)
	trimDashes      = regexp.MustCompile(`^-+|-+$`)
)

var diacriticsReplacer = strings.NewReplacer(
	"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ñ", "n", "ü", "u",
)

// SlugifyOrgName normalizes an organization name into a URL-safe slug base
// for su página pública (/org/:slug). No garantiza unicidad — el caller debe
// agregar un sufijo en caso de colisión, igual que campaign.Slugify.
func SlugifyOrgName(name string) string {
	s := strings.ToLower(name)
	s = diacriticsReplacer.Replace(s)
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	s = trimDashes.ReplaceAllString(s, "")
	if s == "" {
		s = "organizacion"
	}
	if len(s) > 80 {
		s = s[:80]
	}
	return s
}
