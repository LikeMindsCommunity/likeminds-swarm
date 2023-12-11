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
func (helper *activityHelper) CreateActivityHelper(communityID int, actionBy []string, actionOn string, entityType constants.EntityType,
	entityID primitive.ObjectID, entityOwnerID string, action constants.ActivityAction, cta string, isRead bool, isDeleted bool,
	actionByEntityId string) (interface{}, error) {

	// filter to find existing activity
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

	// If activity exist, update activity
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

	} else { // If activity does not exist, create new activity

		// Add action_by's metadata to activity
		actionByMetadata := map[string]entities.ActionByMetadata{
			actionBy[0]: {
				CreatedAt: time.Now(),
				Id:        actionByEntityId,
			},
		}

		activity := entities.NewActivity(communityID, actionBy, actionOn, entityType, entityID, entityOwnerID, action, cta, isRead, isDeleted,
			actionByMetadata)
		activityID, err := helper.activityRepository.Create(&activity)

		return activityID, err
	}
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
	activityIDString := ""

	if _, ok := activityID.(string); ok {
		activityIDString = activityID.(string)
	} else {
		activityIDString = activityID.(primitive.ObjectID).Hex()
	}

	cacheUserActivityFeedKey := fmt.Sprintf(constants.UserActivityFeedCacheKey, userID)
	helper.cacheHelper.LRem(cacheUserActivityFeedKey, 0, activityIDString)
	helper.cacheHelper.LPush(cacheUserActivityFeedKey, activityIDString, 20)

	cacheActivityKey := fmt.Sprintf(constants.ActivityCacheKey, activityIDString)
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
		"$skip":  0,
		"$limit": 20,
		"$sort": gin.H{
			"updated_at": 1,
		},
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

func (helper *activityHelper) WarmupUniversalFeedCache(CommunityID int) []entities.Post {
	postsDAta := []entities.Post{}
	return postsDAta
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
