package interfaces

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/entities"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Interface for PollVotes Repository
type PollVotesRepository interface {
	Create(document interface{}) (interface{}, error)
	Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
	Aggregate(query []map[string]interface{}) ([]gin.H, error)
}

// Interface for PollVotes Helper
type PollVotesHelper interface {
	CreatePollVotesHelper(pollId primitive.ObjectID, uuid string, votes []string, community_id int) (interface{}, error)
	FindPollVotesHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.PollVotes, error)
	UpdatePollVotesByIdHelper(pollVotesId primitive.ObjectID, update map[string]interface{}) error
	CountPollVotesHelper(filter map[string]interface{}) (int64, error)
	AggregatePollVotesHelper(query []map[string]interface{}) ([]gin.H, error)
}
