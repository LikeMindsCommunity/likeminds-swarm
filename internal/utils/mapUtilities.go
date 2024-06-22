package utils

import "sort"

// Utility method to sort map by values and return keys | orderDesc: sort in descending order
func SortMapByValues(m map[string]float64, orderDesc bool) []string {

	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	if orderDesc {
		sort.SliceStable(keys, func(i, j int) bool {
			return m[keys[i]] > m[keys[j]]
		})
	} else {
		sort.SliceStable(keys, func(i, j int) bool {
			return m[keys[i]] < m[keys[j]]
		})
	}

	return keys
}
