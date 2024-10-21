package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Structure for Post Topics
type PostTopic struct {
	ID              primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	CommunityId     int                `json:"community_id" bson:"community_id"`
	PostID          primitive.ObjectID `json:"post_id" bson:"post_id"`
	TopicID         primitive.ObjectID `json:"topic_id" bson:"topic_id"`
	IfOriginalTopic bool               `json:"if_original_topic" bson:"original_topic"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}

// Exposed Method to Create a New Topic Instance
func NewPostTopic(postId primitive.ObjectID, topicId primitive.ObjectID, ifOriginalTopic bool, communityId int,
) PostTopic {
	createdAt := time.Now()
	return PostTopic{
		PostID:          postId,
		TopicID:         topicId,
		IfOriginalTopic: ifOriginalTopic,
		CommunityId:     communityId,
		CreatedAt:       createdAt,
		UpdatedAt:       createdAt,
	}
}
