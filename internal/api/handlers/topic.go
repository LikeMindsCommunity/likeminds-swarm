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
	"go.mongodb.org/mongo-driver/bson"
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

// Exposed Method to Create Topics
func (handlers *FeedHandlers) CreateTopics(c *gin.Context) {
	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var createTopicsRequest requests.CreateTopicsRequest
	if err := c.ShouldBindJSON(&createTopicsRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// throw error if Names key is empty
	if len(createTopicsRequest.Names) == 0 {
		utils.GeneralAPIValidationError(c, "Send topic names to create new topics")
		return
	}

	// parses the topic names from request and creates corresponding list
	topicsList, lowerCaseTopicsList, err := parseAndValidateTopicsRequest(createTopicsRequest.Names)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// filter to find existing topics with same name
	filter := gin.H{
		"$expr": bson.M{
			"$in": bson.A{
				bson.M{
					"$toLower": "$name",
				},
				lowerCaseTopicsList,
			},
		},
	}

	// find the existing topics with same name as in the request
	existingTopics, err := handlers.topicHelper.FindTopicHelper(filter, gin.H{})
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// send error if the topic already exists
	if len(existingTopics) > 0 {
		utils.GeneralAPIValidationError(c, fmt.Sprintf(`Topic %q already exists.`, existingTopics[0].Name))
		return
	}

	// create topics using the helper method
	topicIds, err := handlers.topicHelper.CreateManyTopicsHelper(topicsList, true, communityId)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// convert topic_ids to object ids
	objectTopicIds := helpers.TypecastIdsToObjectIds(topicIds)

	// fetch newly created topic objects from db using IDs
	topicsData, err := fetchTopicsByIDs(handlers.topicHelper, objectTopicIds, communityId, false)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	topicsIndexData := make(map[string]interface{})
	var topicsResponse []requests.TopicResponse

	postsCount := 0

	// parse the topics data for ES indexing and API response
	for _, topicData := range topicsData {
		topicsIndexData[topicData.ID.Hex()] = ParseTopicIndexData(&topicData, &postsCount)
		topicsResponse = append(topicsResponse, parseTopicResponse(&topicData))
	}

	// insert topics data in elastic search
	err = handlers.esHelper.InsertManyDocuments(topicsIndexData, constants.TopicIndexName)
	if err != nil {
		log.Error(err.Error())
	}

	// reponse data
	response := gin.H{
		"success": true,
		"topics":  topicsResponse,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// parses and validates the create topics request
func parseAndValidateTopicsRequest(topicNames []string) ([]string, []string, error) {
	topicsList, lowerCaseTopicsList := []string{}, []string{}

	// strip each name in the names array and check if there are any duplicate topic in the array
	seen := map[string]bool{}

	for i, name := range topicNames {

		topicNames[i] = strings.Trim(name, " ")

		if topicNames[i] == "" {
			return nil, nil, fmt.Errorf("Can't Create Topic With Empty Name")
		}

		lowerCaseTopicName := strings.ToLower(topicNames[i])

		if seen[lowerCaseTopicName] {
			// Duplicate found
			return nil, nil, fmt.Errorf("Can't create duplicate topics")
		}

		lowerCaseTopicsList = append(lowerCaseTopicsList, lowerCaseTopicName)
		seen[lowerCaseTopicName] = true
	}

	topicsList = topicNames
	return topicsList, lowerCaseTopicsList, nil
}

func processTopicSearchData(handlers *FeedHandlers, data map[string]interface{}) []requests.FetchTopicsResponse {
	topicDetails := data["hits"].(map[string]interface{})["hits"].([]interface{})
	fetchTopicsResponse := []requests.FetchTopicsResponse{}

	for _, data := range topicDetails {
		topicData := data.(map[string]interface{})["_source"].(map[string]interface{})
		topicData["_id"] = topicData["id"]

		// convert the data to fetch topic response
		var topic requests.FetchTopicsResponse
		b, _ := json.Marshal(topicData)
		json.Unmarshal(b, &topic)

		fetchTopicsResponse = append(fetchTopicsResponse, topic)
	}

	return fetchTopicsResponse
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

	// fetch min_posts query param
	minPosts, _ := utils.ParseIntFromQueryParam(c.DefaultQuery("min_posts", "0"), 0)

	if minPosts <= 0 {
		minPosts = 0
	}

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
		fetchTopicRequest.Search, communityId, filterIsEnabled, isEnabled, minPosts)
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
		err = handlers.esHelper.UpdateDocument(c, ParseTopicIndexData(topic, nil), topic.ID.Hex(), constants.TopicIndexName)
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
	if err := c.ShouldBindJSON(&deleteTopicsRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

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

	topicidsString := utils.ParseStringArrayToString(deleteTopicsRequest.TopicIds)

	// query to get topics to be deleted
	getTopicsToBeDeletedQuery := GetTopicsByIdQuery(topicidsString)

	// delete topics from elastic search
	err = handlers.esHelper.DeleteByQuery(getTopicsToBeDeletedQuery, constants.TopicIndexName)
	if err != nil {
		fmt.Println(err.Error())
	}

	// Create a filter to find posts to be updated
	filter := bson.M{
		"topic_ids": bson.M{
			"$in": topicIDs,
		},
	}

	// find posts based on the filter
	postResults, err := handlers.postHelper.FindPostHelper(filter, gin.H{})

	// extract postIds from the postResults
	postIDs, postIDsString := []primitive.ObjectID{}, []string{}
	for _, post := range postResults {
		postIDs = append(postIDs, post.ID)
		postIDsString = append(postIDsString, post.ID.String())
	}

	// update the filter to update posts by postIds
	filter = bson.M{
		"_id": bson.M{
			"$in": postIDs,
		},
	}

	// Create an update to pull the specified topic IDs from the array
	update := bson.M{
		"$pull": bson.M{
			"topic_ids": bson.M{
				"$in": topicIDs,
			},
		},
	}

	// deletes the topics from posts based on the passed filter and update query
	err = handlers.postHelper.UpdateManyPostsHelper(filter, update, true)
	if err != nil {
		log.Error(err.Error())
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// fetch the updated posts and update in ES
	updatedPosts, err := handlers.postHelper.FindPostHelper(filter, gin.H{})

	for _, postData := range updatedPosts {
		// update post data in elastic search
		err = handlers.esHelper.IndexDocument(c, ParsePostIndexData(&postData), postData.ID.Hex(), constants.PostIndexName)
		if err != nil {
			log.Error(err.Error())
		}
	}

	// response data
	response := gin.H{
		"success": true,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}
