package helpers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateActivityHelper | create activity entry
func (helper *activityHelper) CreateActivityHelper(communityID int, actionBy []string, actionOn string, entityType constants.EntityType, entityID primitive.ObjectID, entityOwnerID string, action constants.ActivityAction, cta string, isRead bool, isDeleted bool) (interface{}, error) {
	activityFilter := gin.H{
		"community_id": communityID,
		"action_on":    actionOn,
		"entity_id":    entityID,
		"action":       action,
	}

	existingActivity, err := helper.FindActivityHelper(activityFilter, gin.H{})
	if err != nil {
		return nil, err
	}

	if len(existingActivity) > 0 {

		//remove user from action_by list, if exist from previous actions
		existingActionBy := utils.RemoveAllOccurenceStringList(existingActivity[0].ActionBy, actionBy[0])

		updatedActionBy := append(existingActionBy, actionBy...)
		updateData := gin.H{
			"$set": gin.H{
				"action_by": updatedActionBy,
			},
		}
		helper.UpdateActivityByIDHelper(existingActivity[0].ID, updateData, false)

		return existingActivity[0].ID, nil
	}

	activity := entities.NewActivity(communityID, actionBy, actionOn, entityType, entityID, entityOwnerID, action, cta, isRead, isDeleted)
	activityID, err := helper.activityRepository.Create(&activity)

	return activityID, err
}

// Exposed Helper Method to Find Activity
func (helper *activityHelper) FindActivityHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Activity, error) {
	fOpts := mergeFilterOptions(filterOptions)

	err := convertHexIdsToObjectIds(filter, []string{"_id", "entity_id"})
	if err != nil {
		return nil, err
	}

	// Find the document in the collection
	cursor, err := helper.activityRepository.Find(filter, &fOpts)
	if err != nil {
		return nil, err
	}

	// Parse the results from fetched documents
	var results []entities.Activity
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, err
}

// Exposed Helper Method to Update Activity by activity_id
func (helper *activityHelper) UpdateActivityByIDHelper(activityID primitive.ObjectID, update map[string]interface{}, shouldNotUpdateTimestamp bool) error {
	var setData gin.H

	if _, ok := update["$set"]; ok {
		setData = update["$set"].(gin.H)
	}

	setData["updated_at"] = time.Now()

	if shouldNotUpdateTimestamp {
		delete(setData, "updated_at")
	}

	update["$set"] = setData

	err := helper.activityRepository.Update(gin.H{"_id": activityID}, update)

	return err
}

// Exposed Helper Method to Count Activity
func (helper *activityHelper) CountActivityHelper(filter map[string]interface{}) (int64, error) {
	err := convertHexIdsToObjectIds(filter, []string{"_id", "activity_id"})
	if err != nil {
		return 0, err
	}

	count, err := helper.activityRepository.Count(filter)

	return count, err
}

// DeleteActivityHelper | delete activity from repository with filter
func (helper *activityHelper) DeleteActivityHelper(filter map[string]interface{}) error {
	err := convertHexIdsToObjectIds(filter, []string{"_id", "activity_id"})
	if err != nil {
		return err
	}

	delete := gin.H{
		"is_deleted": true,
	}

	err = helper.activityRepository.Update(filter, delete)

	return err
}

// Structure for Activity Helper
type activityHelper struct {
	activityRepository interfaces.ActivityRepository
}

// NewActivityHelper| method to Create New Activity Helper
func NewActivityHelper(activityRepository interfaces.ActivityRepository) interfaces.ActivityHelper {
	return &activityHelper{
		activityRepository: activityRepository,
	}
}
