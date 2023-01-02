package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Activity struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	ActionBy    string             `json:"action_by" bson:"action_by"`
	ActionOn    []string           `json:"action_on" bson:"action_on"`
	CommunityId int                `json:"community_id" bson:"community_id"`
	EntityType  string             `json:"entity_type" bson:"entity_type"`
	EntityId    primitive.ObjectID `json:"entity_id" bson:"entity_id"`
	Action      string             `json:"action" bson:"action"`
	CTA         string             `json:"cta" bson:"cta"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

func NewActivity(action_by string, action_on []string, community_id int, entity_type string, entity_id primitive.ObjectID,
	action string, cta string) Activity {
	created_at := time.Now()
	return Activity{
		ActionBy:    action_by,
		ActionOn:    action_on,
		CommunityId: community_id,
		EntityType:  entity_type,
		EntityId:    entity_id,
		Action:      action,
		CTA:         cta,
		CreatedAt:   created_at,
		UpdatedAt:   created_at,
	}
}
