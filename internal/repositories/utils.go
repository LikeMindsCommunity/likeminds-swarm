package repositories

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	PostCollection           string = "post"
	LikeCollection           string = "like"
	CommentCollection        string = "comment"
	ActivityCollection       string = "activity"
	SaveCollection           string = "save"
	TopicCollection          string = "topic"
	WidgetCollection         string = "customWidget"
	PollVotesCollection      string = "pollVotes"
	ConnectionFeedCollection string = "connectionFeed"
)

// Internal Method to Insert a document in MongoDB
func _createDocumentInDB(db *mongo.Database, collectionName string, document interface{}) (interface{}, error) {
	coll := db.Collection(collectionName)
	result, err := coll.InsertOne(context.TODO(), document)

	return result.InsertedID, err
}

// Internal Method to Insert multiple documents in MongoDB
func _createManyDocumentsInDB(db *mongo.Database, collectionName string, documents []interface{}) ([]interface{}, error) {
	coll := db.Collection(collectionName)
	result, err := coll.InsertMany(context.TODO(), documents)

	return result.InsertedIDs, err
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

// Internal Method to Update all filter document in MongoDB
func _updateAllDocumentsInDB(db *mongo.Database, collectionName string, filter map[string]interface{}, update map[string]interface{}) error {
	coll := db.Collection(collectionName)
	_, err := coll.UpdateMany(context.TODO(), filter, update)

	return err
}

// Internal Method to Delete a document in MongoDB
func _deleteDocumentInDB(db *mongo.Database, collectionName string, filter map[string]interface{}) error {
	coll := db.Collection(collectionName)
	_, err := coll.DeleteOne(context.TODO(), filter)

	return err
}

// Internal Method to Delete many documents in MongoDB
func _deleteManyDocumentsInDB(db *mongo.Database, collectionName string, filter map[string]interface{}) (int64, error) {
	coll := db.Collection(collectionName)
	deleteResults, err := coll.DeleteMany(context.TODO(), filter)

	return deleteResults.DeletedCount, err
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

// Internal Method to Perform Aggregration on Collection
func _aggregateDocumentsInDB(db *mongo.Database, collectionName string, query []map[string]interface{}) ([]gin.H, error) {
	coll := db.Collection(collectionName)
	cursor, err := coll.Aggregate(context.TODO(), query)
	if err != nil {
		return nil, err
	}

	var results = []gin.H{}
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}
