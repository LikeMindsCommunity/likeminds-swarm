package utils

import "strconv"

// Internal Method to fetch params and parse to int
func ParseIntFromQueryParam(paramString string, defaultValue int) (int, error) {
	param, err := strconv.Atoi(paramString)
	if err != nil {
		return defaultValue, err
	}

	return param, nil
}
