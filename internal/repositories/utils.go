package repositories

import "go.mongodb.org/mongo-driver/bson/primitive"

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
