package interfaces

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Interface for Save Repository
type SaveRepository interface {
	Create(document interface{}) (interface{}, error)
	Find(filter map[string]interface{}, filterOpts *options.FindOptions) (*mongo.Cursor, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
}

// Interface for Save Helper
type SaveHelper interface {
	CreateSaveHelper(entityType string, entityId primitive.ObjectID, savedBy string, communityId int) (interface{}, error)
	FindSaveHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Save, error)
	UpdateSaveByIdHelper(activityId primitive.ObjectID, update map[string]interface{}) error
	CountSaveHelper(filter map[string]interface{}) (int64, error)
}
