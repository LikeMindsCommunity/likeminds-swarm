package utils

import "strings"

func CapitalizeFirstLetter(str string) string {
	return strings.ToUpper(string(str[0])) + strings.ToLower(string(str[1:]))
}

// GetDuplicatesFromSlice returns a slice of duplicate strings from the input slice
func GetDuplicatesFromSlice(slice []string) []string {
	seen := make(map[string]struct{})
	var result []string

	for _, value := range slice {
		if _, ok := seen[value]; ok {
			result = append(result, value)
		} else {
			seen[value] = struct{}{}
		}
	}
	return result
}
