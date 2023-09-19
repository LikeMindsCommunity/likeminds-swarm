package helpers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exposed Helper Method to Create Widget Instance
func (helper *widgetHelper) CreateWidgetHelper(createdByLM bool, parentEntityID string, parentEntityType string,
	metaData map[string]interface{}, lmMeta map[string]interface{}, community_id int) (interface{}, error) {
	// Create a new Widget Document
	widget := entities.NewWidget(createdByLM, parentEntityID, parentEntityType, metaData, lmMeta, community_id)

	// Insert the document in the collection
	widgetId, err := helper.widgetRepository.Create(widget)

	return widgetId, err
}

// Exposed Helper Method to Find Widgets
func (helper *widgetHelper) FindWidgetHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Widget,
	error) {
	fOpts := mergeFilterOptions(filterOptions)

	// Parse the object Ids
	err := convertHexIdsToObjectIds(filter, []string{"_id"})
	if err != nil {
		return nil, err
	}

	// Find the document in the collection
	cursor, err := helper.widgetRepository.Find(filter, &fOpts)
	if err != nil {
		return nil, err
	}

	// Parse the results from fetched documents
	var results []entities.Widget
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, err
}

// Exposed Helper Method to Update Widget
func (helper *widgetHelper) UpdateWidgetByIdHelper(widgetId primitive.ObjectID, update map[string]interface{}) error {
	setData := gin.H{}

	// Create set filter
	if _, ok := update["$set"]; ok {
		setData = update["$set"].(gin.H)
	}
	setData["updated_at"] = time.Now()
	update["$set"] = setData

	// Update the document in the collection
	err := helper.widgetRepository.Update(gin.H{"_id": widgetId}, update)

	return err
}

// Exposed Helper Method to Fetch Widget Count
func (helper *widgetHelper) CountWidgetHelper(filter map[string]interface{}) (int64, error) {
	// Parse the object IDs
	err := convertHexIdsToObjectIds(filter, []string{"_id"})
	if err != nil {
		return 0, err
	}

	// Get count of documents from the collection
	count, err := helper.widgetRepository.Count(filter)

	return count, err
}

// Structure for CustomWidget Helper
type widgetHelper struct {
	widgetRepository interfaces.WidgetRepository
}

// Exposed Method to Create New Widget Helper
func NewWidgetHelper(widgetRepository interfaces.WidgetRepository) interfaces.WidgetHelper {
	return &widgetHelper{
		widgetRepository: widgetRepository,
	}
}
