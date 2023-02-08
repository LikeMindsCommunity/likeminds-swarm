package requests

import "go.mongodb.org/mongo-driver/bson/primitive"

// Request Structure for Create Activity
type CreateActivityRequest struct {
	Action string `json:"action" binding:"required"`
}

// Response Structure for User Activity
type UserActivityResponse struct {
	ID              primitive.ObjectID `json:"_id"`
	ActionBy        string             `json:"action_by"`
	ActionOn        []string           `json:"action_on"`
	CommunityId     int                `json:"community_id"`
	EntityType      string             `json:"entity_type"`
	EntityId        primitive.ObjectID `json:"entity_id"`
	Action          string             `json:"action"`
	CTA             string             `json:"cta"`
	ActivityMessage string             `json:"activity_message"`
	CreatedAt       int                `json:"created_at"`
	UpdatedAt       int                `json:"updated_at"`
}
