package interfaces

import (
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SaveRepository interface {
	Create(like *entities.Save) (interface{}, error)
	Find(filter map[string]interface{}, filterOpts *options.FindOptions) ([]entities.Save, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
	Count(filter map[string]interface{}) (int64, error)
}

type SaveHelper interface {
	CreateSaveHelper(entity_type string, entity_id primitive.ObjectID, saved_by string) (interface{}, error)
	FindSaveHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Save, error)
	UpdateSaveByIdHelper(activity_id primitive.ObjectID, update map[string]interface{}) error
	CountSaveHelper(filter map[string]interface{}) (int64, error)
}
