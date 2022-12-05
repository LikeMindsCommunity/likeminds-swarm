package repositories

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
)

func (repository *likeRepository) Create(like *entities.Like) (interface{}, error) {
	created_at := time.Now()
	like.CreatedAt = created_at
	like.UpdatedAt = created_at

	coll := repository.db.Collection("like")
	result, err := coll.InsertOne(context.TODO(), like)

	return result.InsertedID, err
}

func (repository *likeRepository) Find(filter map[string]interface{}) ([]entities.Like, error) {
	var err error

	err = convertMultipleHexIdsToObjectIds(filter, []string{"_id", "entity_id"})
	if err != nil {
		return nil, err
	}

	coll := repository.db.Collection("like")
	cursor, err := coll.Find(context.TODO(), filter)
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
	var err error
	var set_data gin.H

	if _, ok := update["$set"]; ok {
		set_data = update["$set"].(gin.H)
	}
	set_data["updated_at"] = time.Now()
	update["$set"] = set_data

	err = convertMultipleHexIdsToObjectIds(filter, []string{"_id", "entity_id"})
	if err != nil {
		return err
	}

	coll := repository.db.Collection("like")
	_, err = coll.UpdateOne(context.TODO(), filter, update)

	return err
}

func (repository *likeRepository) Count(filter map[string]interface{}) (int64, error) {
	var err error

	err = convertMultipleHexIdsToObjectIds(filter, []string{"_id", "entity_id"})
	if err != nil {
		return 0, err
	}

	coll := repository.db.Collection("like")
	count, err := coll.CountDocuments(context.TODO(), filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}

type likeRepository struct {
	db *mongo.Database
}

func NewLikeRepository(db *mongo.Database) interfaces.LikeRepository {
	return &likeRepository{
		db: db,
	}
}
