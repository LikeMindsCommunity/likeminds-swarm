package repositories

import (
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (repository *userTopicsRepository) Create(document interface{}) (interface{}, error) {
	return _createDocumentInDB(repository.db, UserTopicsCollection, document)
}

func (repository *userTopicsRepository) CreateMany(documents []interface{}) ([]interface{}, error) {
	return _createManyDocumentsInDB(repository.db, UserTopicsCollection, documents)
}

func (repository *userTopicsRepository) Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error) {
	return _findDocumentsInDB(repository.db, UserTopicsCollection, filter, filterOpts)
}

func (repository *userTopicsRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateDocumentsInDB(repository.db, UserTopicsCollection, filter, update)
}

func (repository *userTopicsRepository) UpdateMany(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateAllDocumentsInDB(repository.db, UserTopicsCollection, filter, update)
}

func (repository *userTopicsRepository) Count(filter map[string]interface{}) (int64, error) {
	return _countDocumentsInDB(repository.db, UserTopicsCollection, filter)
}

func (repository *userTopicsRepository) DeleteMany(filter map[string]interface{}) (int64, error) {
	return _deleteManyDocumentsInDB(repository.db, UserTopicsCollection, filter)
}

// Exposed Helper Method to perform Aggregation on User Topics
func (repository *userTopicsRepository) Aggregate(query []map[string]interface{}) (*mongo.Cursor, error) {
	return _aggregateDocumentsInDBReturnCursor(repository.db, UserTopicsCollection, query)
}

type userTopicsRepository struct {
	db *mongo.Database
}

func NewUserTopicsRepository(db *mongo.Database) interfaces.UserTopicsRepository {
	return &userTopicsRepository{
		db: db,
	}
}
