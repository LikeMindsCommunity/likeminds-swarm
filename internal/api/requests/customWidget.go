package requests

import "go.mongodb.org/mongo-driver/bson/primitive"

// Request Structure for Create Custom Widget
type CreateCustomWidgetRequest struct {
	ParentEntityID   string                 `json:"parent_entity_id" binding:"required"`
	ParentEntityType string                 `json:"parent_entity_type" binding:"required"`
	MetaData         map[string]interface{} `json:"metadata"`
}

// Request Structure for Edit Custom Widget
type EditCustomWidgetRequest struct {
	MetaData map[string]interface{} `json:"metadata"`
}

// Request Structure for Fetch Custom Widget
type FetchCustomWidgetRequest struct {
	SearchKey        string `form:"search_key"`
	SearchValue      string `form:"search_value"`
	ParentEntityId   string `form:"parent_entity_id"`
	ParentEntityType string `form:"parent_entity_type"`
}

// Response Structure for Custom Widget
type CustomWidgetResponse struct {
	ID               primitive.ObjectID     `json:"_id"`
	ParentEntityID   string                 `json:"parent_entity_id"`
	ParentEntityType string                 `json:"parent_entity_type"`
	MetaData         map[string]interface{} `json:"metadata"`
	CreatedAt        int                    `json:"created_at"`
	UpdatedAt        int                    `json:"updated_at"`
}
