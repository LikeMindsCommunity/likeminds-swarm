package repositories

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (repository *likeRepository) Create(like *entities.Like) (interface{}, error) {
	coll := repository.db.Collection("like")
	result, err := coll.InsertOne(context.TODO(), like)

	return result.InsertedID, err
}

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

func (repository *likeRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	coll := repository.db.Collection("like")
	_, err := coll.UpdateOne(context.TODO(), filter, update)

	return err
}

func (repository *likeRepository) CountById(filter map[string]interface{}) (int64, error) {
	coll := repository.db.Collection("like")
	count, err := coll.CountDocuments(context.TODO(), filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}

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

type likeRepository struct {
	db *mongo.Database
}

func NewLikeRepository(db *mongo.Database) interfaces.LikeRepository {
	return &likeRepository{
		db: db,
	}
}
