package requests

import "go.mongodb.org/mongo-driver/bson/primitive"

// Request Structure for Create Custom Widget
type CreateWidgetRequest struct {
	ParentEntityID   string                 `json:"parent_entity_id" binding:"required"`
	ParentEntityType string                 `json:"parent_entity_type" binding:"required"`
	MetaData         map[string]interface{} `json:"metadata"`
	UserIsCM         bool                   `json:"user_is_cm"`
}

// Request Structure for Edit Custom Widget
type EditWidgetRequest struct {
	MetaData map[string]interface{} `json:"metadata"`
	UserIsCM bool                   `json:"user_is_cm"`
}

// Request Structure for Fetch Custom Widget
type FetchWidgetRequest struct {
	SearchKey        string `form:"search_key"`
	SearchValue      string `form:"search_value"`
	ParentEntityId   string `form:"parent_entity_id"`
	ParentEntityType string `form:"parent_entity_type"`
	WidgetIds        string `form:"widget_ids"`
}

// Response Structure for Custom Widget
type WidgetResponse struct {
	ID               primitive.ObjectID     `json:"_id"`
	ParentEntityID   string                 `json:"parent_entity_id"`
	ParentEntityType string                 `json:"parent_entity_type"`
	MetaData         map[string]interface{} `json:"metadata"`
	LMMeta           map[string]interface{} `json:"_lm_meta"`
	CreatedAt        int                    `json:"created_at"`
	UpdatedAt        int                    `json:"updated_at"`
}
