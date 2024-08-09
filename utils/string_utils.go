package utils

import (
	"strings"
	"unicode"
)

// ToCamelCase converts snake_case to CamelCase.
func ToCamelCase(snake string) string {
	parts := strings.Split(snake, "_")
	for i, part := range parts {
		parts[i] = strings.Title(part)
	}
	return Capitalize(strings.Join(parts, ""))
}

// Capitalize capitalizes the first character of a string.
func Capitalize(str string) string {
	if len(str) == 0 {
		return str
	}
	runes := []rune(str)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// capitalizeFirstLetter capitalizes the first letter of a string.
func CapitalizeFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
