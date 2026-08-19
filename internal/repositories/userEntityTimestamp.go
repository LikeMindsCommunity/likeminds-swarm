package repositories

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (repository *userEntityTimestampRepository) Create(document interface{}) (interface{}, error) {
	return _createDocumentInDB(repository.db, UserEntityTimestampCollection, document)
}

func (repository *userEntityTimestampRepository) CreateMany(documents []interface{}) ([]interface{}, error) {
	return _createManyDocumentsInDB(repository.db, UserEntityTimestampCollection, documents)
}

func (repository *userEntityTimestampRepository) Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error) {
	return _findDocumentsInDB(repository.db, UserEntityTimestampCollection, filter, filterOpts)
}

func (repository *userEntityTimestampRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateDocumentsInDB(repository.db, UserEntityTimestampCollection, filter, update)
}

func (repository *userEntityTimestampRepository) UpdateMany(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateAllDocumentsInDB(repository.db, UserEntityTimestampCollection, filter, update)
}

func (repository *userEntityTimestampRepository) Count(filter map[string]interface{}) (int64, error) {
	return _countDocumentsInDB(repository.db, UserEntityTimestampCollection, filter)
}

func (repository *userEntityTimestampRepository) DeleteMany(filter map[string]interface{}) (int64, error) {
	return _deleteManyDocumentsInDB(repository.db, UserEntityTimestampCollection, filter)
}

// Exposed Helper Method to perform Aggregation on User Topics
func (repository *userEntityTimestampRepository) Aggregate(query []map[string]interface{}) (*mongo.Cursor, error) {
	return _aggregateDocumentsInDBReturnCursor(repository.db, UserEntityTimestampCollection, query)
}

type userEntityTimestampRepository struct {
	db *mongo.Database
}

func NewUserEntityTimestampRepository(db *mongo.Database) interfaces.UserEntityTimestampRepository {
	return &userEntityTimestampRepository{
		db: db,
	}
}
