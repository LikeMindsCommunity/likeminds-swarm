package helpers

import (
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (helper *saveHelper) CreateSaveHelper(entity_type string, entity_id primitive.ObjectID, saved_by string) (interface{}, error) {
	save := entities.NewSave(entity_type, entity_id, saved_by)
	save_id, err := helper.saveRepository.Create(&save)

	return save_id, err
}

func (helper *saveHelper) FindSaveHelper(filter map[string]interface{}) ([]entities.Save, error) {
	results, err := helper.saveRepository.Find(filter)

	return results, err
}

func (helper *saveHelper) UpdateSaveHelper(filter map[string]interface{}, update map[string]interface{}) error {
	err := helper.saveRepository.Update(filter, update)

	return err
}

type saveHelper struct {
	saveRepository interfaces.SaveRepository
}

func NewSaveHelper(saveRepository interfaces.SaveRepository) interfaces.SaveHelper {
	return &saveHelper{
		saveRepository: saveRepository,
	}
}
