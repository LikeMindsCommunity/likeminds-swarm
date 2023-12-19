package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	log "github.com/nateshr/likeminds-swarm/internal/services/logging"
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

	topicData, err := fetchTopicByID(handlers.topicHelper, topicId.(primitive.ObjectID).Hex(), communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// insert topic data in elastic search
	err = handlers.esHelper.InsertDocument(c, ParseTopicIndexData(topicData), topicData.ID.Hex(),
		constants.TopicIndexName)
	if err != nil {
		log.Error(err.Error())
	}

	topicResponse := parseTopicResponse(topicData)

	// reponse data
	response := gin.H{
		"success": true,
		"topic":   topicResponse,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

func processTopicSearchData(handlers *FeedHandlers, data map[string]interface{}) []requests.TopicResponse {
	topicDetails := data["hits"].(map[string]interface{})["hits"].([]interface{})
	var topicList []entities.Topic

	for _, data := range topicDetails {
		topicData := data.(map[string]interface{})["_source"].(map[string]interface{})
		topicData["_id"] = topicData["id"]

		// convert the data to topic entity
		var topic entities.Topic
		b, _ := json.Marshal(topicData)
		json.Unmarshal(b, &topic)

		topicList = append(topicList, topic)
	}

	topicsResponse := []requests.TopicResponse{}

	// Parse all fetched topics Data
	for _, topic := range topicList {
		topicsResponse = append(topicsResponse, parseTopicResponse(&topic))
	}

	return topicsResponse
}

// Exposed Method to Fetch Topics for a Community
func (handlers *FeedHandlers) FetchTopics(c *gin.Context) {
	// parse fetch topic request
	var fetchTopicRequest requests.FetchTopicRequest
	err := c.BindQuery(&fetchTopicRequest)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// fetch pagination query params
	page, pageSize, err := fetchPaginationParams(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	filterIsEnabled := false
	isEnabled := false
	if fetchTopicRequest.IsEnabled != "" {
		filterIsEnabled = true

		if fetchTopicRequest.IsEnabled == "true" {
			isEnabled = true
		}
	}

	// dsl query to search topics
	topicQuery := GetTopicFilterQuery(page, pageSize, fetchTopicRequest.SearchType,
		fetchTopicRequest.Search, communityId, filterIsEnabled, isEnabled)
	response := handlers.esHelper.ExecuteQuery(topicQuery, constants.TopicIndexName)

	finalResponse := processTopicSearchData(handlers, response)

	// reponse data
	finalParsedResponse := gin.H{
		"success": true,
		"topics":  finalResponse,
	}

	// return final response
	c.JSON(http.StatusOK, finalParsedResponse)
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
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// topic update data
	topicUpdateData := gin.H{
		"$set": gin.H{},
	}

	// strip text to check if it is empty
	editTopicRequest.Name = strings.Trim(editTopicRequest.Name, " ")

	// Update set object with name field, if changed
	if len(editTopicRequest.Name) > 0 && topic.Name != editTopicRequest.Name {
		topicUpdateData["$set"].(gin.H)["name"] = editTopicRequest.Name
	}

	// Update set object with is_enabled field, if changed
	if topic.IsEnabled != editTopicRequest.IsEnabled {
		topicUpdateData["$set"].(gin.H)["is_enabled"] = editTopicRequest.IsEnabled
	}

	// Validation of data change
	if len(topicUpdateData["$set"].(gin.H)) > 0 {
		// update topic using the helper method
		err = handlers.topicHelper.UpdateTopicByIdHelper(topic.ID, topicUpdateData)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		topic, err = fetchTopicByID(handlers.topicHelper, topic.ID.Hex(), communityId)
		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}

		// update topic data in elastic search
		err = handlers.esHelper.UpdateDocument(c, ParseTopicIndexData(topic), topic.ID.Hex(), constants.TopicIndexName)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	// Fetch Updated topic Response
	topicResponse := parseTopicResponse(topic)

	// reponse data
	response := gin.H{
		"success": true,
		"topic":   topicResponse,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to Delete Topics
func (handlers *FeedHandlers) DeleteTopics(c *gin.Context) {
	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var deleteTopicsRequest requests.DeleteTopicsRequest

	// convert topic_ids to object ids
	topicIDs := helpers.ConvertIdsToObjectIds(deleteTopicsRequest.TopicIds)

	// send error if no topic ids are sent in the body
	if len(topicIDs) <= 0 {
		utils.GeneralAPIValidationError(c, "topic_ids can't be empty!")
		return
	}

	// fetch all the topics by topic ids sent in request
	topics, err := fetchTopicsByIDs(handlers.topicHelper, topicIDs, communityId, false)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Validation of Topics
	if len(topics) != len(topicIDs) {
		utils.GeneralAPIValidationError(c, "Invalid topic_ids sent")
		return
	}

	// delete the topics from db
	err = handlers.topicHelper.DeleteTopicsHelper(topicIDs)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// reponse data
	response := gin.H{
		"success": true,
	}

	// todo: delete from es and other things

	// return final response
	c.JSON(http.StatusOK, response)
}
