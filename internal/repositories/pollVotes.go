package repositories

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Exposed Helper Method to Create Poll Votes
func (repository *pollVotesRepository) Create(document interface{}) (interface{}, error) {
	return _createDocumentInDB(repository.db, PollVotesCollection, document)
}

// Exposed Helper Method to Find Topics
func (repository *pollVotesRepository) Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error) {
	return _findDocumentsInDB(repository.db, PollVotesCollection, filter, filterOpts)
}

// Exposed Helper Method to Update Topics
func (repository *pollVotesRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateDocumentsInDB(repository.db, PollVotesCollection, filter, update)
}

// Exposed Helper Method to Fetch Topics Count
func (repository *pollVotesRepository) Count(filter map[string]interface{}) (int64, error) {
	return _countDocumentsInDB(repository.db, PollVotesCollection, filter)
}

// Exposed Helper Method to perform Aggregration on Post
func (repository *pollVotesRepository) Aggregate(query []map[string]interface{}) ([]gin.H, error) {
	return _aggregateDocumentsInDB(repository.db, PollVotesCollection, query)
}

// Structure for PollVotes Repository
type pollVotesRepository struct {
	db *mongo.Database
}

// Exposed Method to Create New Widget Repository
func NewPollVotesRepository(db *mongo.Database) interfaces.PollVotesRepository {
	return &pollVotesRepository{
		db: db,
	}
}
