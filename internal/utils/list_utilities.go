package utils

// RemoveAllOccurenceStringList | removes all occurences of item in string list
func RemoveAllOccurenceStringList(list [](string), item string) [](string) {
	j := 0
	for i, v := range list {
		if v != item {
			list[j] = list[i]
			j++
		}
	}
	var zero string
	for k := j; k < len(list); k++ {
		list[k] = zero
	}

	return list[:j]
}
