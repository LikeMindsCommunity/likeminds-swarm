package utils

import "strings"

func CapitalizeFirstLetter(str string) string {
	return strings.ToUpper(string(str[0])) + strings.ToLower(string(str[1:]))
}
