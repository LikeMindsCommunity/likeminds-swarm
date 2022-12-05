package interfaces

import (
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SaveRepository interface {
	Create(like *entities.Save) (interface{}, error)
	Find(filter map[string]interface{}) ([]entities.Save, error)
	Update(filter map[string]interface{}, update map[string]interface{}) error
}

type SaveHelper interface {
	CreateSaveHelper(entity_type string, entity_id primitive.ObjectID, saved_by string) (interface{}, error)
	FindSaveHelper(filter map[string]interface{}) ([]entities.Save, error)
	UpdateSaveHelper(filter map[string]interface{}, update map[string]interface{}) error
}
