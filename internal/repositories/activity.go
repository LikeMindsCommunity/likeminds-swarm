package repositories

import (
	"context"

	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (repository *activityRepository) Create(like *entities.Activity) (interface{}, error) {
	coll := repository.db.Collection("activity")
	result, err := coll.InsertOne(context.TODO(), like)

	return result.InsertedID, err
}

func (repository *activityRepository) Find(filter map[string]interface{}, filterOptions *options.FindOptions) ([]entities.Activity, error) {
	coll := repository.db.Collection("activity")
	cursor, err := coll.Find(context.TODO(), filter, filterOptions)
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
	coll := repository.db.Collection("activity")
	_, err := coll.UpdateOne(context.TODO(), filter, update)

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
