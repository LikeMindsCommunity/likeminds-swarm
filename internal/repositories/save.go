package repositories

import (
	"context"

	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (repository *saveRepository) Create(like *entities.Save) (interface{}, error) {
	coll := repository.db.Collection("save")
	result, err := coll.InsertOne(context.TODO(), like)

	return result.InsertedID, err
}

func (repository *saveRepository) Find(filter map[string]interface{}, filterOpts *options.FindOptions) ([]entities.Save, error) {
	coll := repository.db.Collection("save")
	cursor, err := coll.Find(context.TODO(), filter, filterOpts)
	if err != nil {
		return nil, err
	}

	var results []entities.Save
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (repository *saveRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	coll := repository.db.Collection("save")
	_, err := coll.UpdateOne(context.TODO(), filter, update)

	return err
}

func (repository *saveRepository) Count(filter map[string]interface{}) (int64, error) {
	coll := repository.db.Collection("save")
	count, err := coll.CountDocuments(context.TODO(), filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}

type saveRepository struct {
	db *mongo.Database
}

func NewSaveRepository(db *mongo.Database) interfaces.SaveRepository {
	return &saveRepository{
		db: db,
	}
}
