package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserEntityTimestamps struct {
	UserId         string             `json:"user_id" bson:"user_id"`
	EntityType     string             `json:"entity_type" bson:"entity_type"`
	EntityID       primitive.ObjectID `json:"entity_id" bson:"entity_id"`
	EpochTimestamp int                `json:"epoch_timestamp" bson:"epoch_timestamp"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
}

func NewUserEntityTimestamps(userId string, entityType string, entityId primitive.ObjectID, epochTimestamp int,
	createdAtInInt int) UserEntityTimestamps {
	createdAt := time.Now()

	if createdAtInInt > 0 {
		createdAt = time.Unix(0, int64(createdAtInInt)*int64(time.Millisecond))
	}

	// Create UserEntityTimestamps entity
	return UserEntityTimestamps{
		UserId:         userId,
		EntityType:     entityType,
		EntityID:       entityId,
		EpochTimestamp: epochTimestamp,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	}
}
