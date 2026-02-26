package utils

import (
	"regexp"
	"strings"
)

var nonAlphanumRegex = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts a text to a URL-safe slug
func Slugify(text string) string {
	slug := strings.ToLower(text)
	slug = nonAlphanumRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}
