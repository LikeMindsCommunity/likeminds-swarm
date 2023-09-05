package helpers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exposed Helper Method to Create CustomWidget Instance
func (helper *customWidgetHelper) CreateCustomWidgetHelper(createdByLM bool, parentEntityID string, parentEntityType string,
	metaData map[string]interface{}, community_id int) (interface{}, error) {
	// Create a new CustomWidget Document
	customWidget := entities.NewCustomWidget(createdByLM, parentEntityID, parentEntityType, metaData, community_id)

	// Insert the document in the collection
	customWidgetId, err := helper.customWidgetRepository.Create(customWidget)

	return customWidgetId, err
}

// Exposed Helper Method to Find CustomWidgets
func (helper *customWidgetHelper) FindCustomWidgetHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.CustomWidget,
	error) {
	fOpts := mergeFilterOptions(filterOptions)

	// Parse the object Ids
	err := convertHexIdsToObjectIds(filter, []string{"_id"})
	if err != nil {
		return nil, err
	}

	// Find the document in the collection
	cursor, err := helper.customWidgetRepository.Find(filter, &fOpts)
	if err != nil {
		return nil, err
	}

	// Parse the results from fetched documents
	var results []entities.CustomWidget
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, err
}

// Exposed Helper Method to Update CustomWidget
func (helper *customWidgetHelper) UpdateCustomWidgetByIdHelper(customWidgetId primitive.ObjectID, update map[string]interface{}) error {
	setData := gin.H{}

	// Create set filter
	if _, ok := update["$set"]; ok {
		setData = update["$set"].(gin.H)
	}
	setData["updated_at"] = time.Now()
	update["$set"] = setData

	// Update the document in the collection
	err := helper.customWidgetRepository.Update(gin.H{"_id": customWidgetId}, update)

	return err
}

// Exposed Helper Method to Fetch CustomWidget Count
func (helper *customWidgetHelper) CountCustomWidgetHelper(filter map[string]interface{}) (int64, error) {
	// Parse the object IDs
	err := convertHexIdsToObjectIds(filter, []string{"_id"})
	if err != nil {
		return 0, err
	}

	// Get count of documents from the collection
	count, err := helper.customWidgetRepository.Count(filter)

	return count, err
}

// Structure for CustomWidget Helper
type customWidgetHelper struct {
	customWidgetRepository interfaces.CustomWidgetRepository
}

// Exposed Method to Create New CustomWidget Helper
func NewCustomWidgetHelper(customWidgetRepository interfaces.CustomWidgetRepository) interfaces.CustomWidgetHelper {
	return &customWidgetHelper{
		customWidgetRepository: customWidgetRepository,
	}
}
