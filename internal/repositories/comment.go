package repositories

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
)

func (repository *commentRepository) Create(comment *entities.Comment) (interface{}, error) {
	created_at := time.Now()
	comment.CreatedAt = created_at
	comment.UpdatedAt = created_at

	coll := repository.db.Collection("comment")
	result, err := coll.InsertOne(context.TODO(), comment)

	if err != nil {
		return nil, err
	}
	return result.InsertedID, err
}

func (repository *commentRepository) Find(filter map[string]interface{}) ([]entities.Comment, error) {
	var err error

	err = convertMultipleHexIdsToObjectIds(filter, []string{"_id", "post_id"})
	if err != nil {
		return nil, err
	}

	coll := repository.db.Collection("comment")
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	var results []entities.Comment
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (repository *commentRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
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

	coll := repository.db.Collection("comment")
	_, err = coll.UpdateOne(context.TODO(), filter, update)

	return err
}

type commentRepository struct {
	db *mongo.Database
}

func NewCommentRepository(db *mongo.Database) interfaces.CommentRepository {
	return &commentRepository{
		db: db,
	}
}
