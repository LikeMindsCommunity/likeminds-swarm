package repositories

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Exposed Helper Method to Create Post
func (repository *postRepository) Create(document interface{}) (interface{}, error) {
	return _createDocumentInDB(repository.db, PostCollection, document)
}

// Exposed Helper Method to Find Post
func (repository *postRepository) Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error) {
	return _findDocumentsInDB(repository.db, PostCollection, filter, filterOpts)
}

// Exposed Helper Method to Update Post
func (repository *postRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateDocumentsInDB(repository.db, PostCollection, filter, update)
}

// Exposed Helper Method to Fetch Post Count
func (repository *postRepository) Count(filter map[string]interface{}) (int64, error) {
	return _countDocumentsInDB(repository.db, PostCollection, filter)
}

// Exposed Helper Method to perform Aggregration on Post
func (repository *postRepository) Aggregate(query []map[string]interface{}) ([]gin.H, error) {
	return _aggregateDocumentsInDB(repository.db, PostCollection, query)
}

// Structure for Post Repository
type postRepository struct {
	db *mongo.Database
}

// Exposed Method to Create New Post Repository
func NewPostRepository(db *mongo.Database) interfaces.PostRepository {
	return &postRepository{
		db: db,
	}
}
