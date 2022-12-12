package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	ID           primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Text         string             `json:"text" bson:"text"`
	ApiKey       string             `json:"api_key" bson:"api_key"`
	IsPinned     bool               `json:"is_pinned" bson:"is_pinned"`
	UserId       string             `json:"user_id" bson:"user_id"`
	Attachments  []Widget           `json:"attachments" bson:"attachments"`
	IsDeleted    bool               `json:"is_deleted" bson:"is_deleted"`
	DeletedBy    string             `json:"deleted_by" bson:"deleted_by,omitempty"`
	DeleteReason string             `json:"delete_reason" bson:"delete_reason,omitempty"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

func NewPost(text string, api_key string, user_id string, attachments []Widget) Post {
	created_at := time.Now()
	return Post{
		Text:        text,
		ApiKey:      api_key,
		IsPinned:    false,
		UserId:      user_id,
		Attachments: attachments,
		IsDeleted:   false,
		CreatedAt:   created_at,
		UpdatedAt:   created_at,
	}
}
