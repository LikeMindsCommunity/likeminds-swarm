package repositories

import (
	"context"

	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Exposed Repository Method to Create Comment
func (repository *commentRepository) Create(comment *entities.Comment) (interface{}, error) {
	coll := repository.db.Collection("comment")
	result, err := coll.InsertOne(context.TODO(), comment)

	if err != nil {
		return nil, err
	}
	return result.InsertedID, err
}

// Exposed Repository Method to Find Comment
func (repository *commentRepository) Find(filter map[string]interface{}, filterOptions *options.FindOptions) ([]entities.Comment, error) {
	coll := repository.db.Collection("comment")
	cursor, err := coll.Find(context.TODO(), filter, filterOptions)
	if err != nil {
		return nil, err
	}

	var results []entities.Comment
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}

// Exposed Repository Method to Update Comment
func (repository *commentRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	coll := repository.db.Collection("comment")
	_, err := coll.UpdateOne(context.TODO(), filter, update)

	return err
}

// Exposed Repository Method to Find Comment Count
func (repository *commentRepository) Count(filter map[string]interface{}) (int64, error) {
	coll := repository.db.Collection("comment")
	count, err := coll.CountDocuments(context.TODO(), filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Structure for Comment Repository
type commentRepository struct {
	db *mongo.Database
}

// Exposed Method to Create New Comment Repository
func NewCommentRepository(db *mongo.Database) interfaces.CommentRepository {
	return &commentRepository{
		db: db,
	}
}
