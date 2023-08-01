package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Structure for Topic
type Topic struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	IsEnabled   bool               `json:"is_enabled" bson:"is_enabled"`
	CommunityId int                `json:"community_id" bson:"community_id"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// Exposed Method to Create a New Topic Instance
func NewTopic(name string, is_enabled bool, community_id int) Topic {
	created_at := time.Now()
	return Topic{
		Name:        name,
		IsEnabled:   is_enabled,
		CommunityId: community_id,
		CreatedAt:   created_at,
		UpdatedAt:   created_at,
	}
}
