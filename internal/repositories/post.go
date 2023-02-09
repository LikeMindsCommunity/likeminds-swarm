package repositories

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Exposed Helper Method to Create Post
func (repository *postRepository) Create(post *entities.Post) (interface{}, error) {
	coll := repository.db.Collection("post")
	result, err := coll.InsertOne(context.TODO(), post)

	return result.InsertedID, err
}

// Exposed Helper Method to Find Post
func (repository *postRepository) Find(filter map[string]interface{}, filterOpts *options.FindOptions) ([]entities.Post, error) {
	coll := repository.db.Collection("post")
	cursor, err := coll.Find(context.TODO(), filter, filterOpts)
	if err != nil {
		return nil, err
	}

	var results []entities.Post
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}

// Exposed Helper Method to Update Post
func (repository *postRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	coll := repository.db.Collection("post")
	_, err := coll.UpdateOne(context.TODO(), filter, update)

	return err
}

// Exposed Helper Method to Fetch Post Count
func (repository *postRepository) Count(filter map[string]interface{}) (int64, error) {
	coll := repository.db.Collection("post")
	count, err := coll.CountDocuments(context.TODO(), filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Exposed Helper Method to perform Aggregration on Post
func (repository *postRepository) Aggregate(query []map[string]interface{}) ([]gin.H, error) {
	coll := repository.db.Collection("post")
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

// Structure for Post Repository
type postRepository struct {
	db *mongo.Database
}

// Exposed Method to Create New Post Repository
func NewPostRepository(db *mongo.Database) interfaces.PostRepository {
	return &postRepository{
		db: db,
	}
}
