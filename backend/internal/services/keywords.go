package services

import "strings"

func ParseCommaSeparatedKeywords(value string) []string {
	return NormalizeKeywords(strings.Split(value, ","))
}

func NormalizeKeywords(values []string) []string {
	keywords := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		keywords = append(keywords, trimmed)
	}
	return keywords
}
