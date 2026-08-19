package repositories

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/interfaces"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Exposed Helper Method to Create Post
func (repository *pendingPostRepository) Create(document interface{}) (interface{}, error) {
	return _createDocumentInDB(repository.db, PendingPostCollection, document)
}

// Exposed Helper Method to Find Post
func (repository *pendingPostRepository) Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error) {
	return _findDocumentsInDB(repository.db, PendingPostCollection, filter, filterOpts)
}

// Exposed Helper Method to Update Post
func (repository *pendingPostRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateDocumentsInDB(repository.db, PendingPostCollection, filter, update)
}

// Exposed Helper Method to Fetch Post Count
func (repository *pendingPostRepository) Count(filter map[string]interface{}) (int64, error) {
	return _countDocumentsInDB(repository.db, PendingPostCollection, filter)
}

// Exposed Helper Method to perform Aggregration on Post
func (repository *pendingPostRepository) Aggregate(query []map[string]interface{}) ([]gin.H, error) {
	return _aggregateDocumentsInDB(repository.db, PendingPostCollection, query)
}

// Structure for Post Repository
type pendingPostRepository struct {
	db *mongo.Database
}

// Exposed Method to Create New Post Repository
func NewPendingPostRepository(db *mongo.Database) interfaces.PendingPostRepository {
	return &pendingPostRepository{
		db: db,
	}
}
