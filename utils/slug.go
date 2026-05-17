package utils

import (
	"regexp"
	"strings"
)

func GenerateSlug(title string) string {
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug := strings.ToLower(title)
	slug = re.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}
