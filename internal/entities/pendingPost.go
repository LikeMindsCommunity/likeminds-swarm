package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Structure for Pending Post
type PendingPost struct {
	ID           primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	PostData     Post               `json:"post_data" bson:"post_data"`
	Status       string             `json:"status" bson:"status"`
	UserId       string             `json:"user_id" bson:"user_id"`
	CommunityID  int                `json:"community_id" bson:"community_id"`
	IsDeleted    bool               `json:"is_deleted" bson:"is_deleted"`
	UUIDs        []string           `json:"uuids" bson:"uuids"`
	ReportID     int                `json:"-" bson:"report_id"`
	NormalPostId string             `json:"-" bson:"normal_post_id"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

// Exposed Method to Create a New Pending Post
func NewPendingPost(text string, heading string, communityId int, userId string, attachments []Attachment,
	chatroomId int, tempId *string, topicIds []primitive.ObjectID, originalAuthorUUID string,
	visibility string, isRepost bool, createdAt int, status string, UUIDs []string) PendingPost {

	post := NewPost(text, heading, communityId, userId, attachments, chatroomId, tempId, topicIds, originalAuthorUUID,
		visibility, isRepost, createdAt)

	// create pending post entity
	pendingPostEntity := PendingPost{
		PostData:    post,
		Status:      status,
		UserId:      userId,
		CommunityID: communityId,
		IsDeleted:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		UUIDs:       UUIDs,
	}

	return pendingPostEntity
}
