package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to parse topic for response
func parseTopicResponse(topic *entities.Topic) requests.TopicResponse {
	var response requests.TopicResponse

	response.ID = topic.ID
	response.Name = topic.Name
	response.IsEnabled = topic.IsEnabled

	return response
}

// Internal Method to fetch topic using topic_id and community_id
func fetchTopicByID(helper interfaces.TopicHelper, topicId string, communityId int) (*entities.Topic, error) {
	// topic filter data
	topicFilterData := gin.H{
		"_id":          topicId,
		"community_id": communityId,
	}

	// fetch topic using helper method
	topicResults, err := helper.FindTopicHelper(topicFilterData, gin.H{})
	if err != nil {
		return nil, err
	}

	// validation of topic_id
	if len(topicResults) == 0 {
		return nil, fmt.Errorf("invalid topic_id sent")
	}

	return &topicResults[0], nil
}

// Internal Method to fetch topics using topic_ids and community_id
func fetchTopicsByIDs(helper interfaces.TopicHelper, topicIds []primitive.ObjectID, communityId int,
	filterEnabled bool) ([]entities.Topic, error) {
	// topic filter data
	topicsFilterData := gin.H{
		"_id": gin.H{
			"$in": topicIds,
		},
		"community_id": communityId,
	}

	if filterEnabled {
		topicsFilterData["is_enabled"] = filterEnabled
	}

	// fetch topic using helper method
	topicResults, err := helper.FindTopicHelper(topicsFilterData, gin.H{})
	if err != nil {
		return nil, err
	}

	return topicResults, nil
}

// Internal Method to fetch topics using community_id
func fetchTopicsByCommunityID(helper interfaces.TopicHelper, communityId int, isEnabled string,
	filterOptions map[string]interface{}) ([]entities.Topic, error) {
	// topic filter data
	topicFilterData := gin.H{
		"community_id": communityId,
	}

	if isEnabled != "" {
		if isEnabled == "true" {
			topicFilterData["is_enabled"] = true
		}
		if isEnabled == "false" {
			topicFilterData["is_enabled"] = false
		}
	}

	// fetch topic using helper method
	topicResults, err := helper.FindTopicHelper(topicFilterData, filterOptions)
	if err != nil {
		return nil, err
	}

	return topicResults, nil
}

// Internal Method to fetch topic data
func fetchTopicByIDResponse(handlers *FeedHandlers, topicId string, communityId int) (interface{}, error) {
	topicData, err := fetchTopicByID(handlers.topicHelper, topicId, communityId)
	if err != nil {
		return nil, err
	}

	topicResponse := parseTopicResponse(topicData)

	return topicResponse, nil
}

// Internal Method to fetch multiple topics data
func fetchTopicsByCommunityIDResponse(handlers *FeedHandlers, communityId int, isEnabled string, filterOptions map[string]interface{}) ([]interface{}, error) {
	topicsData, err := fetchTopicsByCommunityID(handlers.topicHelper, communityId, isEnabled, filterOptions)
	if err != nil {
		return nil, err
	}

	topicsResponse := []interface{}{}

	// Parse all fetched topics Data
	for _, topic := range topicsData {
		topicsResponse = append(topicsResponse, parseTopicResponse(&topic))
	}

	return topicsResponse, nil
}

// Exposed Method to Create a Topic
func (handlers *FeedHandlers) CreateTopic(c *gin.Context) {
	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var createTopicRequest requests.CreateTopicRequest
	if err := c.ShouldBindJSON(&createTopicRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// strip text to check if it is empty
	createTopicRequest.Name = strings.Trim(createTopicRequest.Name, " ")

	if createTopicRequest.Name == "" {
		utils.GeneralAPIValidationError(c, "Can't Create Topic With Empty Name")
		return
	}

	// create topic using the helper method
	topicId, err := handlers.topicHelper.CreateTopicHelper(createTopicRequest.Name, true, communityId)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// fetch topic data using new topic_id
	topicResponse, err := fetchTopicByIDResponse(handlers, topicId.(primitive.ObjectID).Hex(), communityId)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"success": true,
		"topic":   topicResponse,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to Fetch Topics for a Community
func (handlers *FeedHandlers) FetchTopics(c *gin.Context) {
	isEnabled := c.Query("is_enabled")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// filter options
	filterOptions, err := generatePageFilterOptions(c, "name", OrderTypeAscending)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch topics data using new communityId
	topicsResponse, err := fetchTopicsByCommunityIDResponse(handlers, communityId,
		isEnabled, filterOptions)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"success": true,
		"topics":  topicsResponse,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to Edit a Topic
func (handlers *FeedHandlers) EditTopic(c *gin.Context) {
	// fetch url params
	topicId := c.Param("topic_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var editTopicRequest requests.EditTopicRequest
	if err := c.ShouldBindJSON(&editTopicRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch topic using topic_id
	topic, err := fetchTopicByID(handlers.topicHelper, topicId, communityId)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// topic update data
	topic_update_data := gin.H{
		"$set": gin.H{},
	}

	// strip text to check if it is empty
	editTopicRequest.Name = strings.Trim(editTopicRequest.Name, " ")

	// Update set object with name field, if changed
	if len(editTopicRequest.Name) > 0 && topic.Name != editTopicRequest.Name {
		topic_update_data["$set"].(gin.H)["name"] = editTopicRequest.Name
	}

	// Update set object with is_enabled field, if changed
	if topic.IsEnabled != editTopicRequest.IsEnabled {
		topic_update_data["$set"].(gin.H)["is_enabled"] = editTopicRequest.IsEnabled
	}

	// Validation of data change
	if len(topic_update_data["$set"].(gin.H)) > 0 {
		// update topic using the helper method
		err = handlers.topicHelper.UpdateTopicByIdHelper(topic.ID, topic_update_data)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	}

	// Fetch Updated topic Response
	topicResponse, err := fetchTopicByIDResponse(handlers, topicId, communityId)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"success": true,
		"topic":   topicResponse,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}
