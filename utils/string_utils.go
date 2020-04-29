package utils

import (
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

// NormalizeText return normalize text from raw text
func NormalizeText(rawText string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	normStr1, _, _ := transform.String(t, rawText)
	return normStr1
}
