package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Structure for Comment
type Comment struct {
	ID           primitive.ObjectID   `json:"_id" bson:"_id,omitempty"`
	TempID       string               `json:"temp_id" bson:"temp_id"`
	Text         string               `json:"text" bson:"text"`
	PostId       primitive.ObjectID   `json:"post_id" bson:"post_id"`
	CommunityId  int                  `json:"community_id" bson:"community_id"`
	Level        int                  `json:"level" bson:"level"`
	Replies      []primitive.ObjectID `json:"replies" bson:"replies"`
	UserId       string               `json:"user_id" bson:"user_id"`
	IsEdited     bool                 `json:"is_edited" bson:"is_edited"`
	IsDeleted    bool                 `json:"is_deleted" bson:"is_deleted"`
	DeletedBy    string               `json:"deleted_by" bson:"deleted_by,omitempty"`
	DeleteReason string               `json:"delete_reason" bson:"delete_reason,omitempty"`
	CreatedAt    time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at" bson:"updated_at"`
}

// Exposed Method to Create a New Comment
func NewComment(text string, postId primitive.ObjectID, communityId int, level int, userId string, tempId string) Comment {
	createdAt := time.Now()
	return Comment{
		Text:        text,
		TempID:      tempId,
		PostId:      postId,
		CommunityId: communityId,
		Level:       level,
		UserId:      userId,
		Replies:     []primitive.ObjectID{},
		IsDeleted:   false,
		IsEdited:    false,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}
}
