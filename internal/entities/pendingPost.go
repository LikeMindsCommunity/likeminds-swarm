package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

// Structure for Pending Post
type PendingPost struct {
	Post
	PostType string `json:"post_type" bson:"post_type"`
}

// Exposed Method to Create a New Pending Post
func NewPendingPost(text string, heading string, communityId int, userId string, attachments []Attachment,
	chatroomId int, tempId *string, topicIds []primitive.ObjectID, originalAuthorUUID string,
	visibility string, CreatedAt int, postType string) PendingPost {

	post := NewPost(text, heading, communityId, userId, attachments, chatroomId, tempId, topicIds, originalAuthorUUID, visibility, CreatedAt)

	// create pending post entity
	pendingPostEntity := PendingPost{
		Post:     post,
		PostType: postType,
	}

	return pendingPostEntity
}
