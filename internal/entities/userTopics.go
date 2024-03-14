package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserTopic struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	UserID      string             `json:"user_id" bson:"user_id"`
	TopicID     primitive.ObjectID `json:"topic_id" bson:"topic_id"`
	CommunityID int                `json:"community_id" bson:"community_id"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

func NewUserTopic(userId string, topicId primitive.ObjectID, communityId int) UserTopic {
	createdAt := time.Now()
	return UserTopic{
		UserID:      userId,
		TopicID:     topicId,
		CommunityID: communityId,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}
}
