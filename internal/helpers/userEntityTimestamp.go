package helpers

import (
	"context"

	"github.com/LikeMindsCommunity/likeminds-swarm/internal/entities"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/interfaces"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/services/logging"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exposed Helper Method to Create many User Topic Instances
func (helper *UserEntityTimestampHelper) CreateUserEntityTimestampHelper(userId string, communityId int, entityType string, entityIds []primitive.ObjectID, epochTimeStamp int) ([]primitive.ObjectID, error) {

	var userEntityTimestamp []interface{}

	for _, entityId := range entityIds {
		userEntityTimestamp = append(userEntityTimestamp, entities.NewUserEntityTimestamps(userId, communityId, entityType, entityId, epochTimeStamp, 0))
	}

	userEntityTimestamp, err := helper.UserEntityTimestampRepository.CreateMany(userEntityTimestamp)

	return TypecastIdsToObjectIds(userEntityTimestamp), err
}

// Exposed Helper Method to Find User EntityTimestamp
func (helper *UserEntityTimestampHelper) FindUserEntityTimestampHelper(filter map[string]interface{}, filterOptions map[string]interface{},
) ([]entities.UserTopic, error) {

	fOpts := mergeFilterOptions(filterOptions)

	err := convertHexIdsToObjectIds(filter, []string{"_id", "topic_id"})
	if err != nil {
		return nil, err
	}

	cursor, err := helper.UserEntityTimestampRepository.Find(filter, &fOpts)
	if err != nil {
		return nil, err
	}

	// Parse the cursor to entities
	var userEntityTimestamp []entities.UserTopic
	if err = cursor.All(context.TODO(), &userEntityTimestamp); err != nil {
		return nil, err
	}

	return userEntityTimestamp, nil
}

// Exposed Helper Method to Update many User EntityTimestamp
func (helper *UserEntityTimestampHelper) UpdateManyUserEntityTimestampHelper(filter map[string]interface{}, update map[string]interface{},
) error {

	err := convertHexIdsToObjectIds(filter, []string{"_id", "topic_id"})
	if err != nil {
		return err
	}

	err = helper.UserEntityTimestampRepository.UpdateMany(filter, update)

	return err
}

// Exposed Helper Method to return User EntityTimestamp count
func (helper *UserEntityTimestampHelper) CountUserEntityTimestampHelper(filter map[string]interface{}) (int64, error) {

	err := convertHexIdsToObjectIds(filter, []string{"_id", "topic_id"})
	if err != nil {
		return 0, err
	}

	count, err := helper.UserEntityTimestampRepository.Count(filter)

	return count, err
}

// Exposed Helper Method to Delete User EntityTimestamp
func (helper *UserEntityTimestampHelper) DeleteUserEntityTimestampHelper(filter map[string]interface{}) error {

	err := convertHexIdsToObjectIds(filter, []string{"_id", "topic_id"})
	if err != nil {
		return err
	}

	deletedCount, err := helper.UserEntityTimestampRepository.DeleteMany(filter)
	if deletedCount == 0 {
		logging.Error("No User EntityTimestamp deleted")
	}

	return err
}

// Exposed Helper Method to perform Aggregration on Posts
func (helper *UserEntityTimestampHelper) AggregateUserEntityTimestampHelper(query []map[string]interface{}) ([]gin.H, error) {
	results, err := helper.UserEntityTimestampRepository.Aggregate(query)
	if err != nil {
		return nil, err
	}

	var userEntityTimestampList []gin.H

	if err = results.All(context.TODO(), &userEntityTimestampList); err != nil {
		return userEntityTimestampList, err
	}

	return userEntityTimestampList, nil
}

type UserEntityTimestampHelper struct {
	UserEntityTimestampRepository interfaces.UserEntityTimestampRepository
}

func NewUserEntityTimestampHelper(userEntityTimestampRepository interfaces.UserEntityTimestampRepository) interfaces.UserEntityTimestampHelper {
	return &UserEntityTimestampHelper{
		UserEntityTimestampRepository: userEntityTimestampRepository,
	}
}
