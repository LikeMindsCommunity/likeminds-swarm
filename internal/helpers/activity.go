package helpers

import (
	"fmt"

	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func fetchActivityCtaForAction(action string, cta_data map[string]interface{}) string {
	var cta string = ""

	switch action {
	case constants.LikeAction:
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

	case constants.CreatePostPermitAddedAction:
		cta = utils.CreatePostRoute

	case constants.CreateCommentPermissionAddedAction:
		cta = utils.HomeFeedRoute
	}

	return cta
}

func (helper *activityHelper) CreateActivityHelper(action_by string, action_on string, api_key string, entity_type string,
	entity_id primitive.ObjectID, action string, cta_data map[string]interface{}) (interface{}, error) {
	cta := fetchActivityCtaForAction(action, cta_data)
	activity := entities.NewActivity(action_by, action_on, api_key, entity_type, entity_id, action, cta)
	activity_id, err := helper.activityRepository.Create(&activity)

	return activity_id, err
}

func (helper *activityHelper) FindActivityHelper(filter map[string]interface{}) ([]entities.Activity, error) {
	results, err := helper.activityRepository.Find(filter)

	return results, err
}

func (helper *activityHelper) UpdateActivityHelper(filter map[string]interface{}, update map[string]interface{}) error {
	err := helper.activityRepository.Update(filter, update)

	return err
}

type activityHelper struct {
	activityRepository interfaces.ActivityRepository
}

func NewActivityHelper(activityRepository interfaces.ActivityRepository) interfaces.ActivityHelper {
	return &activityHelper{
		activityRepository: activityRepository,
	}
}
