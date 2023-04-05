package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Structure for Post
type Post struct {
	ID           primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Text         string             `json:"text" bson:"text"`
	Heading      string             `json:"heading" bson:"heading"`
	CommunityId  int                `json:"community_id" bson:"community_id"`
	ChatroomId   int                `json:"chatroom_id" bson:"chatroom_id"`
	IsPinned     bool               `json:"is_pinned" bson:"is_pinned"`
	UserId       string             `json:"user_id" bson:"user_id"`
	Attachments  []Attachment       `json:"attachments" bson:"attachments"`
	IsDeleted    bool               `json:"is_deleted" bson:"is_deleted"`
	IsEdited     bool               `json:"is_edited" bson:"is_edited"`
	DeletedBy    string             `json:"deleted_by" bson:"deleted_by,omitempty"`
	DeleteReason string             `json:"delete_reason" bson:"delete_reason,omitempty"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

// Exposed Method to Create a New Post
func NewPost(text string, heading string, community_id int, user_id string, attachments []Attachment, chatroom_id int) Post {
	created_at := time.Now()
	return Post{
		Text:        text,
		Heading:     heading,
		CommunityId: community_id,
		ChatroomId:  chatroom_id,
		IsPinned:    false,
		UserId:      user_id,
		Attachments: attachments,
		IsDeleted:   false,
		IsEdited:    false,
		CreatedAt:   created_at,
		UpdatedAt:   created_at,
	}
}
