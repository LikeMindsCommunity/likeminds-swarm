package utils

import "strconv"

func ConvertNumberToString(number int) string {
	return strconv.Itoa(number)
}

// GetOrdinal | get the ordinal of a number (e.g. 1st, 2nd, 3rd, 4th, 5th, etc.)
func GetOrdinal(number int) string {
	lastDigit := number % 10
	if lastDigit == 1 && number != 11 {
		return strconv.Itoa(number) + "st"
	}
	if lastDigit == 2 && number != 12 {
		return strconv.Itoa(number) + "nd"
	}
	if lastDigit == 3 && number != 13 {
		return strconv.Itoa(number) + "rd"
	}
	return strconv.Itoa(number) + "th"
}
