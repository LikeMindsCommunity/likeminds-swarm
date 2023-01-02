package repositories

import (
	"context"

	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (repository *postRepository) Create(post *entities.Post) (interface{}, error) {
	coll := repository.db.Collection("post")
	result, err := coll.InsertOne(context.TODO(), post)

	return result.InsertedID, err
}

func (repository *postRepository) Find(filter map[string]interface{}, filterOpts *options.FindOptions) ([]entities.Post, error) {
	coll := repository.db.Collection("post")
	cursor, err := coll.Find(context.TODO(), filter, filterOpts)
	if err != nil {
		return nil, err
	}

	var results []entities.Post
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (repository *postRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	coll := repository.db.Collection("post")
	_, err := coll.UpdateOne(context.TODO(), filter, update)

	return err
}

func (repository *postRepository) Count(filter map[string]interface{}) (int64, error) {
	coll := repository.db.Collection("post")
	count, err := coll.CountDocuments(context.TODO(), filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}

type postRepository struct {
	db *mongo.Database
}

func NewPostRepository(db *mongo.Database) interfaces.PostRepository {
	return &postRepository{
		db: db,
	}
}
