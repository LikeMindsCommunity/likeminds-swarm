package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	TopicCollection string = "topic"
)

// Internal Method to Insert a document in MongoDB
func _createDocumentInDB(db *mongo.Database, collectionName string, document interface{}) (interface{}, error) {
	coll := db.Collection(collectionName)
	result, err := coll.InsertOne(context.TODO(), document)

	return result.InsertedID, err
}

// Internal Method to Find a document in MongoDB
func _findDocumentsInDB(db *mongo.Database, collectionName string, filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error) {
	coll := db.Collection(collectionName)
	cursor, err := coll.Find(context.TODO(), filter, filterOpts)

	return cursor, err
}

// Internal Method to Update a document in MongoDB
func _updateDocumentsInDB(db *mongo.Database, collectionName string, filter map[string]interface{}, update map[string]interface{}) error {
	coll := db.Collection(collectionName)
	_, err := coll.UpdateOne(context.TODO(), filter, update)

	return err
}

// Internal Method to Fetch Documents Count
func _countDocumentsInDB(db *mongo.Database, collectionName string, filter map[string]interface{}) (int64, error) {
	coll := db.Collection(collectionName)
	count, err := coll.CountDocuments(context.TODO(), filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}
