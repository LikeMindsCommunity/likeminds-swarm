package repositories

import (
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Exposed Helper Method to Create Save
func (repository *saveRepository) Create(document interface{}) (interface{}, error) {
	return _createDocumentInDB(repository.db, SaveCollection, document)
}

// Exposed Helper Method to Find Saves
func (repository *saveRepository) Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error) {
	return _findDocumentsInDB(repository.db, SaveCollection, filter, filterOpts)
}

// Exposed Helper Method to Update Saves
func (repository *saveRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateDocumentsInDB(repository.db, SaveCollection, filter, update)
}

// Exposed Helper Method to Fetch Saves Count
func (repository *saveRepository) Count(filter map[string]interface{}) (int64, error) {
	return _countDocumentsInDB(repository.db, SaveCollection, filter)
}

// Structure for Save Repository
type saveRepository struct {
	db *mongo.Database
}

// Exposed Method to Create New Save Repository
func NewSaveRepository(db *mongo.Database) interfaces.SaveRepository {
	return &saveRepository{
		db: db,
	}
}
