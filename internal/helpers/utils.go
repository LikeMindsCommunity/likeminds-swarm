package helpers

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

func convertMultipleHexIdsToObjectIds(filter map[string]interface{}, keys []string) error {
	for _, key := range keys {
		err := convertHexIdToObjectId(filter, key)
		if err != nil {
			return err
		}
	}

	return nil
}

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
