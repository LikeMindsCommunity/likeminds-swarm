package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Structure for Save
type Save struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	EntityType  string             `json:"entity_type" bson:"entity_type"`
	EntityId    primitive.ObjectID `json:"entity_id" bson:"entity_id"`
	CommunityId int                `json:"community_id" bson:"community_id"`
	SavedBy     string             `json:"saved_by" bson:"saved_by"`
	IsDeleted   bool               `json:"is_deleted" bson:"is_deleted"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// Exposed Method to Create a New Save Instance
func NewSave(entityType string, entityId primitive.ObjectID, savedBy string, communityId int) Save {
	createdAt := time.Now()
	return Save{
		EntityType:  entityType,
		EntityId:    entityId,
		CommunityId: communityId,
		SavedBy:     savedBy,
		IsDeleted:   false,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}
}
