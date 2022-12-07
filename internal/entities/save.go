package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Save struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	EntityType string             `json:"entity_type" bson:"entity_type"`
	EntityId   primitive.ObjectID `json:"entity_id" bson:"entity_id"`
	SavedBy    string             `json:"saved_by" bson:"saved_by"`
	IsDeleted  bool               `json:"is_deleted" bson:"is_deleted"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}

func NewSave(entity_type string, entity_id primitive.ObjectID, saved_by string) Save {
	created_at := time.Now()
	return Save{
		EntityType: entity_type,
		EntityId:   entity_id,
		SavedBy:    saved_by,
		IsDeleted:  false,
		CreatedAt:  created_at,
		UpdatedAt:  created_at,
	}
}
