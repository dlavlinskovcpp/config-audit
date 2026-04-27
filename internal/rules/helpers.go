package rules

import (
	"strings"
	"unicode"
)

func normalizeIdentifier(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))

	for _, r := range strings.ToLower(value) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

func pathHasSuffix(path string, suffix string) bool {
	path = strings.ToLower(path)
	suffix = strings.ToLower(suffix)

	return path == suffix || strings.HasSuffix(path, "."+suffix)
}
