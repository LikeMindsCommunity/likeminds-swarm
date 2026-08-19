package repositories

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Exposed Helper Method to Create Like
func (repository *likeRepository) Create(document interface{}) (interface{}, error) {
	return _createDocumentInDB(repository.db, LikeCollection, document)
}

// Exposed Helper Method to Find Like
func (repository *likeRepository) Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error) {
	return _findDocumentsInDB(repository.db, LikeCollection, filter, filterOpts)
}

// Exposed Helper Method to Update Like
func (repository *likeRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateDocumentsInDB(repository.db, LikeCollection, filter, update)
}

// Exposed Helper Method to Fetch Likes Count
func (repository *likeRepository) Count(filter map[string]interface{}) (int64, error) {
	return _countDocumentsInDB(repository.db, LikeCollection, filter)
}

// Exposed Helper Method to perform Aggregration on Likes
func (repository *likeRepository) Aggregate(query []map[string]interface{}) (interface{}, error) {
	return _aggregateDocumentsInDB(repository.db, LikeCollection, query)
}

// Structure for Like Repository
type likeRepository struct {
	db *mongo.Database
}

// Exposed Method to Create New Like Repository
func NewLikeRepository(db *mongo.Database) interfaces.LikeRepository {
	return &likeRepository{
		db: db,
	}
}
