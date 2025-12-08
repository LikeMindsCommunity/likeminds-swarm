package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/services/environment"
	"github.com/nateshr/likeminds-swarm/internal/services/searchElastic"
)

// syncCollection fetches all docs from MongoDB and indexes them into Elasticsearch
func syncCollection(ctx context.Context, db *mongo.Database, esHelper searchElastic.EsHelper, collectionName, indexName string) {
	log.Printf("Starting sync for collection: %s to index: %s", collectionName, indexName)

	coll := db.Collection(collectionName)

	// First, get the total count of documents
	totalCount, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Printf("Failed to count documents in %s: %v\n", collectionName, err)
		return
	}
	log.Printf("Total documents in %s: %d", collectionName, totalCount)

	if totalCount == 0 {
		log.Printf("No documents found in %s, skipping sync", collectionName)
		return
	}

	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Failed to read from %s: %v\n", collectionName, err)
		return
	}
	defer cursor.Close(ctx)

	docs := make(map[string]interface{})
	count := 0
	batchCount := 0

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			log.Printf("Decode error in %s: %v\n", collectionName, err)
			continue
		}

		// Normalize Mongo _id to string for ES document ID and source
		var id string
		switch v := doc["_id"].(type) {
		case primitive.ObjectID:
			id = v.Hex()
		default:
			id = fmt.Sprintf("%v", v)
		}
		// Optionally store _id as string in the document source to avoid nested $oid
		doc["_id"] = id
		docs[id] = doc
		count++

		delete(doc, "_id")

		// Batch processing (for performance)
		if len(docs) >= 100 {
			batchCount++
			log.Printf("Processing batch %d with %d documents for %s", batchCount, len(docs), collectionName)

			if err := esHelper.InsertManyDocuments(docs, indexName); err != nil {
				log.Printf("Bulk insert failed for %s batch %d: %v\n", indexName, batchCount, err)
				// Continue with next batch even if this one fails
			} else {
				log.Printf("Successfully inserted batch %d with %d documents into %s", batchCount, len(docs), indexName)
			}
			docs = make(map[string]interface{})
		}
	}

	// Process remaining documents
	if len(docs) > 0 {
		batchCount++
		log.Printf("Processing final batch %d with %d documents for %s", batchCount, len(docs), collectionName)

		if err := esHelper.InsertManyDocuments(docs, indexName); err != nil {
			log.Printf("Bulk insert failed for %s final batch: %v\n", indexName, err)
		} else {
			log.Printf("Successfully inserted final batch with %d documents into %s", len(docs), indexName)
		}
	}

	log.Printf("Completed sync: %d documents from %s into %s (processed %d batches)\n", count, collectionName, indexName, batchCount)
}

// connectMongo creates and returns a MongoDB client
func connectMongo(uri string) (*mongo.Client, context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		cancel()
		return nil, nil, nil, err
	}
	return client, ctx, cancel, nil
}

func main() {
	log.Println("Starting Elasticsearch data load process...")

	// Initialize Elasticsearch client using your existing InitiateES()
	log.Println("Initializing Elasticsearch client...")
	esClient := searchElastic.InitiateES()
	esHelper := searchElastic.NewESHelper(esClient)
	log.Println("Elasticsearch client initialized successfully")

	// Connect to MongoDB
	log.Println("Connecting to MongoDB...")
	mongoURI := environment.GoDotEnvVariable("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI environment variable is not set")
	}

	mongoClient, ctx, cancel, err := connectMongo(mongoURI)
	if err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
	}
	defer cancel()
	defer mongoClient.Disconnect(ctx)

	// Ping MongoDB to ensure connection is successful before proceeding
	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB ping failed: %v", err)
	}
	log.Println("Successfully connected and pinged MongoDB.")

	dbName := environment.GoDotEnvVariable("DB_NAME")
	if dbName == "" {
		log.Fatal("DB_NAME environment variable is not set")
	}

	db := mongoClient.Database(dbName)
	log.Printf("Using database: %s", dbName)

	// Sync all collections you want to index
	log.Println("Starting data synchronization...")
	syncCollection(ctx, db, esHelper, "post", constants.PostIndexName)
	syncCollection(ctx, db, esHelper, "topic", constants.TopicIndexName)
	syncCollection(ctx, db, esHelper, "customWidget", constants.WidgetIndexName)

	log.Println("Elasticsearch sync completed successfully!")
}
