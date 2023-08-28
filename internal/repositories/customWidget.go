package repositories

import (
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Exposed Helper Method to Create CustomWidget
func (repository *customWidgetRepository) Create(document interface{}) (interface{}, error) {
	return _createDocumentInDB(repository.db, CustomWidgetCollection, document)
}

// Exposed Helper Method to Find Topics
func (repository *customWidgetRepository) Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error) {
	return _findDocumentsInDB(repository.db, CustomWidgetCollection, filter, filterOpts)
}

// Exposed Helper Method to Update Topics
func (repository *customWidgetRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateDocumentsInDB(repository.db, CustomWidgetCollection, filter, update)
}

// Exposed Helper Method to Fetch Topics Count
func (repository *customWidgetRepository) Count(filter map[string]interface{}) (int64, error) {
	return _countDocumentsInDB(repository.db, CustomWidgetCollection, filter)
}

// Structure for Save Repository
type customWidgetRepository struct {
	db *mongo.Database
}

// Exposed Method to Create New CustomWidget Repository
func NewCustomWidgetRepository(db *mongo.Database) interfaces.CustomWidgetRepository {
	return &customWidgetRepository{
		db: db,
	}
}
