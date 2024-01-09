package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Structure for Pending Post
type PendingPost struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	PostData    Post               `json:"post_data" bson:"post_data"`
	PostType    string             `json:"post_type" bson:"post_type"`
	UserId      string             `json:"user_id" bson:"user_id"`
	CommunityID int                `json:"community_id" bson:"community_id"`
	IsDeleted   bool               `json:"is_deleted" bson:"is_deleted"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// Exposed Method to Create a New Pending Post
func NewPendingPost(text string, heading string, communityId int, userId string, attachments []Attachment,
	chatroomId int, tempId *string, topicIds []primitive.ObjectID, originalAuthorUUID string,
	visibility string, CreatedAt int, postType string) PendingPost {

	post := NewPost(text, heading, communityId, userId, attachments, chatroomId, tempId, topicIds, originalAuthorUUID, visibility, CreatedAt)

	// create pending post entity
	pendingPostEntity := PendingPost{
		PostData:    post,
		PostType:    postType,
		UserId:      userId,
		CommunityID: communityId,
		IsDeleted:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return pendingPostEntity
}
