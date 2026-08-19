package repositories

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Exposed Helper Method to Create Widget
func (repository *widgetRepository) Create(document interface{}) (interface{}, error) {
	return _createDocumentInDB(repository.db, WidgetCollection, document)
}

// Exposed Helper Method to Find Widget
func (repository *widgetRepository) Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error) {
	return _findDocumentsInDB(repository.db, WidgetCollection, filter, filterOpts)
}

// Exposed Helper Method to Update Widget
func (repository *widgetRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateDocumentsInDB(repository.db, WidgetCollection, filter, update)
}

// Exposed Helper Method to Delete Widget
func (repository *widgetRepository) Delete(filter map[string]interface{}) error {
	return _deleteDocumentInDB(repository.db, WidgetCollection, filter)
}

// Exposed Helper Method to Delete mulitple widgets
func (repository *widgetRepository) DeleteMany(filter map[string]interface{}) (int64, error) {
	deletedCount, err := _deleteManyDocumentsInDB(repository.db, WidgetCollection, filter)
	return deletedCount, err
}

// Exposed Helper Method to Fetch Widget Count
func (repository *widgetRepository) Count(filter map[string]interface{}) (int64, error) {
	return _countDocumentsInDB(repository.db, WidgetCollection, filter)
}

// Structure for Save Repository
type widgetRepository struct {
	db *mongo.Database
}

// Exposed Method to Create New Widget Repository
func NewWidgetRepository(db *mongo.Database) interfaces.WidgetRepository {
	return &widgetRepository{
		db: db,
	}
}
