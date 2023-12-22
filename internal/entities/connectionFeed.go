package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Structure for ConnectionFeed
type ConnectionFeed struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	PostId      primitive.ObjectID `json:"post_id" bson:"post_id"`
	UserId      string             `json:"user_id" bson:"user_id"`
	CommunityId int                `json:"community_id" bson:"community_id"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// Exposed Method to Create a New User ConnectionFeed Item
func NewConnectionFeedItem(postId primitive.ObjectID, userId string, communityId int) ConnectionFeed {
	createdAt := time.Now()

	return ConnectionFeed{
		PostId:      postId,
		UserId:      userId,
		CommunityId: communityId,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}
}
