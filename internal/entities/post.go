package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Structure for Post
type Post struct {
	ID           primitive.ObjectID   `json:"_id" bson:"_id,omitempty"`
	TempId       *string              `json:"temp_id" bson:"temp_id"`
	TopicIds     []primitive.ObjectID `json:"topic_ids" bson:"topic_ids,omitempty"`
	Text         string               `json:"text" bson:"text"`
	Heading      string               `json:"heading" bson:"heading"`
	CommunityId  int                  `json:"community_id" bson:"community_id"`
	ChatroomId   int                  `json:"chatroom_id" bson:"chatroom_id"`
	IsPinned     bool                 `json:"is_pinned" bson:"is_pinned"`
	UserId       string               `json:"user_id" bson:"user_id"`
	Attachments  []Attachment         `json:"attachments" bson:"attachments"`
	IsDeleted    bool                 `json:"is_deleted" bson:"is_deleted"`
	IsEdited     bool                 `json:"is_edited" bson:"is_edited"`
	DeletedBy    string               `json:"deleted_by" bson:"deleted_by,omitempty"`
	DeleteReason string               `json:"delete_reason" bson:"delete_reason,omitempty"`
	CreatedAt    time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at" bson:"updated_at"`
}

// Exposed Method to Create a New Post
func NewPost(text string, heading string, communityId int, userId string, attachments []Attachment,
	chatroomId int, tempId *string, topicIds []primitive.ObjectID) Post {
	createdAt := time.Now()
	return Post{
		Text:        text,
		TempId:      tempId,
		TopicIds:    topicIds,
		Heading:     heading,
		CommunityId: communityId,
		ChatroomId:  chatroomId,
		IsPinned:    false,
		UserId:      userId,
		Attachments: attachments,
		IsDeleted:   false,
		IsEdited:    false,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}
}
