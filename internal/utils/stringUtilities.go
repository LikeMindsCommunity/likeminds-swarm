package utils

import (
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

// RemoveDuplicatesFromSlice returns a slice of unique strings from the input slice
func RemoveDuplicatesFromStringSlice(slice []string) []string {
	seen := make(map[string]struct{})
	var result []string

	for _, value := range slice {
		if _, ok := seen[value]; !ok {
			result = append(result, value)
			seen[value] = struct{}{}
		}
	}
	return result
}

// RemoveDuplicatesFromObjectIDSlice returns a slice of unique ObjectIDs from the input slice
func RemoveDuplicatesFromObjectIDSlice(slice []primitive.ObjectID) []primitive.ObjectID {
	seen := make(map[primitive.ObjectID]struct{})
	var result []primitive.ObjectID

	for _, value := range slice {
		if _, ok := seen[value]; !ok {
			result = append(result, value)
			seen[value] = struct{}{}
		}
	}
	return result
}
