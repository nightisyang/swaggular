package Dtos

import (
	"regexp"
	"strings"
)

// ToCamelCase converts a string to camelCase.
func ToCamelCase(input string) string {
	input = strings.ReplaceAll(input, "_", " ")
	words := regexp.MustCompile(`[A-Za-z][^A-Z\s]*`).FindAllString(input, -1)

	for i := range words {
		if i == 0 {
			words[i] = strings.ToLower(words[i])
		} else {
			words[i] = strings.Title(words[i])
		}
	}

	return strings.Join(words, "")
}

// OptionalSuffix adds a question mark for optional query parameters.
func OptionalSuffix(isRequired bool) string {
	if isRequired {
		return ""
	}
	return "?"
}
