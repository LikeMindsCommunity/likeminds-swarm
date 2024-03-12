package repositories

import (
	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Exposed Helper Method to Create Post Topics
func (repository *postTopicsRepository) Create(document interface{}) (interface{}, error) {
	return _createDocumentInDB(repository.db, PostTopicsCollection, document)
}

// Exposed Helper Method to Find Post Topics
func (repository *postTopicsRepository) Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error) {
	return _findDocumentsInDB(repository.db, PostTopicsCollection, filter, filterOpts)
}

// Exposed Helper Method to Update Post Topics
func (repository *postTopicsRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateDocumentsInDB(repository.db, PostTopicsCollection, filter, update)
}

// Exposed Helper Method to Create or Update Many Post Topics
func (repository *postTopicsRepository) CreateorUpdateMany(filterWithUpdateList [][]gin.H) error {
	return _updateAllDocumentsInDBInBulk(repository.db, PostTopicsCollection, filterWithUpdateList)
}

// Exposed Helper Method to Fetch Post Topics Count
func (repository *postTopicsRepository) Count(filter map[string]interface{}) (int64, error) {
	return _countDocumentsInDB(repository.db, PostTopicsCollection, filter)
}

// Exposed Helper Method to Delete Many Post Topics
func (repository *postTopicsRepository) DeleteMany(filter map[string]interface{}) (int64, error) {
	return _deleteManyDocumentsInDB(repository.db, PostTopicsCollection, filter)
}

// Exposed Helper Method to perform Aggregation on Post Topics
func (repository *postTopicsRepository) Aggregate(query []map[string]interface{}) (*mongo.Cursor, error) {
	return _aggregateDocumentsInDBReturnCursor(repository.db, PostTopicsCollection, query)
}

// Structure for Save Repository
type postTopicsRepository struct {
	db *mongo.Database
}

// Exposed Method to Create New Post Topics Repository
func NewPostTopicsRepository(db *mongo.Database) interfaces.PostTopicsRepository {
	return &postTopicsRepository{
		db: db,
	}
}
