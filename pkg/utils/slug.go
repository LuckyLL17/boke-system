package utils

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	nonWordRegex = regexp.MustCompile(`[^\p{L}\p{N}\s-]`)
	spaceRegex   = regexp.MustCompile(`[\s-]+`)
)

func GenerateSlug(input string) string {
	if input == "" {
		return ""
	}

	slug := strings.TrimSpace(input)
	slug = nonWordRegex.ReplaceAllString(slug, "")
	slug = spaceRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	var result strings.Builder
	for _, r := range slug {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			result.WriteRune(unicode.ToLower(r))
		}
	}

	slug = result.String()
	if len(slug) > 100 {
		slug = slug[:100]
		slug = strings.TrimRight(slug, "-")
	}
	return slug
}

func UniqueSlug(base string, exists func(string) bool) string {
	slug := GenerateSlug(base)
	if !exists(slug) {
		return slug
	}
	for i := 1; i < 1000; i++ {
		candidate := slug + "-" + toString(i)
		if !exists(candidate) {
			return candidate
		}
	}
	return slug + "-" + toStringInt64(Now().Unix())
}

func toString(n int) string {
	return toStringInt64(int64(n))
}

func toStringInt64(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
