package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Structure for Widget
type Widget struct {
	ID               primitive.ObjectID     `json:"_id" bson:"_id,omitempty"`
	CreatedByLM      bool                   `json:"created_by_lm" bson:"created_by_lm"`
	ParentEntityID   string                 `json:"parent_entity_id" bson:"parent_entity_id"`
	ParentEntityType string                 `json:"parent_entity_type" bson:"parent_entity_type"`
	MetaData         map[string]interface{} `json:"metadata" bson:"metadata"`
	LMMeta           map[string]interface{} `json:"_lm_meta" bson:"_lm_meta"`
	CommunityId      int                    `json:"community_id" bson:"community_id"`
	CreatedAt        time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" bson:"updated_at"`
}

// Exposed Method to Create a New Widget Instance
func NewWidget(createdByLM bool, parentEntityID string, parentEntityType string, metaData map[string]interface{},
	lmMeta map[string]interface{}, communityId int) Widget {
	createdAt := time.Now()

	return Widget{
		CreatedByLM:      createdByLM,
		ParentEntityID:   parentEntityID,
		ParentEntityType: parentEntityType,
		MetaData:         metaData,
		LMMeta:           lmMeta,
		CommunityId:      communityId,
		CreatedAt:        createdAt,
		UpdatedAt:        createdAt,
	}
}
