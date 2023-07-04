package requests

import (
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateActivityRequest | defines create activity type
type CreateActivityRequest struct {
	Action string `json:"action" binding:"required"`
}

// UserActivityResponse | defines activity response schema
type UserActivityResponse struct {
	ID                 primitive.ObjectID       `json:"_id"`
	ActionBy           []string                 `json:"action_by"`
	ActionOn           string                   `json:"action_on"`
	EntityType         constants.EntityType     `json:"entity_type"`
	EntityID           primitive.ObjectID       `json:"entity_id"`
	EntityOwnerID      string                   `json:"entity_owner_id"`
	UUID               string                   `json:"uuid"`
	Action             constants.ActivityAction `json:"action"`
	CTA                string                   `json:"cta"`
	IsRead             bool                     `json:"is_read"`
	CreatedAt          int                      `json:"created_at"`
	UpdatedAt          int                      `json:"updated_at"`
	ActivityEntityData interface{}              `json:"activity_entity_data"`
	ActivityText       string                   `json:"activity_text"`
}
