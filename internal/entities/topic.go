package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Structure for Topic
type Topic struct {
	ID              primitive.ObjectID   `json:"_id" bson:"_id,omitempty"`
	Name            string               `json:"name" bson:"name"`
	IsEnabled       bool                 `json:"is_enabled" bson:"is_enabled"`
	CommunityId     int                  `json:"community_id" bson:"community_id"`
	Priority        float32              `json:"priority" bson:"priority"`
	IsSearchable    bool                 `json:"is_searchable" bson:"is_searchable"`
	ParentId        primitive.ObjectID   `json:"parent_id" bson:"parent_id"`
	ParentName      string               `json:"parent_name" bson:"parent_name"`
	AllParentIds    []primitive.ObjectID `json:"all_parent_ids" bson:"all_parent_ids"`
	Level           int                  `json:"level" bson:"level"`
	WidgetId        primitive.ObjectID   `json:"widget_id" bson:"widget_id"`
	TotalChildCount int                  `json:"total_child_count" bson:"total_child_count"`
	Access          string               `json:"access" bson:"access"`
	CreatedAt       time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at" bson:"updated_at"`
}

// Exposed Method to Create a New Topic Instance
func NewTopic(name string, isEnabled bool, priority float32, isSearchable bool, parentId primitive.ObjectID, parentName string,
	allParentIds []primitive.ObjectID, level int, widgetId primitive.ObjectID, totalChildCount int, access string, communityId int) Topic {
	createdAt := time.Now()

	return Topic{
		Name:            name,
		IsEnabled:       isEnabled,
		Priority:        priority,
		IsSearchable:    isSearchable,
		ParentId:        parentId,
		ParentName:      parentName,
		AllParentIds:    allParentIds,
		Level:           level,
		WidgetId:        widgetId,
		TotalChildCount: totalChildCount,
		Access:          access,
		CommunityId:     communityId,
		CreatedAt:       createdAt,
		UpdatedAt:       createdAt,
	}
}
