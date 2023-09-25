package utils

// RemoveAllOccurenceStringList | removes all occurences of item in string list
func RemoveAllOccurenceStringList(list [](string), item string) [](string) {
	newIter := 0

	// copy to list if list value != item
	for iter, value := range list {
		if value != item {
			list[newIter] = list[iter]
			newIter++
		}
	}

	// fill rest of the list with zero values for garbage collection
	var zero string
	for iterRem := newIter; iterRem < len(list); iterRem++ {
		list[iterRem] = zero
	}

	// slice the list till valid values and return
	return list[:newIter]
}
