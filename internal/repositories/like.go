package repositories

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Exposed Helper Method to Create Like
func (repository *likeRepository) Create(like *entities.Like) (interface{}, error) {
	coll := repository.db.Collection("like")
	result, err := coll.InsertOne(context.TODO(), like)

	return result.InsertedID, err
}

// Exposed Helper Method to Find Like
func (repository *likeRepository) Find(filter map[string]interface{}, filterOpts *options.FindOptions) ([]entities.Like, error) {
	coll := repository.db.Collection("like")
	cursor, err := coll.Find(context.TODO(), filter, filterOpts)
	if err != nil {
		return nil, err
	}

	var results []entities.Like
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}

// Exposed Helper Method to Update Like
func (repository *likeRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	coll := repository.db.Collection("like")
	_, err := coll.UpdateOne(context.TODO(), filter, update)

	return err
}

// Exposed Helper Method to Fetch Likes Count
func (repository *likeRepository) Count(filter map[string]interface{}) (int64, error) {
	coll := repository.db.Collection("like")
	count, err := coll.CountDocuments(context.TODO(), filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Exposed Helper Method to perform Aggregration on Likes
func (repository *likeRepository) Aggregate(query []interface{}) (interface{}, error) {
	coll := repository.db.Collection("like")
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

// Structure for Like Repository
type likeRepository struct {
	db *mongo.Database
}

// Exposed Method to Create New Like Repository
func NewLikeRepository(db *mongo.Database) interfaces.LikeRepository {
	return &likeRepository{
		db: db,
	}
}
