package requests

import (
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateActivityRequest | defines create activity type
type CreateActivityRequest struct {
	Action       string   `json:"action" binding:"required"`
	ActionBy     string   `json:"action_by" binding:"required"`
	ActionOn     []string `json:"action_on" binding:"required"`
	EntityType   string   `json:"entity_type" binding:"required"`
	EntityId     string   `json:"entity_id"`
	ActivityText string   `json:"activity_text"`
}

type UserActivityResponse struct {
	ID                 primitive.ObjectID `json:"_id"`
	ActionBy           []string           `json:"action_by"`
	ActionOn           string             `json:"action_on"`
	EntityID           primitive.ObjectID `json:"entity_id"`
	EntityOwnerID      string             `json:"entity_owner_id"`
	UUID               string             `json:"uuid"`
	CTA                string             `json:"cta"`
	IsRead             bool               `json:"is_read"`
	CreatedAt          int                `json:"created_at"`
	UpdatedAt          int                `json:"updated_at"`
	ActivityEntityData interface{}        `json:"activity_entity_data"`
	ActivityText       string             `json:"activity_text"`
}

// UserActivityResponse | defines activity response schema
type UserActivityResponseOld struct {
	UserActivityResponse
	EntityType constants.EntityType     `json:"entity_type"`
	Action     constants.ActivityAction `json:"action"`
}

type UserActivityResponseV1 struct {
	UserActivityResponse
	EntityType enums.EntityType     `json:"entity_type"`
	Action     enums.ActivityAction `json:"action"`
}
