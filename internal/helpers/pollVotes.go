package helpers

import (
	"context"
	"time"

	"github.com/LikeMindsCommunity/likeminds-swarm/internal/entities"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/interfaces"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exposed Helper Method to Create PollVotes Instance
func (helper *pollVotesHelper) CreatePollVotesHelper(pollId primitive.ObjectID, uuid string, votes []string,
	community_id int) (interface{}, error) {
	// Create a new PollVotes Document
	pollVotes := entities.NewPollVotes(pollId, uuid, votes, community_id)

	// Insert the document in the collection
	pollVotesId, err := helper.pollVotesRepository.Create(pollVotes)

	return pollVotesId, err
}

// Exposed Helper Method to Find PollVotes Instances
func (helper *pollVotesHelper) FindPollVotesHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.PollVotes,
	error) {
	fOpts := mergeFilterOptions(filterOptions)

	// Parse the object Ids
	err := convertHexIdsToObjectIds(filter, []string{"_id", "poll_id"})
	if err != nil {
		return nil, err
	}

	// Find the document in the collection
	cursor, err := helper.pollVotesRepository.Find(filter, &fOpts)
	if err != nil {
		return nil, err
	}

	// Parse the results from fetched documents
	var results []entities.PollVotes
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, err
}

// Exposed Helper Method to Update PollVotes Instance
func (helper *pollVotesHelper) UpdatePollVotesByIdHelper(pollVotesId primitive.ObjectID, update map[string]interface{}) error {
	setData := gin.H{}

	// Create set filter
	if _, ok := update["$set"]; ok {
		setData = update["$set"].(gin.H)
	}
	setData["updated_at"] = time.Now()
	update["$set"] = setData

	// Update the document in the collection
	err := helper.pollVotesRepository.Update(gin.H{"_id": pollVotesId}, update)

	return err
}

// Exposed Helper Method to Fetch PollVotes Count
func (helper *pollVotesHelper) CountPollVotesHelper(filter map[string]interface{}) (int64, error) {
	// Parse the object IDs
	err := convertHexIdsToObjectIds(filter, []string{"_id", "poll_id"})
	if err != nil {
		return 0, err
	}

	// Get count of documents from the collection
	count, err := helper.pollVotesRepository.Count(filter)

	return count, err
}

// Exposed Helper Method to perform Aggregration on PollVotes
func (helper *pollVotesHelper) AggregatePollVotesHelper(query []map[string]interface{}) ([]gin.H, error) {
	for _, value := range query {
		if matchGroup, ok := value["$match"]; ok {
			err := convertHexIdsToObjectIds(matchGroup.(bson.M), []string{"_id", "poll_id"})
			if err != nil {
				return nil, err
			}
		}
	}

	results, err := helper.pollVotesRepository.Aggregate(query)

	return results, err
}

// Structure for PollVotes Helper
type pollVotesHelper struct {
	pollVotesRepository interfaces.PollVotesRepository
}

// Exposed Method to Create New PollVotes Helper
func NewPollVotesHelper(pollVotesRepository interfaces.PollVotesRepository) interfaces.PollVotesHelper {
	return &pollVotesHelper{
		pollVotesRepository: pollVotesRepository,
	}
}
