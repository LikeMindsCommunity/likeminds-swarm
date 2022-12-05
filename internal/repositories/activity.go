package repositories

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (repository *activityRepository) Create(like *entities.Activity) (interface{}, error) {
	created_at := time.Now()
	like.CreatedAt = created_at
	like.UpdatedAt = created_at

	coll := repository.db.Collection("activity")
	result, err := coll.InsertOne(context.TODO(), like)

	return result.InsertedID, err
}

func (repository *activityRepository) Find(filter map[string]interface{}) ([]entities.Activity, error) {
	var err error

	err = convertMultipleHexIdsToObjectIds(filter, []string{"_id", "entity_id"})
	if err != nil {
		return nil, err
	}

	coll := repository.db.Collection("activity")
	cursor, err := coll.Find(context.TODO(), filter, options.Find())
	if err != nil {
		return nil, err
	}

	var results []entities.Activity
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (repository *activityRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
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

	coll := repository.db.Collection("activity")
	_, err = coll.UpdateOne(context.TODO(), filter, update)

	return err
}

type activityRepository struct {
	db *mongo.Database
}

func NewActivityRepository(db *mongo.Database) interfaces.ActivityRepository {
	return &activityRepository{
		db: db,
	}
}
