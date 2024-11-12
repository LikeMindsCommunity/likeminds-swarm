package entities

import (
	"time"

	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ActionByMetadata struct {
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	EntityId  primitive.ObjectID `json:"entity_id" bson:"entity_id"`
}

// Activity | Schema for activity
type Activity struct {
	ID               primitive.ObjectID          `json:"_id" bson:"_id,omitempty"`
	CommunityID      int                         `json:"community_id" bson:"community_id"`
	ActionBy         []string                    `json:"action_by" bson:"action_by"`
	ActionByMetadata map[string]ActionByMetadata `json:"action_by_metadata" bson:"action_by_metadata"`
	ActionOn         string                      `json:"action_on" bson:"action_on"`
	EntityType       constants.EntityType        `json:"entity_type" bson:"entity_type"`
	EntityID         primitive.ObjectID          `json:"entity_id" bson:"entity_id"`
	EntityOwnerID    string                      `json:"entity_owner_id" bson:"entity_owner_id"`
	Action           constants.ActivityAction    `json:"action" bson:"action"`
	ActivityText     string                      `json:"activity_text" bson:"activity_text"`
	CTA              string                      `json:"cta" bson:"cta"`
	IsRead           bool                        `json:"is_read" bson:"is_read"`
	IsDeleted        bool                        `json:"is_deleted" bson:"is_deleted"`
	CreatedAt        time.Time                   `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time                   `json:"updated_at" bson:"updated_at"`
}

// NewActivity | Constructor method to create activity instance
func NewActivity(CommunityID int, ActionBy []string, ActionOn string, EntityType constants.EntityType, EntityID primitive.ObjectID,
	EntityOwnerID string, Action constants.ActivityAction, CTA string, IsRead bool, IsDeleted bool,
	ActionByMetadata map[string]ActionByMetadata, activityText string) Activity {

	TimeNow := time.Now()
	return Activity{
		CommunityID:      CommunityID,
		ActionBy:         ActionBy,
		ActionByMetadata: ActionByMetadata,
		ActionOn:         ActionOn,
		EntityType:       EntityType,
		EntityID:         EntityID,
		EntityOwnerID:    EntityOwnerID,
		Action:           Action,
		ActivityText:     activityText,
		CTA:              CTA,
		IsRead:           IsRead,
		IsDeleted:        IsDeleted,
		CreatedAt:        TimeNow,
		UpdatedAt:        TimeNow,
	}
}
