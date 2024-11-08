package helpers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exposed Helper Method to Create many User Topic Instances
func (helper *UserTopicsHelper) CreateUsersTopicsHelper(usersTopicIds map[string][]primitive.ObjectID, communityId int,
) ([]primitive.ObjectID, error) {

	var userTopics []interface{}

	for userId, topic := range usersTopicIds {

		for _, topic := range topic {
			userTopics = append(userTopics, entities.NewUserTopic(userId, topic, communityId))
		}
	}

	userTopics, err := helper.UserTopicsRepository.CreateMany(userTopics)

	return TypecastIdsToObjectIds(userTopics), err
}

// Exposed Helper Method to Find User Topics
func (helper *UserTopicsHelper) FindUserTopicsHelper(filter map[string]interface{}, filterOptions map[string]interface{},
) ([]entities.UserTopic, error) {

	fOpts := mergeFilterOptions(filterOptions)

	err := convertHexIdsToObjectIds(filter, []string{"_id", "topic_id"})
	if err != nil {
		return nil, err
	}

	cursor, err := helper.UserTopicsRepository.Find(filter, &fOpts)
	if err != nil {
		return nil, err
	}

	// Parse the cursor to entities
	var userTopics []entities.UserTopic
	if err = cursor.All(context.TODO(), &userTopics); err != nil {
		return nil, err
	}

	return userTopics, nil
}

// Exposed Helper Method to Update many User Topics
func (helper *UserTopicsHelper) UpdateManyUserTopicsHelper(filter map[string]interface{}, update map[string]interface{},
) error {

	err := convertHexIdsToObjectIds(filter, []string{"_id", "topic_id"})
	if err != nil {
		return err
	}

	err = helper.UserTopicsRepository.UpdateMany(filter, update)

	return err
}

// Exposed Helper Method to return User Topics count
func (helper *UserTopicsHelper) CountUserTopicsHelper(filter map[string]interface{}) (int64, error) {

	err := convertHexIdsToObjectIds(filter, []string{"_id", "topic_id"})
	if err != nil {
		return 0, err
	}

	count, err := helper.UserTopicsRepository.Count(filter)

	return count, err
}

// Exposed Helper Method to Delete User Topics
func (helper *UserTopicsHelper) DeleteUserTopicsHelper(filter map[string]interface{}) error {

	err := convertHexIdsToObjectIds(filter, []string{"_id", "topic_id"})
	if err != nil {
		return err
	}

	deletedCount, err := helper.UserTopicsRepository.DeleteMany(filter)
	if deletedCount == 0 {
		logging.Error("No User Topics deleted")
	}

	return err
}

// Exposed Helper Method to perform Aggregration on Posts
func (helper *UserTopicsHelper) AggregateUserTopicsHelper(query []map[string]interface{}) ([]gin.H, error) {
	results, err := helper.UserTopicsRepository.Aggregate(query)
	if err != nil {
		return nil, err
	}

	var topicIdsList []gin.H

	if err = results.All(context.TODO(), &topicIdsList); err != nil {
		return topicIdsList, err
	}

	return topicIdsList, nil
}

type UserTopicsHelper struct {
	UserTopicsRepository interfaces.UserTopicsRepository
}

func NewUserTopicsHelper(userTopicsRepository interfaces.UserTopicsRepository) interfaces.UserTopicsHelper {
	return &UserTopicsHelper{
		UserTopicsRepository: userTopicsRepository,
	}
}
