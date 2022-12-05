package repositories

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
)

func (repository *saveRepository) Create(like *entities.Save) (interface{}, error) {
	created_at := time.Now()
	like.CreatedAt = created_at
	like.UpdatedAt = created_at

	coll := repository.db.Collection("save")
	result, err := coll.InsertOne(context.TODO(), like)

	return result.InsertedID, err
}

func (repository *saveRepository) Find(filter map[string]interface{}) ([]entities.Save, error) {
	var err error

	err = convertMultipleHexIdsToObjectIds(filter, []string{"_id", "entity_id"})
	if err != nil {
		return nil, err
	}

	coll := repository.db.Collection("save")
	cursor, err := coll.Find(context.TODO(), filter)
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

	coll := repository.db.Collection("save")
	_, err = coll.UpdateOne(context.TODO(), filter, update)

	return err
}

type saveRepository struct {
	db *mongo.Database
}

func NewSaveRepository(db *mongo.Database) interfaces.SaveRepository {
	return &saveRepository{
		db: db,
	}
}
