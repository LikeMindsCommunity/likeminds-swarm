package repositories

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Exposed Helper Method to Create Like
func (repository *connectionFeedRepository) Create(document interface{}) (interface{}, error) {
	return _createDocumentInDB(repository.db, ConnectionFeedCollection, document)
}

// Exposed Helper Method to Find Like
func (repository *connectionFeedRepository) Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error) {
	return _findDocumentsInDB(repository.db, ConnectionFeedCollection, filter, filterOpts)
}

// Exposed Helper Method to Update Like
func (repository *connectionFeedRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateDocumentsInDB(repository.db, ConnectionFeedCollection, filter, update)
}

// Exposed Helper Method to Fetch Likes Count
func (repository *connectionFeedRepository) Count(filter map[string]interface{}) (int64, error) {
	return _countDocumentsInDB(repository.db, ConnectionFeedCollection, filter)
}

// Exposed Helper Method to perform Aggregration on Likes
func (repository *connectionFeedRepository) Aggregate(query []map[string]interface{}) (interface{}, error) {
	return _aggregateDocumentsInDB(repository.db, ConnectionFeedCollection, query)
}

// Structure for Connection Feed Repository
type connectionFeedRepository struct {
	db *mongo.Database
}

// Exposed Method to Create New Connection Feed Repository
func NewConnectionFeedRepository(db *mongo.Database) interfaces.ConnectionFeedRepository {
	return &connectionFeedRepository{
		db: db,
	}
}
