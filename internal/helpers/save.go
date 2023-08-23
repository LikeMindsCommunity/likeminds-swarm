package helpers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exposed Helper Method to Create Save Instance
func (helper *saveHelper) CreateSaveHelper(entityType string, entityId primitive.ObjectID, savedBy string, communityId int) (interface{}, error) {
	save := entities.NewSave(entityType, entityId, savedBy, communityId)
	saveId, err := helper.saveRepository.Create(&save)

	return saveId, err
}

// Exposed Helper Method to Find Saves
func (helper *saveHelper) FindSaveHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Save, error) {
	fOpts := mergeFilterOptions(filterOptions)

	err := convertHexIdsToObjectIds(filter, []string{"_id", "entity_id"})
	if err != nil {
		return nil, err
	}

	// Find the document in the collection
	cursor, err := helper.saveRepository.Find(filter, &fOpts)
	if err != nil {
		return nil, err
	}

	// Parse the results from fetched documents
	var results []entities.Save
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, err
}

// Exposed Helper Method to Update Saves
func (helper *saveHelper) UpdateSaveByIdHelper(activityId primitive.ObjectID, update map[string]interface{}) error {
	setData := gin.H{}

	if _, ok := update["$set"]; ok {
		setData = update["$set"].(gin.H)
	}
	setData["updated_at"] = time.Now()
	update["$set"] = setData

	err := helper.saveRepository.Update(gin.H{"_id": activityId}, update)

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
