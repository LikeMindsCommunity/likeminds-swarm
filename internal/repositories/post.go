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

func (repository *postRepository) Create(post *entities.Post) error {
	created_at := time.Now()
	post.CreatedAt = created_at
	post.UpdatedAt = created_at

	coll := repository.db.Collection("post")
	_, err := coll.InsertOne(context.TODO(), post)

	return err
}

func (repository *postRepository) Find(filter map[string]interface{}) ([]entities.Post, error) {
	var err error
	filterOptions := options.Find()

	err = convertMultipleHexIdsToObjectIds(filter, []string{"_id"})
	if err != nil {
		return nil, err
	}

	coll := repository.db.Collection("post")
	cursor, err := coll.Find(context.TODO(), filter, filterOptions)
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
	var err error
	update["updated_at"] = time.Now()
	update_data := gin.H{"$set": update}

	err = convertMultipleHexIdsToObjectIds(filter, []string{"_id"})
	if err != nil {
		return err
	}

	coll := repository.db.Collection("post")
	_, err = coll.UpdateOne(context.TODO(), filter, update_data)

	return err
}

type postRepository struct {
	db *mongo.Database
}

func NewPostRepository(db *mongo.Database) interfaces.PostRepository {
	return &postRepository{
		db: db,
	}
}
