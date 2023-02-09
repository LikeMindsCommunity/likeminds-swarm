package helpers

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to parse CTA Data
func parseCTAData(cta_data map[string]interface{}) string {
	var cta string = ""

	if entity_type, ok := cta_data["entity_type"]; ok {
		if post_id, ok := cta_data["post_id"]; ok {
			switch entity_type {
			case constants.PostEntityType:
				cta = fmt.Sprintf(utils.PostDetailRoute, post_id)

			case constants.CommentEntityType:
				if comment_id, ok := cta_data["comment_id"]; ok {
					cta = fmt.Sprintf(utils.CommentDetailRoute, post_id, comment_id)
				}
			}
		}
	}

	return cta
}

// Internal Method to parse CTA for corresponding action
func fetchActivityCtaForAction(action string, cta_data map[string]interface{}) string {
	var cta string = ""

	switch action {
	case constants.LikeAction, constants.AlsoCommentAction, constants.CommentAction, constants.TagAction, constants.SaveAction:
		cta = parseCTAData(cta_data)

	case constants.CreatePostPermitAddedAction:
		cta = utils.CreatePostRoute

	case constants.CreateCommentPermissionAddedAction:
		cta = utils.HomeFeedRoute
	}

	return cta
}

// Exposed Helper Method to Create Activity
func (helper *activityHelper) CreateActivityHelper(action_by string, action_on []string, community_id int, entity_type string,
	entity_id primitive.ObjectID, action string, cta_data map[string]interface{}) (interface{}, error) {
	cta := fetchActivityCtaForAction(action, cta_data)
	activity := entities.NewActivity(action_by, action_on, community_id, entity_type, entity_id, action, cta)
	activity_id, err := helper.activityRepository.Create(&activity)

	return activity_id, err
}

// Exposed Helper Method to Find Activity
func (helper *activityHelper) FindActivityHelper(filter map[string]interface{}, filterOptions map[string]interface{}) ([]entities.Activity, error) {
	fOpts := mergeFilterOptions(filterOptions)

	err := convertHexIdsToObjectIds(filter, []string{"_id", "entity_id"})
	if err != nil {
		return nil, err
	}

	results, err := helper.activityRepository.Find(filter, &fOpts)

	return results, err
}

// Exposed Helper Method to Update Activity by activity_id
func (helper *activityHelper) UpdateActivityByIdHelper(activity_id primitive.ObjectID, update map[string]interface{}) error {
	var set_data gin.H

	if _, ok := update["$set"]; ok {
		set_data = update["$set"].(gin.H)
	}
	set_data["updated_at"] = time.Now()
	update["$set"] = set_data

	err := helper.activityRepository.Update(gin.H{"_id": activity_id}, update)

	return err
}

// Structure for Activity Helper
type activityHelper struct {
	activityRepository interfaces.ActivityRepository
}

// Exposed Method to Create New Activity Helper
func NewActivityHelper(activityRepository interfaces.ActivityRepository) interfaces.ActivityHelper {
	return &activityHelper{
		activityRepository: activityRepository,
	}
}
