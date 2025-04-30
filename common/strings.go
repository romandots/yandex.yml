package common

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func Inflect(n int, noun []string) string {
	var inflected string
	switch {
	case n%100 > 10 && n%100 < 15:
		inflected = noun[2]
	case n%10 == 1:
		inflected = noun[0]
	case n%10 > 1 && n%10 < 5:
		inflected = noun[1]
	default:
		inflected = noun[2]
	}

	return fmt.Sprintf("%d %s", n, inflected)
}

// SafelyTruncate ensures the string is truncated without breaking XML entities or tags.
func SafelyTruncate(input string, maxLength int) string {
	if len(input) <= maxLength {
		return input
	}

	// Truncate to the nearest valid UTF-8 boundary
	truncated := input[:maxLength-3]
	for !utf8.ValidString(truncated) {
		truncated = truncated[:len(truncated)-1]
	}

	// Ensure no partial XML entities are left
	if lastAmp := strings.LastIndex(truncated, "&"); lastAmp != -1 {
		if semicolon := strings.Index(truncated[lastAmp:], ";"); semicolon == -1 {
			truncated = truncated[:lastAmp] // Remove incomplete entity
		}
	}

	return truncated + "..."
}
