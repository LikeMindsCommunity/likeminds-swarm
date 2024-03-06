package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to parse topic for response
func parseTopicResponse(topic *entities.Topic) responses.TopicResponse {
	var response responses.TopicResponse

	response.ID = topic.ID.Hex()
	response.Name = topic.Name
	response.IsEnabled = topic.IsEnabled
	response.Priority = topic.Priority
	response.IsSearchable = topic.IsSearchable
	response.ParentName = topic.ParentName
	response.Level = topic.Level

	if topic.ParentId != primitive.NilObjectID {
		response.ParentId = topic.ParentId.Hex()
	}

	if topic.WidgetId != primitive.NilObjectID {
		response.WidgetId = topic.WidgetId.Hex()
	}

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
func fetchTopicsByIDs(helper interfaces.TopicHelper, topicIds []primitive.ObjectID, communityId int, filterEnabled bool) ([]entities.Topic, error) {

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

// validate and update create topics request with default values
func validateAndUpdateCreateTopicsRequest(topicHelper interfaces.TopicHelper, createTopicsRequest requests.CreateTopicsRequest, communityId int) ([]requests.CreateTopicRequest, error) {

	// throw error if Names or Topics are empty
	if len(createTopicsRequest.Names) == 0 && len(createTopicsRequest.Topics) == 0 {
		return nil, fmt.Errorf("send names or topics to create new topics")
	}

	// Update Topics array with names if names are sent instead of topics meta
	if len(createTopicsRequest.Names) > 0 && len(createTopicsRequest.Topics) == 0 {
		createTopicsRequest.Topics = make([]requests.CreateTopicRequest, len(createTopicsRequest.Names))

		for i, name := range createTopicsRequest.Names {
			createTopicsRequest.Topics[i].Name = name
		}

	}

	topicNames, parentIds := make([]string, len(createTopicsRequest.Topics)), []primitive.ObjectID{}

	for i, topic := range createTopicsRequest.Topics {

		// validate and update topic name
		createTopicsRequest.Topics[i].Name = strings.Trim(topic.Name, " ")
		if createTopicsRequest.Topics[i].Name == "" {
			return nil, fmt.Errorf("can't Create Topic With Empty Name")
		}

		topicNames[i] = createTopicsRequest.Topics[i].Name

		// get parent_id from the topic and convert it to ObjectID
		if topic.ParentId != "" {
			parentId, err := primitive.ObjectIDFromHex(topic.ParentId)
			if err != nil {
				return nil, fmt.Errorf("invalid Parent ID")
			}

			parentIds = append(parentIds, parentId)
		}
	}

	// Check if any duplicate topic names are sent
	duplicates := utils.GetDuplicatesFromSlice(topicNames)
	if len(duplicates) > 0 {
		return nil, fmt.Errorf(fmt.Sprintf("Duplicate Topic Names Sent: %v", duplicates))
	}

	// fetch and validate parent topics
	parentTopics, err := fetchTopicsByIDs(topicHelper, parentIds, communityId, false)
	if err != nil {
		return nil, err
	}

	if len(parentTopics) != len(parentIds) {
		return nil, fmt.Errorf("invalid Parent topic ID/s Sent")
	}

	// update topics with default and parent data
	for i, topic := range createTopicsRequest.Topics {

		// set is_searchable to true if not sent
		if topic.IsSearchable == nil {
			ok := true
			createTopicsRequest.Topics[i].IsSearchable = &ok
		}

		// set is_enabled to true if not sent
		if topic.IsEnabled == nil {
			ok := true
			createTopicsRequest.Topics[i].IsEnabled = &ok
		}

		// update parent meta for topic
		if createTopicsRequest.Topics[i].ParentId != "" {
			createTopicsRequest.Topics[i].ParentName = parentTopics[i].Name
			createTopicsRequest.Topics[i].Level = parentTopics[i].Level + 1

			if parentTopics[i].ParentId != primitive.NilObjectID {
				createTopicsRequest.Topics[i].AllParentIds = append(parentTopics[i].AllParentIds, parentTopics[i].ParentId)
			} else {
				createTopicsRequest.Topics[i].AllParentIds = append([]primitive.ObjectID{}, parentTopics[i].ID)
			}
		}
	}

	// check if the topics already exists with the same parent
	var conditions []bson.M

	for _, topic := range createTopicsRequest.Topics {

		parentId := primitive.ObjectID{}

		if topic.ParentId != "" {
			// Convert parent_id to ObjectID
			parentId, err = primitive.ObjectIDFromHex(topic.ParentId)
			if err != nil {
				return nil, fmt.Errorf("invalid Parent ID")
			}
		}

		condition := bson.M{
			"$and": bson.A{
				bson.M{"name": bson.M{"$regex": topic.Name, "$options": "i"}},
				bson.M{"parent_id": parentId},
			},
		}
		conditions = append(conditions, condition)
	}

	filter := gin.H{
		"$or":          conditions,
		"community_id": communityId,
	}

	// find if any existing topics exists
	existingTopics, err := topicHelper.FindTopicHelper(filter, gin.H{})
	if err != nil {
		return nil, err
	}

	// send error if the topic already exists
	if len(existingTopics) > 0 {
		if existingTopics[0].ParentName != "" {
			return nil, fmt.Errorf(fmt.Sprintf("Topic %v already exists with a Parent %v", existingTopics[0].Name, existingTopics[0].ParentName))
		}
		return nil, fmt.Errorf(fmt.Sprintf("Topic %v already exists", existingTopics[0].Name))
	}

	return createTopicsRequest.Topics, nil
}

// internal method to create topics after validation
func createTopicsAfterValidation(handlers *FeedHandlers, topicsRequest []requests.CreateTopicRequest, communityId int) ([]entities.Topic, error) {

	// create topics using the helper method
	topicIds, err := handlers.topicHelper.CreateManyTopicsHelper(topicsRequest, communityId)
	if err != nil {
		return nil, err
	}

	// fetch newly created topic objects from db using IDs
	topicsData, err := fetchTopicsByIDs(handlers.topicHelper, topicIds, communityId, false)
	if err != nil {
		return nil, err
	}

	// Create Topic widget metadata if exists
	for i, topicRequest := range topicsRequest {

		if topicRequest.Metadata != nil {

			// create widget from given metadata
			widgetData, err := createWidget(handlers, false, topicsData[i].ID.Hex(), enums.WidgetParentEntityTypeTopic, topicRequest.Metadata, nil, communityId)
			if err != nil {
				return nil, err
			}

			setData := gin.H{
				"$set": gin.H{
					"widget_id": widgetData.ID,
				},
			}

			// update topic with widget_id
			err = handlers.topicHelper.UpdateTopicByIdHelper(topicsData[i].ID, setData, false)
			if err != nil {
				return nil, err
			}

			topicsData[i].WidgetId = widgetData.ID
		}
	}

	parentTopicsIds := []primitive.ObjectID{}

	for _, topic := range topicsData {
		if topic.ParentId != primitive.NilObjectID {
			parentTopicsIds = append(parentTopicsIds, topic.ParentId)
		}
	}

	// update parent topics with child count
	if len(parentTopicsIds) > 0 {

		filterData := bson.M{"_id": bson.M{
			"$in": parentTopicsIds,
		}}

		updateData := bson.M{
			"$inc": bson.M{
				"total_child_count": 1,
			},
		}

		// Update and fetch parent topics
		handlers.topicHelper.UpdateManyTopicsHelper(filterData, updateData, false)

		topics, err := handlers.topicHelper.FindTopicHelper(filterData, gin.H{})
		if err != nil {
			logging.Error("Error while fecthing data for reindexing | ", err.Error())
		}

		topicsIndexData := map[string]interface{}{}

		// parse the topics data for ES indexing and API response
		for _, topicData := range topics {
			topicsIndexData[topicData.ID.Hex()] = ParseTopicIndexData(handlers.postHelper, &topicData, false)
		}

		// index topics data in elastic search
		err = handlers.esHelper.InsertManyDocuments(topicsIndexData, constants.TopicIndexName)
		if err != nil {
			logging.Error(err.Error())
		}
	}

	topicsIndexData := make(map[string]interface{})

	// parse the topics data for ES indexing and API response
	for _, topicData := range topicsData {
		topicsIndexData[topicData.ID.Hex()] = ParseTopicIndexData(handlers.postHelper, &topicData, false)

	}

	// index topics data in elastic search
	err = handlers.esHelper.InsertManyDocuments(topicsIndexData, constants.TopicIndexName)
	if err != nil {
		logging.Error(err.Error())
	}

	return topicsData, nil
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

	// validate create topics request
	topicsRequest, err := validateAndUpdateCreateTopicsRequest(handlers.topicHelper, createTopicsRequest, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// create topics using the validated request
	topicsData, err := createTopicsAfterValidation(handlers, topicsRequest, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// parse the topics data for API response
	topicsResponse := []responses.TopicResponse{}

	for _, topicsData := range topicsData {
		topicsResponse = append(topicsResponse, parseTopicResponse(&topicsData))
	}

	// generate final response
	utils.GenerateSuccessResponse(c, gin.H{"topics": topicsResponse})
}

func processTopicSearchData(data map[string]interface{}) []responses.TopicResponseWithMeta {
	topicDetails := data["hits"].(map[string]interface{})["hits"].([]interface{})
	fetchTopicsResponse := []responses.TopicResponseWithMeta{}

	for _, data := range topicDetails {
		topicData := data.(map[string]interface{})["_source"].(map[string]interface{})
		topicData["_id"] = topicData["id"]

		// convert the data to fetch topic response
		var topic responses.TopicResponseWithMeta
		b, _ := json.Marshal(topicData)
		json.Unmarshal(b, &topic)

		fetchTopicsResponse = append(fetchTopicsResponse, topic)
	}

	return fetchTopicsResponse
}

func validateAndGetFetchTopicsFilterParams(fetchTopicRequest requests.FetchTopicRequest) (int, bool, bool, []string, []string, error) {

	// fetch min_posts query param
	minPosts, err := utils.ParseIntFromQueryParam(fetchTopicRequest.MinPosts, 0)
	if err != nil {
		return 0, false, false, nil, nil, err
	}

	if minPosts <= 0 {
		minPosts = 0
	}

	filterIsEnabled := false
	isEnabled := false
	if fetchTopicRequest.IsEnabled != "" {
		filterIsEnabled = true

		if fetchTopicRequest.IsEnabled == "true" {
			isEnabled = true
		}
	}

	// validate order_by query param
	orderByParams := []string{enums.OrderByAlphabeticalAsc}
	if fetchTopicRequest.OrderBy != "" {
		orderByParams := parseStringArrayParam(fetchTopicRequest.OrderBy)
		for _, orderBy := range orderByParams {
			isValid := enums.IsValidOrderByParam(orderBy)
			if !isValid {
				return 0, false, false, nil, nil, fmt.Errorf("Invalid value for order_by")
			}
		}
	}

	parentTopicsIds := parseStringArrayParam(fetchTopicRequest.ParentTopicIds)

	return minPosts, filterIsEnabled, isEnabled, orderByParams, parentTopicsIds, nil
}

// Exposed Method to Fetch Topics for a Community
func (handlers *FeedHandlers) FetchTopics(c *gin.Context) {

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// parse fetch topic request
	var fetchTopicRequest requests.FetchTopicRequest
	err := c.BindQuery(&fetchTopicRequest)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	minPosts, filterIsEnabled, isEnabled, orderByParams, parentTopicsIds, err := validateAndGetFetchTopicsFilterParams(fetchTopicRequest)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch pagination query params
	page, pageSize, err := fetchPaginationParams(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	response := gin.H{
		"topics":       []responses.TopicResponseWithMeta{},
		"child_topics": map[string][]responses.TopicResponseWithMeta{},
	}

	if len(parentTopicsIds) > 0 {
		parentTopicQuery := GetTopicIdsFilterQuery(parentTopicsIds, communityId)
		esResponse := handlers.esHelper.ExecuteQuery(parentTopicQuery, constants.TopicIndexName)

		parentTopics := processTopicSearchData(esResponse)
		if len(parentTopics) != len(parentTopicsIds) {
			utils.GeneralAPIValidationError(c, "Invalid Parent Topic Ids")
			return
		}

		response["topics"] = parentTopics

		for _, parentTopic := range parentTopics {

			// fetch child topics of parent
			topicQuery := GetTopicFilterQuery(page, pageSize, fetchTopicRequest.SearchType, fetchTopicRequest.Search,
				communityId, filterIsEnabled, isEnabled, minPosts, orderByParams, parentTopic.ID)
			esResponse := handlers.esHelper.ExecuteQuery(topicQuery, constants.TopicIndexName)
			childTopics := processTopicSearchData(esResponse)

			// add child topics to the response
			response["child_topics"].(map[string]interface{})[parentTopic.ID] = childTopics
		}

	} else {

		// ES query to search topics
		topicQuery := GetTopicFilterQuery(page, pageSize, fetchTopicRequest.SearchType,
			fetchTopicRequest.Search, communityId, filterIsEnabled, isEnabled, minPosts, orderByParams, "")

		// execute the query
		esResponse := handlers.esHelper.ExecuteQuery(topicQuery, constants.TopicIndexName)

		// process the response data and add it to response
		response["topics"] = processTopicSearchData(esResponse)
	}

	// return final response
	utils.GenerateSuccessResponse(c, response)
}

func validateEditTopicRequest(handlers *FeedHandlers, topicId string, editTopicRequest requests.EditTopicRequest,
	communityId int) (*entities.Topic, gin.H, error) {

	// fetch topic using topic_id
	topic, err := fetchTopicByID(handlers.topicHelper, topicId, communityId)
	if err != nil {
		return nil, nil, err
	}

	// topic update data
	topicUpdateData := gin.H{
		"$set": gin.H{},
	}

	// strip text to check if it is empty
	editTopicRequest.Name = strings.Trim(editTopicRequest.Name, " ")

	// Update set object with name field, if changed
	if len(editTopicRequest.Name) > 0 && topic.Name != editTopicRequest.Name {

		// check if the topic name already exists
		filter := gin.H{
			"name":         editTopicRequest.Name,
			"parent_id":    topic.ParentId,
			"community_id": communityId,
		}

		// fetch topic using helper method
		topicResults, err := handlers.topicHelper.FindTopicHelper(filter, gin.H{})
		if err != nil {
			return nil, nil, err
		}

		// send error if the topic already exists
		if len(topicResults) > 0 {
			return nil, nil, fmt.Errorf(fmt.Sprintf("Topic %v with this name already exists", editTopicRequest.Name))
		}

		topicUpdateData["$set"].(gin.H)["name"] = editTopicRequest.Name
	}

	// Update set object with is_enabled field, if changed
	if editTopicRequest.IsEnabled != nil && *editTopicRequest.IsEnabled != topic.IsEnabled {
		topicUpdateData["$set"].(gin.H)["is_enabled"] = editTopicRequest.IsEnabled
	}

	// Update set object with priority field, if changed
	if topic.Priority != editTopicRequest.Priority {
		topicUpdateData["$set"].(gin.H)["priority"] = editTopicRequest.Priority
	}

	// Update set object with is_searchable field, if changed
	if editTopicRequest.IsSearchable != nil && *editTopicRequest.IsSearchable != topic.IsSearchable {
		topicUpdateData["$set"].(gin.H)["is_searchable"] = editTopicRequest.IsSearchable
	}

	return topic, topicUpdateData, nil
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

	topic, topicUpdateData, err := validateEditTopicRequest(handlers, topicId, editTopicRequest, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Update set object with widget_id field, if metadata is sent
	if editTopicRequest.Metadata != nil {

		// check if the topic already has a widget
		if topic.WidgetId != primitive.NilObjectID {

			// update widget using the helper method
			_, err := editWidget(handlers, topic.WidgetId.Hex(), topicId, enums.WidgetParentEntityTypeTopic, false, editTopicRequest.Metadata, nil, communityId)
			if err != nil {
				utils.GeneralAPIInternalError(c, err.Error())
				return
			}
		} else {

			// create widget from given metadata
			widgetData, err := createWidget(handlers, false, topic.ID.Hex(), enums.WidgetParentEntityTypeTopic, editTopicRequest.Metadata, nil, communityId)
			if err != nil {
				utils.GeneralAPIInternalError(c, err.Error())
				return
			}

			topicUpdateData["$set"].(gin.H)["widget_id"] = widgetData.ID
		}
	}

	// Validation of data change
	if len(topicUpdateData["$set"].(gin.H)) > 0 {

		// update topic using the helper method
		err := handlers.topicHelper.UpdateTopicByIdHelper(topic.ID, topicUpdateData, true)
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
		err = handlers.esHelper.IndexDocument(ParseTopicIndexData(handlers.postHelper, topic, true), topic.ID.Hex(), constants.TopicIndexName)
		if err != nil {
			logging.Error(err.Error())
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

func validateDeleteTopicsRequest(handlers *FeedHandlers, communityId int, deleteTopicsRequest requests.DeleteTopicsRequest,
) ([]primitive.ObjectID, []primitive.ObjectID, error) {

	// convert topic_ids to object ids
	topicIDs := helpers.ConvertIdsToObjectIds(deleteTopicsRequest.TopicIds)

	// send error if no topic ids are sent in the body
	if len(topicIDs) <= 0 {
		return nil, nil, fmt.Errorf("topic_ids can't be empty!")
	}

	// fetch all the topics by topic ids sent in request
	topics, err := fetchTopicsByIDs(handlers.topicHelper, topicIDs, communityId, false)
	if err != nil {
		return nil, nil, err
	}

	// Validation of Topics
	if len(topics) != len(topicIDs) {
		return nil, nil, fmt.Errorf("Invalid topic_ids sent")
	}

	widgetIds := []primitive.ObjectID{}

	for _, topic := range topics {
		// check if the topic has a widget
		if topic.WidgetId != primitive.NilObjectID {
			widgetIds = append(widgetIds, topic.WidgetId)
		}
	}

	return topicIDs, widgetIds, nil
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

	topicIDs, widgetIDs, err := validateDeleteTopicsRequest(handlers, communityId, deleteTopicsRequest)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// delete the topics from db
	err = handlers.topicHelper.DeleteTopicsHelper(topicIDs)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// delete the widgets for the topics
	if len(widgetIDs) > 0 {

		filter := gin.H{
			"_id": gin.H{
				"$in": widgetIDs,
			},
		}

		count, err := handlers.widgetHelper.DeleteWidgetsHelper(filter)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
		if int(count) != len(widgetIDs) {
			logging.Error(fmt.Sprintf("Only %v Widgets deleted out of %v", count, widgetIDs))
		}
	}

	topicidsString := utils.ParseStringArrayToString(deleteTopicsRequest.TopicIds)

	// query to get topics to be deleted
	getTopicsToBeDeletedQuery := GetTopicsByIdQuery(topicidsString)

	// delete topics from elastic search
	err = handlers.esHelper.DeleteByQuery(getTopicsToBeDeletedQuery, constants.TopicIndexName)
	if err != nil {
		logging.Error(err.Error())
	}

	err = deleteTopicsFromPostsAndUpdatePost(handlers, topicIDs)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// response data
	response := gin.H{
		"success": true,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// deletes topics from posts and update the post in ES index
func deleteTopicsFromPostsAndUpdatePost(handlers *FeedHandlers, topicIDs []primitive.ObjectID) error {

	// Create a filter to find posts to be updated
	filter := bson.M{
		"topic_ids": bson.M{
			"$in": topicIDs,
		},
	}

	// find posts based on the filter
	postResults, err := handlers.postHelper.FindPostHelper(filter, gin.H{})
	if err != nil {
		return err
	}

	// extract postIds from the postResults
	postIDs := []primitive.ObjectID{}
	for _, post := range postResults {
		postIDs = append(postIDs, post.ID)
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
		return err
	}

	// fetch the updated posts and update in ES
	updatedPosts, err := handlers.postHelper.FindPostHelper(filter, gin.H{})
	if err != nil {
		return err
	}

	for _, postData := range updatedPosts {
		// update post data in elastic search
		err = handlers.esHelper.IndexDocument(ParsePostIndexData(&postData), postData.ID.Hex(), constants.PostIndexName)
		if err != nil {
			logging.Error(err.Error())
		}
	}

	return nil
}
