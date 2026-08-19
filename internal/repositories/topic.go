package repositories

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Exposed Helper Method to Create Topic
func (repository *topicRepository) Create(document interface{}) (interface{}, error) {
	return _createDocumentInDB(repository.db, TopicCollection, document)
}

// Exposed Helper Method to Create Multiple Topic
func (repository *topicRepository) CreateMany(documents []interface{}) ([]interface{}, error) {
	return _createManyDocumentsInDB(repository.db, TopicCollection, documents)
}

// Exposed Helper Method to Find Topics
func (repository *topicRepository) Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error) {
	return _findDocumentsInDB(repository.db, TopicCollection, filter, filterOpts)
}

// Exposed Helper Method to Update Topics
func (repository *topicRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateDocumentsInDB(repository.db, TopicCollection, filter, update)
}

// Exposed Helper Method to Update Many Topics
func (repository *topicRepository) UpdateMany(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateAllDocumentsInDB(repository.db, TopicCollection, filter, update)
}

// Exposed Helper Method to Fetch Topics Count
func (repository *topicRepository) Count(filter map[string]interface{}) (int64, error) {
	return _countDocumentsInDB(repository.db, TopicCollection, filter)
}

// Exposed Helper Method to Delete Many topics
func (repository *topicRepository) DeleteMany(filter map[string]interface{}) (int64, error) {
	return _deleteManyDocumentsInDB(repository.db, TopicCollection, filter)
}

// Structure for Save Repository
type topicRepository struct {
	db *mongo.Database
}

// Exposed Method to Create New Topic Repository
func NewTopicRepository(db *mongo.Database) interfaces.TopicRepository {
	return &topicRepository{
		db: db,
	}
}
