package utils

import (
	"strconv"
	"strings"
)

// Internal Method to parse an Integer Array from query params
func ParseIntArrayParam(param string) []int {
	response := []int{}

	intermediate_strings := ParseStringArrayParam(param)

	for _, value := range intermediate_strings {
		convertedValue, err := strconv.Atoi(value)
		if err == nil {
			response = append(response, convertedValue)
		}
	}

	return response
}

// Internal Method to parse a String Array from query params
func ParseStringArrayParam(param string) []string {
	response := []string{}

	// Removal of square braces from array string
	if len(param) > 0 && param[0] == '[' {
		param = param[1:]
	}

	if len(param) > 0 && param[len(param)-1] == ']' {
		param = param[:len(param)-1]
	}

	// Removal of extra spaces from the array string
	param = strings.TrimSpace(param)

	if len(param) > 0 {
		paramValues := strings.Split(param, ",")

		for _, value := range paramValues {
			value = strings.TrimSpace(value)

			// Removal of quotes from each string from array
			if len(value) > 0 && value[0] == '"' {
				value = value[1:]
			}

			if len(value) > 0 && value[len(value)-1] == '"' {
				value = value[:len(value)-1]
			}

			response = append(response, value)
		}
	}

	return response
}

// Internal Method to fetch params and parse to int
func ParseIntFromQueryParam(paramString string, defaultValue int) (int, error) {
	if paramString == "" {
		return defaultValue, nil
	}
	param, err := strconv.Atoi(paramString)
	if err != nil {
		return defaultValue, err
	}

	return param, nil
}
