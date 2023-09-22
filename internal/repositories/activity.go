package repositories

import (
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Exposed Repository Method to Create Activity
func (repository *activityRepository) Create(document interface{}) (interface{}, error) {
	return _createDocumentInDB(repository.db, ActivityCollection, document)
}

// Exposed Repository Method to Find Activity
func (repository *activityRepository) Find(filter map[string]interface{}, filterOptions *options.FindOptions) (*mongo.Cursor, error) {
	return _findDocumentsInDB(repository.db, ActivityCollection, filter, filterOptions)
}

// Exposed Repository Method to Update Activity
func (repository *activityRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateDocumentsInDB(repository.db, ActivityCollection, filter, update)
}

// Exposed Repository Method to Update All filter Activity
func (repository *activityRepository) UpdateAll(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateAllDocumentsInDB(repository.db, ActivityCollection, filter, update)
}

// Count | returns count of activity with filter
func (repository *activityRepository) Count(filter map[string]interface{}) (int64, error) {
	return _countDocumentsInDB(repository.db, ActivityCollection, filter)
}

// Structure for Activity Repository
type activityRepository struct {
	db *mongo.Database
}

// Exposed Method to Create New Activity Repository
func NewActivityRepository(db *mongo.Database) interfaces.ActivityRepository {
	return &activityRepository{
		db: db,
	}
}
