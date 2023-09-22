package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
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
				"action_by":  updatedActionBy,
				"is_deleted": false,
			},
		}
		helper.UpdateActivityByIDHelper(existingActivity[0].ID, updateData, false, true)

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
func (helper *activityHelper) UpdateActivityByIDHelper(activityID primitive.ObjectID, update map[string]interface{}, shouldNotUpdateTimestamp bool, shouldPushActivityToCache bool) error {
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

	if shouldPushActivityToCache {
		helper.PushActivitytoCache(activityID.Hex())
	} else {
		helper.UpdateActivityInCache(activityID.Hex())
	}

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
		"$set": gin.H{
			"is_deleted": true,
		},
	}

	err = helper.activityRepository.UpdateAll(filter, delete)

	return err
}

// UpdateActivityInCache | update activity in cache storage
func (helper *activityHelper) UpdateActivityInCache(activityID string) {
	cacheActivityKey := fmt.Sprintf(constants.ActivityCacheKey, activityID)
	cacheActivityString := helper.cacheHelper.Get(cacheActivityKey)

	if cacheActivityString == nil {
		return
	}
	activityFilter := gin.H{
		"_id": activityID,
	}
	activity, err := helper.FindActivityHelper(activityFilter, gin.H{})
	if err != nil {
		return
	}
	activtyBytes, err := json.Marshal(activity[0])
	if err != nil {
		return
	}
	activityString := string(activtyBytes)

	helper.cacheHelper.Set(cacheActivityKey, activityString, 0)
}

func (helper *activityHelper) PushActivitytoCache(activityID interface{}) {
	activityFilter := gin.H{
		"_id": activityID,
	}
	activity, err := helper.FindActivityHelper(activityFilter, gin.H{})
	if err != nil {
		return
	}

	userID := activity[0].ActionOn
	activtyBytes, err := json.Marshal(activity[0])
	if err != nil {
		return
	}
	activityString := string(activtyBytes)

	cacheUserActivityFeedKey := fmt.Sprintf(constants.UserActivityFeedCacheKey, userID)
	helper.cacheHelper.LRem(cacheUserActivityFeedKey, 0, activityID.(primitive.ObjectID).Hex())
	helper.cacheHelper.LPush(cacheUserActivityFeedKey, activityID.(primitive.ObjectID).Hex(), 20)

	cacheActivityKey := fmt.Sprintf(constants.ActivityCacheKey, activityID.(primitive.ObjectID).Hex())
	helper.cacheHelper.Set(cacheActivityKey, activityString, 0)
}

// WarmupUserActivityFeedCache | push user activity feed first page to cache
func (helper *activityHelper) WarmupUserActivityFeedCache(communityID int, userID string) []entities.Activity {
	helper.deleteUserActivityFeedCacheData(userID)

	userActivities := []entities.Activity{}

	// activity filter data
	activityFilterData := gin.H{
		"action_on":    userID,
		"community_id": communityID,
		"is_deleted":   false,
	}

	activityFilterOptions := gin.H{
		"": "",
	}

	// fetch activity using helper method
	activityResults, err := helper.FindActivityHelper(activityFilterData, activityFilterOptions)
	if err != nil {
		return userActivities
	}

	helper.createUserActivityFeedCacheData(userID, activityResults)

	return userActivities
}

func (helper *activityHelper) createUserActivityFeedCacheData(userID string, activities []entities.Activity) {
	userActivityIDs := [](string){}

	for _, activity := range activities {

		cacheActivityKey := fmt.Sprintf(constants.ActivityCacheKey, activity.ID.Hex())
		activityBytes, err := json.Marshal(activity)
		if err != nil {
			return
		}
		activityString := string(activityBytes)
		helper.cacheHelper.Set(cacheActivityKey, activityString, 0)

		userActivityIDs = append(userActivityIDs, activity.ID.Hex())
	}

	cacheUserActivityFeedKey := fmt.Sprintf(constants.UserActivityFeedCacheKey, userID)
	helper.cacheHelper.Set(cacheUserActivityFeedKey, userActivityIDs, 0)
}

func (helper *activityHelper) deleteUserActivityFeedCacheData(userID string) {
	userActivityFeedKey := fmt.Sprintf(constants.UserActivityFeedCacheKey, userID)

	cacheUserActivityIDsString := helper.cacheHelper.Get(userActivityFeedKey)
	cacheUserActivityIDs := [](string){cacheUserActivityIDsString.Val()}

	cacheActivityKeys := [](string){}
	for _, cacheUserActivityID := range cacheUserActivityIDs {
		cacheActivityKey := fmt.Sprintf(constants.ActivityCacheKey, cacheUserActivityID)
		cacheActivityKeys = append(cacheActivityKeys, cacheActivityKey)
	}

	helper.cacheHelper.DelMultiple(cacheActivityKeys)
	helper.cacheHelper.Del(userActivityFeedKey)
}

// Structure for Activity Helper
type activityHelper struct {
	activityRepository interfaces.ActivityRepository
	cacheHelper        cache.Helper
}

// NewActivityHelper | method to Create New Activity Helper
func NewActivityHelper(activityRepository interfaces.ActivityRepository, cacheHelper cache.Helper) interfaces.ActivityHelper {
	return &activityHelper{
		activityRepository: activityRepository,
		cacheHelper:        cacheHelper,
	}
}
