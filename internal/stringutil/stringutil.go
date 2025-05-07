package stringutil

import (
	"bytes"
	"strings"
	"unicode"
)

func PascalToSnake(s string) string {
	var b bytes.Buffer

	for i, c := range s {
		if unicode.IsUpper(c) {
			if i > 0 && (unicode.IsLower(rune(s[i-1])) || (i+1 < len(s) && unicode.IsLower(rune(s[i+1])))) {
				b.WriteByte('_')
			}

			b.WriteRune(unicode.ToLower(c))
		} else {
			b.WriteRune(c)
		}
	}

	return b.String()
}

func PascalToTitle(s string) string {
	var b bytes.Buffer

	var last rune
	for i, r := range s {
		if i == 0 || last == ' ' || last == '_' || unicode.IsDigit(last) {
			b.WriteRune(unicode.ToUpper(r))
		} else if unicode.IsUpper(r) && (i+1 == len(s) || unicode.IsLower(rune(s[i+1]))) {
			b.WriteRune(r)
		} else if i+1 < len(s) && (unicode.IsLower(r) || unicode.IsDigit(r)) && (unicode.IsUpper(rune(s[i+1])) || s[i+1] == '_') {
			b.WriteRune(r)
			b.WriteRune(' ')
		} else {
			b.WriteRune(unicode.ToLower(r))
		}
		last = r
	}

	return b.String()
}

func LooksTrue(s string) bool {
	switch strings.ToLower(s) {
	case "true", "yes", "1", "on", "enabled", "enable", "active", "ok", "okay":
		return true
	default:
		return false
	}
}
