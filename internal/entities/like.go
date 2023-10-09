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
func NewLike(entityType string, entityId primitive.ObjectID, likedBy string, CreatedAt int) Like {
	createdAt := time.Now()

	if CreatedAt > 0 {
		createdAt = time.Unix(0, int64(CreatedAt)*int64(time.Millisecond))
	}

	return Like{
		EntityType: entityType,
		EntityId:   entityId,
		LikedBy:    likedBy,
		IsDeleted:  false,
		CreatedAt:  createdAt,
		UpdatedAt:  createdAt,
	}
}
