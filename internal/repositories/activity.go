package repositories

import (
	"context"

	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Exposed Repository Method to Create Activity
func (repository *activityRepository) Create(like *entities.Activity) (interface{}, error) {
	coll := repository.db.Collection("activity")
	result, err := coll.InsertOne(context.TODO(), like)

	return result.InsertedID, err
}

// Exposed Repository Method to Find Activity
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

// Exposed Repository Method to Update Activity
func (repository *activityRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	coll := repository.db.Collection("activity")
	_, err := coll.UpdateOne(context.TODO(), filter, update)

	return err
}

// Structure for Activity Repository
type activityRepository struct {
	db *mongo.Database
}

// Exposed Method to Create New Activity Repository
func NewActivityRepository(db *mongo.Database) interfaces.ActivityRepository {
	return &activityRepository{
		db: db,
	}
}
