package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Structure for PollVotes
type PollVotes struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	PollId      primitive.ObjectID `json:"poll_id" bson:"poll_id"`
	UUID        string             `json:"uuid" bson:"uuid"`
	Votes       []string           `json:"votes" bson:"votes"`
	CommunityId int                `json:"community_id" bson:"community_id"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// Exposed Method to Create a New Poll Vote for a User
func NewPollVotes(pollId primitive.ObjectID, uuid string, votes []string, communityId int) PollVotes {
	createdAt := time.Now()
	return PollVotes{
		PollId:      pollId,
		UUID:        uuid,
		Votes:       votes,
		CommunityId: communityId,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}
}
