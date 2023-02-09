package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Structure for Like
type Like struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	EntityType string             `json:"entity_type" bson:"entity_type"`
	EntityId   primitive.ObjectID `json:"entity_id" bson:"entity_id"`
	LikedBy    string             `json:"liked_by" bson:"liked_by"`
	IsDeleted  bool               `json:"is_deleted" bson:"is_deleted"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}

// Exposed Method to Create a New Like
func NewLike(entity_type string, entity_id primitive.ObjectID, liked_by string) Like {
	created_at := time.Now()
	return Like{
		EntityType: entity_type,
		EntityId:   entity_id,
		LikedBy:    liked_by,
		IsDeleted:  false,
		CreatedAt:  created_at,
		UpdatedAt:  created_at,
	}
}
