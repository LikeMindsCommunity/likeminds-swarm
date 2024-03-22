package repositories

import (
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Exposed Repository Method to Create Comment
func (repository *commentRepository) Create(document interface{}) (interface{}, error) {
	return _createDocumentInDB(repository.db, CommentCollection, document)
}

// Exposed Repository Method to Find Comment
func (repository *commentRepository) Find(filter map[string]interface{}, filterOptions *options.FindOptions) (*mongo.Cursor, error) {
	return _findDocumentsInDB(repository.db, CommentCollection, filter, filterOptions)
}

// Exposed Repository Method to Update Comment
func (repository *commentRepository) Update(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateDocumentsInDB(repository.db, CommentCollection, filter, update)
}

// Exposed Repository Method to Update Many Comments
func (repository *commentRepository) UpdateMany(filter map[string]interface{}, update map[string]interface{}) error {
	return _updateAllDocumentsInDB(repository.db, CommentCollection, filter, update)
}

// Exposed Repository Method to Find Comment Count
func (repository *commentRepository) Count(filter map[string]interface{}) (int64, error) {
	return _countDocumentsInDB(repository.db, CommentCollection, filter)
}

// Exposed Helper Method to perform Aggregration on Comment
func (repository *commentRepository) Aggregate(query []map[string]interface{}) (*mongo.Cursor, error) {
	return _aggregateDocumentsInDBReturnCursor(repository.db, CommentCollection, query)
}

// Structure for Comment Repository
type commentRepository struct {
	db *mongo.Database
}

// Exposed Method to Create New Comment Repository
func NewCommentRepository(db *mongo.Database) interfaces.CommentRepository {
	return &commentRepository{
		db: db,
	}
}
