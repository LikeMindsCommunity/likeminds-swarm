package helpers

import (
	"net/url"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Internal Method to Convert Hex Id to Object Id
func convertHexIdToObjectId(filter map[string]interface{}, key string) error {
	var err error

	if idValue, ok := filter[key]; ok {
		if _, ok := idValue.(string); ok {
			filter[key], err = primitive.ObjectIDFromHex(idValue.(string))

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Internal Method to convert Multiple Hex Ids to Object Ids
func convertHexIdsToObjectIds(filter map[string]interface{}, keys []string) error {
	for _, key := range keys {
		err := convertHexIdToObjectId(filter, key)
		if err != nil {
			return err
		}
	}

	return nil
}

// Internal Method to create filters according to available values
func mergeFilterOptions(filterOptions map[string]interface{}) options.FindOptions {
	fOpts := options.FindOptions{}

	if value, ok := filterOptions["$sort"]; ok {
		fOpts.SetSort(value)
	}

	if value, ok := filterOptions["$skip"]; ok {
		fOpts.SetSkip(int64(value.(int)))
	}

	if value, ok := filterOptions["$limit"]; ok {
		fOpts.SetLimit(int64(value.(int)))
	}

	return fOpts
}

// Helper Method to convert Multiple Ids to Object Ids without throwing error
func ConvertIdsToObjectIds(listIds []string) []primitive.ObjectID {

	hexIds := make([]primitive.ObjectID, len(listIds))

	for i, id := range listIds {
		hexIds[i], _ = primitive.ObjectIDFromHex(id)
	}

	return hexIds
}

// Helper method to check if string is a valid URL or not
func IsValidURL(urlString string) bool {
	u, err := url.Parse(urlString)

	if err != nil || u.Host == "" || u.Scheme == "" {
		return false
	}

	return true
}

// Helper method to check if strings are valid URLs or not
func AreValidURLs(urlStrings []string) string {
	for _, urlString := range urlStrings {
		if urlString != "" {
			isValid := IsValidURL(urlString)

			if !isValid {
				return "Invalid url in attachments meta_data: " + urlString
			}
		}
	}

	return ""
}
