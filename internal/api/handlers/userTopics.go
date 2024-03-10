package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func fetchUserTopicsForUserIds(userTopicsHelper interfaces.UserTopicsHelper, userIds []string, communityId int,
) (map[string][]primitive.ObjectID, []primitive.ObjectID, error) {

	topicIds := []primitive.ObjectID{}

	// fetch user topics
	filterData := gin.H{
		"user_id":   gin.H{"$in": userIds},
		"community": communityId,
	}

	userTopics, err := userTopicsHelper.FindUserTopicsHelper(filterData, gin.H{})
	if err != nil {
		return nil, nil, err
	}

	userTopicsMap := map[string][]primitive.ObjectID{}

	for _, userTopic := range userTopics {

		if _, ok := userTopicsMap[userTopic.UserID]; !ok {
			userTopicsMap[userTopic.UserID] = []primitive.ObjectID{}
		}

		userTopicsMap[userTopic.UserID] = append(userTopicsMap[userTopic.UserID], userTopic.TopicID)
		topicIds = append(topicIds, userTopic.TopicID)
	}

	return userTopicsMap, topicIds, nil
}

// Internal method to fetch user topics along with topics and widgets data
func fetchUserTopicsForResponse(handlers *FeedHandlers, userIds []string, communityId int, userId string,
) (map[string][]primitive.ObjectID, map[string]responses.TopicResponse, map[string]requests.WidgetResponse, error) {

	// fetch user topics
	userTopicsMap, topicIds, err := fetchUserTopicsForUserIds(handlers.userTopicsHelper, userIds, communityId)
	if err != nil {
		return nil, nil, nil, err
	}

	// fetch parsed topics
	topicsMap, err := fetchAndParseTopicsForResponse(handlers.topicHelper, topicIds, communityId)
	if err != nil {
		return nil, nil, nil, err
	}

	// fetch widgets
	widgetsMap := getWidgetDataFromPostsAndTopics(handlers, gin.H{"topics": topicsMap}, communityId, userId)

	return userTopicsMap, topicsMap, widgetsMap, nil
}

// Exposed Handler Method to fetch User Topics
func (handlers *FeedHandlers) FetchUsersTopics(c *gin.Context) {

	headers := utils.GetHeaders(c)
	userId := headers[utils.HeadersMemberId]

	// validate community id
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		utils.GeneralAPIValidationError(c, "Invalid community id")
		return
	}

	var futr requests.FetchUserTopicsRequest
	if err := c.ShouldBindJSON(&futr); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	if len(futr.UUIDs) == 0 {
		utils.GeneralAPIValidationError(c, "Please send uuids in request")
		return
	}

	// fetch user topics along with topics and widgets data
	userTopicsMap, topicsMap, widgetsMap, err := fetchUserTopicsForResponse(handlers, futr.UUIDs, communityId, userId)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// final response
	response := gin.H{
		"user_topics": userTopicsMap,
		"topics":      topicsMap,
		"widgets":     widgetsMap,
	}

	utils.GenerateSuccessResponse(c, response)
}

func validateUpdateUserTopicsRequest(topicHelper interfaces.TopicHelper, userTopicsHelper interfaces.UserTopicsHelper,
	uutr requests.UpdateUserTopicsRequest, uuid string, communityId int,
) ([]primitive.ObjectID, []primitive.ObjectID, error) {

	// validate sent topic_ids
	if len(uutr.TopicIds) == 0 {
		return nil, nil, fmt.Errorf("Invalid topic_ids")
	}

	topicsToAdd, topicsToRemove := []primitive.ObjectID{}, []primitive.ObjectID{}

	for topicId, add := range uutr.TopicIds {
		objectId, err := primitive.ObjectIDFromHex(topicId)
		if err != nil {
			return nil, nil, err
		}

		if add {
			topicsToAdd = append(topicsToAdd, objectId)
		} else {
			topicsToRemove = append(topicsToRemove, objectId)
		}
	}

	topicsIds := append(topicsToAdd, topicsToRemove...)

	// fetch enabled topics
	topics, err := fetchTopicsByIDs(topicHelper, topicsIds, communityId, true)
	if err != nil {
		logging.Error("Error fetching topics: ", err)
		return nil, nil, err
	}

	if len(topics) != len(topicsIds) {
		return nil, nil, fmt.Errorf("Invalid topic_ids")
	}

	// validate if user does not already have the topics
	filterData := gin.H{
		"user_id":   uuid,
		"topic_id":  gin.H{"$in": topicsToAdd},
		"community": communityId,
	}

	userTopics, err := userTopicsHelper.FindUserTopicsHelper(filterData, gin.H{})
	if err != nil {
		logging.Error("Error fetching user topics: ", err)
		return nil, nil, err
	}

	if len(userTopics) > 0 {
		return nil, nil, fmt.Errorf("User already has topics: ", topicsToAdd)
	}

	// validate if user has the topics to remove
	filterData = gin.H{
		"user_id":   uuid,
		"topic_id":  gin.H{"$in": topicsToRemove},
		"community": communityId,
	}

	userTopics, err = userTopicsHelper.FindUserTopicsHelper(filterData, gin.H{})
	if err != nil {
		logging.Error("Error fetching user topics: ", err)
		return nil, nil, err
	}

	if len(userTopics) != len(topicsToRemove) {
		return nil, nil, fmt.Errorf("User does not have topics: ", topicsToRemove)
	}

	return topicsToAdd, topicsToRemove, nil
}

// Internal method to add topics for a user
func addTopicsForUser(userTopicsHelper interfaces.UserTopicsHelper, userId string, topicsToAdd []primitive.ObjectID,
	communityId int,
) error {

	// Add user topics
	userTopics := map[string][]primitive.ObjectID{
		userId: topicsToAdd,
	}

	_, err := userTopicsHelper.CreateUsersTopicsHelper(userTopics, communityId)
	if err != nil {
		return err
	}

	// Invalidate Kettle Cache for user topics //TODO: move to constants
	go externalHelpers.InvalidateKettleCache(fmt.Sprintf("%d_%s_user_topics", communityId, userId))

	return nil
}

// Internal method to delete topics for a user
func deleteTopicsForUser(userTopicsHelper interfaces.UserTopicsHelper, userId string, topicsToRemove []primitive.ObjectID,
	communityId int,
) error {

	// remove user topics
	filterData := gin.H{
		"user_id":   userId,
		"topic_id":  gin.H{"$in": topicsToRemove},
		"community": communityId,
	}

	err := userTopicsHelper.DeleteUserTopicsHelper(filterData)
	if err != nil {
		return err
	}

	// Invalidate Kettle Cache for user topics //TODO: move to constants
	go externalHelpers.InvalidateKettleCache(fmt.Sprintf("%d_%s_user_topics", communityId, userId))

	return nil
}

// Exposed Handler Method to update User Topics
func (handlers *FeedHandlers) UpdateUserTopics(c *gin.Context) {

	headers := utils.GetHeaders(c)
	userId := headers[utils.HeadersMemberId]
	uuid := c.Param("user_id")

	// validate community id
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		utils.GeneralAPIValidationError(c, "Invalid community id")
		return
	}

	var uutr requests.UpdateUserTopicsRequest
	if err := c.ShouldBindJSON(&uutr); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validate if user is authorized to perform this action
	if !uutr.UserIsCM && userId != uuid {
		utils.GeneralAPIValidationError(c, "You are not authorized to perform this action")
		return
	}

	// validate request and fetch topics to add and remove
	topicsToAdd, topicsToRemove, err := validateUpdateUserTopicsRequest(handlers.topicHelper, handlers.userTopicsHelper, uutr, uuid, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// add topics for user
	if len(topicsToAdd) > 0 {
		err = addTopicsForUser(handlers.userTopicsHelper, uuid, topicsToAdd, communityId)
		if err != nil {
			utils.GeneralAPIInternalError(c, "Error adding topics for user")
			return
		}
	}

	// remove topics for user
	if len(topicsToRemove) > 0 {
		err = deleteTopicsForUser(handlers.userTopicsHelper, uuid, topicsToRemove, communityId)
		if err != nil {
			utils.GeneralAPIInternalError(c, "Error removing topics for user")
			return
		}
	}

	// return final response
	utils.GenerateSuccessResponse(c, nil)
}
