package helpers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exposed Helper Method to Create Save Instance
func (helper *saveHelper) CreateSaveHelper(entity_type string, entity_id primitive.ObjectID, saved_by string, community_id int) (interface{}, error) {
	save := entities.NewSave(entity_type, entity_id, saved_by, community_id)
	save_id, err := helper.saveRepository.Create(&save)

	return save_id, err
}

// Exposed Helper Method to Find Saves
func (helper *saveHelper) FindSaveHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Save, error) {
	fOpts := mergeFilterOptions(filterOptions)

	err := convertHexIdsToObjectIds(filter, []string{"_id", "entity_id"})
	if err != nil {
		return nil, err
	}

	results, err := helper.saveRepository.Find(filter, &fOpts)

	return results, err
}

// Exposed Helper Method to Update Saves
func (helper *saveHelper) UpdateSaveByIdHelper(activity_id primitive.ObjectID, update map[string]interface{}) error {
	set_data := gin.H{}

	if _, ok := update["$set"]; ok {
		set_data = update["$set"].(gin.H)
	}
	set_data["updated_at"] = time.Now()
	update["$set"] = set_data

	err := helper.saveRepository.Update(gin.H{"_id": activity_id}, update)

	return err
}

// Exposed Helper Method to Fetch Saves Count
func (helper *saveHelper) CountSaveHelper(filter map[string]interface{}) (int64, error) {
	err := convertHexIdsToObjectIds(filter, []string{"_id", "entity_id"})
	if err != nil {
		return 0, err
	}

	count, err := helper.saveRepository.Count(filter)

	return count, err
}

// Structure for Save Helper
type saveHelper struct {
	saveRepository interfaces.SaveRepository
}

// Exposed Method to Create New Save Helper
func NewSaveHelper(saveRepository interfaces.SaveRepository) interfaces.SaveHelper {
	return &saveHelper{
		saveRepository: saveRepository,
	}
}
