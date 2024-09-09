package utils

import (
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

// This function is used to parse String array to json string using json marshal
func ParseStringArrayToString(array []string) string {
	temp_params, _ := json.Marshal(array)

	str := fmt.Sprintf("%v", string(temp_params))

	return str
}

// returns the elements in `a` that aren't in `b`.
func GetDifferenceBetweenArray(a []primitive.ObjectID, b []primitive.ObjectID) []primitive.ObjectID {
	mb := make(map[primitive.ObjectID]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []primitive.ObjectID
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

// returns the elements in `a` that aren't in `b`.
func GetDifferenceBetweenStringArray(a []string, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}

	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

// returns the elements in `a` that are in `b`.
func GetSimilarBetweenArray(a []string, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var similar []string
	for _, x := range a {
		if _, found := mb[x]; found {
			similar = append(similar, x)
		}
	}
	return similar
}

// This function returns the minimum number in the array
func GetMinimumFromArray(vars ...float64) float64 {
	min := vars[0]

	for _, i := range vars {
		if min > i {
			min = i
		}
	}

	return min
}
