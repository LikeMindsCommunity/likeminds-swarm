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
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/services/searchElastic"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to create a new widget with indexing
func createWidget(handlers *FeedHandlers, createdByLM bool, parentEntityID string, parentEntityType string,
	metaData map[string]interface{}, lmMeta map[string]interface{}, communityId int) (*entities.Widget, error) {

	// create Widget using the helper method
	widgetId, err := handlers.widgetHelper.CreateWidgetHelper(createdByLM, parentEntityID, parentEntityType, metaData, lmMeta,
		communityId)
	if err != nil {
		return nil, err
	}

	widgetData, err := fetchWidgetByID(handlers.widgetHelper, widgetId.(primitive.ObjectID).Hex(), createdByLM, communityId)
	if err != nil {
		return nil, err
	}

	// insert widget data in elastic search
	err = handlers.esHelper.IndexDocument(ParseWidgetIndexData(widgetData), widgetData.ID.Hex(), constants.WidgetIndexName)
	if err != nil {
		logging.Error(err.Error())
	}

	return widgetData, nil
}

// Internal Method to edit an existing widget with index update
func editWidget(handlers *FeedHandlers, widgetId string, parentEntityId string, parentEntityType string,
	createdByLM bool, metaData map[string]interface{}, lmMeta map[string]interface{}, communityId int) (*entities.Widget, error) {

	// fetch Widget using widget_id
	widget, err := fetchWidgetByID(handlers.widgetHelper, widgetId, createdByLM, communityId)
	if err != nil {
		return widget, err
	}

	// widget update data
	widgetUpdateData := gin.H{
		"$set": gin.H{},
	}

	// Update set object with parentEntityId field, if changed
	if parentEntityId != "" {
		widgetUpdateData["$set"].(gin.H)["parent_entity_id"] = parentEntityId
	}

	// Update set object with parentEntityType field, if changed
	if parentEntityType != "" {
		widgetUpdateData["$set"].(gin.H)["parent_entity_type"] = parentEntityType
	}

	// Update set object with metadata field, if changed
	if metaData != nil {
		widgetUpdateData["$set"].(gin.H)["metadata"] = metaData
	}

	// Update set object with lmmeta field, if changed
	if lmMeta != nil {
		widgetUpdateData["$set"].(gin.H)["_lm_meta"] = lmMeta
	}

	// Validation of data change
	if len(widgetUpdateData["$set"].(gin.H)) > 0 {
		// update widget using the helper method
		err := handlers.widgetHelper.UpdateWidgetByIdHelper(widget.ID, widgetUpdateData)
		if err != nil {
			return nil, err
		}

		widget, err = fetchWidgetByID(handlers.widgetHelper, widget.ID.Hex(), createdByLM, communityId)
		if err != nil {
			return nil, err
		}

		// Index updated widget data in elastic search
		err = handlers.esHelper.IndexDocument(ParseWidgetIndexData(widget), widget.ID.Hex(),
			constants.WidgetIndexName)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	return widget, nil
}

// Internal Method to fetch pollVotes Data Map for poll options
func fetchPollVotesDataMap(handlers *FeedHandlers, entityId string, metaData map[string]interface{},
	communityId int, uuid string) (gin.H, error) {
	pollVotesData := []gin.H{}
	parsedPollVotesData := gin.H{}
	var err error

	pollType, pollTypeExists := metaData["poll_type"]
	pollVote, _ := GetPollVoteOfUUID(handlers, entityId, communityId, uuid)

	if !(pollTypeExists && pollType == enums.InstantPollType && pollVote == nil) {
		// Fetch poll Votes Data
		pollVotesData, err = getPollVotesDataUsingAggregation(handlers, entityId, communityId, uuid)
		if err != nil {
			return parsedPollVotesData, err
		}
	}

	// Process poll votes data
	for _, pollVoteData := range pollVotesData {
		if optionId, exists := pollVoteData["_id"]; exists {
			parsedPollVotesData[optionId.(string)] = pollVoteData
		}
	}

	return parsedPollVotesData, nil
}

// Internal Method to parse LM meta object for response
func parseLMMeta(handlers *FeedHandlers, entityId string, metaData map[string]interface{}, lmMeta map[string]interface{},
	communityId int, uuid string) map[string]interface{} {
	// If option exists in LM Meta, it is a poll widget
	if _, exists := lmMeta["options"]; exists {
		// fetch poll votes data
		parsedPollVotesData, err := fetchPollVotesDataMap(handlers, entityId, metaData, communityId, uuid)
		if err != nil {
			return lmMeta
		}

		// option data conversion to desired type
		options := []gin.H{}
		convertedOptions, _ := json.Marshal(lmMeta["options"])
		_ = json.Unmarshal(convertedOptions, &options)

		// Merge option data with votes data
		for _, option := range options {
			if optionId, exists := option["_id"]; exists {
				voteData := parsedPollVotesData[optionId.(string)]

				if voteData != nil {
					for key, value := range voteData.(gin.H) {
						option[key] = value
					}
				} else {
					option["vote_count"] = 0
					option["is_selected"] = false
					option["percentage"] = 0
				}
			}
		}

		lmMeta["options"] = options
	}

	return lmMeta
}

// Internal Method to parse Widget for response
func parseWidgetResponse(handlers *FeedHandlers, widget *entities.Widget, communityId int, uuid string) requests.WidgetResponse {
	var response requests.WidgetResponse

	response.ID = widget.ID
	response.ParentEntityID = widget.ParentEntityID
	response.ParentEntityType = widget.ParentEntityType
	response.MetaData = widget.MetaData
	response.LMMeta = parseLMMeta(handlers, widget.ID.Hex(), widget.MetaData, widget.LMMeta, communityId, uuid)

	response.CreatedAt = int(widget.CreatedAt.UnixMilli())
	response.UpdatedAt = int(widget.UpdatedAt.UnixMilli())
	return response
}

// Internal Method to fetch widgets using widgetIds and communityId
func fetchWidgetsByIDs(helper interfaces.WidgetHelper, widgetIds []primitive.ObjectID, communityId int) ([]entities.Widget, error) {
	// widget filter data
	widgetFilterData := gin.H{
		"_id": gin.H{
			"$in": widgetIds,
		},
		"community_id": communityId,
	}

	// fetch widget using helper method
	widgetResults, err := helper.FindWidgetHelper(widgetFilterData, gin.H{})
	if err != nil {
		return nil, err
	}

	return widgetResults, nil
}

// Internal Method to fetch widget using widget id and communityId
func fetchWidgetByID(helper interfaces.WidgetHelper, widgetId string, createdByLM bool, communityId int) (*entities.Widget, error) {
	// widget filter data
	widgetFilterData := gin.H{
		"_id":           widgetId,
		"community_id":  communityId,
		"created_by_lm": createdByLM,
	}

	// fetch widget using helper method
	widgetResults, err := helper.FindWidgetHelper(widgetFilterData, gin.H{})
	if err != nil {
		return nil, err
	}

	// validation of widgetId
	if len(widgetResults) == 0 {
		return nil, fmt.Errorf("invalid widget id sent")
	}

	return &widgetResults[0], nil
}

// Exposed Method to Create a Widget
func (handlers *FeedHandlers) CreateWidget(c *gin.Context) {
	// fetch headers
	headers := utils.GetHeaders(c)

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var createWidgetRequest requests.CreateWidgetRequest
	if err := c.ShouldBindJSON(&createWidgetRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// create Widget
	widgetData, err := createWidget(handlers, false, createWidgetRequest.ParentEntityID, createWidgetRequest.ParentEntityType,
		createWidgetRequest.MetaData, nil, communityId)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	widgetResponse := parseWidgetResponse(handlers, widgetData, communityId, headers[utils.HeadersMemberId])

	// response data
	response := gin.H{
		"success": true,
		"widget":  widgetResponse,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

func processWidgetSearchData(handlers *FeedHandlers, data map[string]interface{}, communityId int, uuid string) []requests.WidgetResponse {
	widgetDetails := data["hits"].(map[string]interface{})["hits"].([]interface{})
	var widgetList []entities.Widget

	for _, data := range widgetDetails {
		widgetData := data.(map[string]interface{})["_source"].(map[string]interface{})
		widgetData["_id"] = widgetData["id"]

		// convert the data to widget entity
		var widget entities.Widget
		b, _ := json.Marshal(widgetData)
		json.Unmarshal(b, &widget)

		widgetList = append(widgetList, widget)
	}

	widgetResponse := []requests.WidgetResponse{}

	// Parse all fetched Widget Data
	for _, widget := range widgetList {
		widgetResponse = append(widgetResponse, parseWidgetResponse(handlers, &widget, communityId, uuid))
	}

	return widgetResponse
}

// internal method to fetch widgets from DB
func fetchWidgetsFromDB(handlers *FeedHandlers, fetchWidgetRequest *requests.FetchWidgetRequest, communityId int,
	uuid string, page int, pageSize int) ([]requests.WidgetResponse, error) {

	var filter gin.H

	filter_options := gin.H{
		"$skip":  pageSize * (page - 1),
		"$limit": pageSize,
	}
	filter_options = addSortingOptions(filter_options, "updated_at", OrderTypeDescending)

	if fetchWidgetRequest.WidgetIds != "" {

		widgetObjectIds := helpers.ConvertIdsToObjectIds(utils.ParseStringArrayParam(fetchWidgetRequest.WidgetIds))
		filter = gin.H{
			"_id": gin.H{
				"$in": widgetObjectIds,
			},
			"community_id": communityId,
		}

	} else if fetchWidgetRequest.ParentEntityId != "" && fetchWidgetRequest.ParentEntityType != "" {

		filter = gin.H{
			"parent_entity_id":   fetchWidgetRequest.ParentEntityId,
			"parent_entity_type": fetchWidgetRequest.ParentEntityType,
			"community_id":       communityId,
		}
	} else if fetchWidgetRequest.SearchKey != "" && fetchWidgetRequest.SearchValue != "" {

		// remove leading and trailing (inverted commas) from search value
		fetchWidgetRequest.SearchKey = strings.Trim(fetchWidgetRequest.SearchKey, "\"")
		fetchWidgetRequest.SearchValue = strings.Trim(fetchWidgetRequest.SearchValue, "\"")

		// searching only works for string type values
		filter = gin.H{
			fetchWidgetRequest.SearchKey: fetchWidgetRequest.SearchValue,
			"community_id":               communityId,
		}
	} else {
		return nil, fmt.Errorf("invalid search params sent")
	}

	// fetch widget using helper method
	widgetResults, err := handlers.widgetHelper.FindWidgetHelper(filter, filter_options)
	if err != nil {
		return nil, err
	}

	// Parse all fetched Widget Data
	widgetResponse := []requests.WidgetResponse{}
	for _, widget := range widgetResults {
		widgetResponse = append(widgetResponse, parseWidgetResponse(handlers, &widget, communityId, uuid))
	}

	return widgetResponse, nil

}

// internal method to fetch widgets from ES
func fetchParsedWidgetsFromES(handlers *FeedHandlers, fetchWidgetRequest *requests.FetchWidgetRequest, communityId int,
	uuid string, page int, pageSize int) ([]requests.WidgetResponse, error) {

	widgetQuery := ""

	if fetchWidgetRequest.WidgetIds != "" {

		// if widgetIds are sent, fetch widgets by widgetIds
		widgetQuery = GetWidgetByIdsFilterQuery(communityId, fetchWidgetRequest.WidgetIds)

	} else if fetchWidgetRequest.ParentEntityId != "" && fetchWidgetRequest.ParentEntityType != "" {

		// if parentEntityId and parentEntityType are sent, fetch widgets by parentEntityId and parentEntityType
		widgetQuery = GetWidgetsByParentEntityFilterQuery(communityId, fetchWidgetRequest.ParentEntityId,
			fetchWidgetRequest.ParentEntityType)

	} else if fetchWidgetRequest.SearchKey != "" && fetchWidgetRequest.SearchValue != "" {

		// if searchKey and searchValue are sent, fetch widgets by searchKey and searchValue
		widgetQuery = GetWidgetFilterQuery(page, pageSize, communityId, fetchWidgetRequest.SearchKey,
			fetchWidgetRequest.SearchValue)

	}

	// Check for JSON errors
	isValid := json.Valid([]byte(widgetQuery)) // returns bool
	if !isValid {
		return nil, fmt.Errorf("invalid search params sent")
	}

	// Execute ES query
	esResponse := handlers.esHelper.ExecuteQuery(widgetQuery, constants.WidgetIndexName)
	parsedWidgets := processWidgetSearchData(handlers, esResponse, communityId, uuid)

	return parsedWidgets, nil
}

// Exposed Method to Fetch CustomWidgets based on given params
func (handlers *FeedHandlers) FetchWidget(c *gin.Context) {
	// fetch headers
	headers := utils.GetHeaders(c)

	// parse fetch Widget request
	var fetchWidgetRequest requests.FetchWidgetRequest
	err := c.BindQuery(&fetchWidgetRequest)
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

	var parsedWidgets []requests.WidgetResponse

	// fetch widgets from DB or ES
	if constants.FetchWidgetsFromDb {

		// fetch widgets from DB
		parsedWidgets, err = fetchWidgetsFromDB(handlers, &fetchWidgetRequest, communityId,
			headers[utils.HeadersMemberId], page, pageSize)
		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}

	} else {

		// fetch widgets from ES index
		parsedWidgets, err = fetchParsedWidgetsFromES(handlers, &fetchWidgetRequest, communityId,
			headers[utils.HeadersMemberId], page, pageSize)
		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}
	}

	// reponse data
	finalParsedResponse := gin.H{
		"success": true,
		"widgets": parsedWidgets,
	}

	// return final response
	c.JSON(http.StatusOK, finalParsedResponse)
}

// Exposed Method to Edit a Widget
func (handlers *FeedHandlers) EditWidget(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	widgetId := c.Param("widget_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var editWidgetRequest requests.EditWidgetRequest
	if err := c.ShouldBindJSON(&editWidgetRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// edit Widget
	widget, err := editWidget(handlers, widgetId, "", "", false, editWidgetRequest.MetaData, nil, communityId)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// Fetch Updated Widget Response
	widgetResponse := parseWidgetResponse(handlers, widget, communityId, headers[utils.HeadersMemberId])

	// reponse data
	response := gin.H{
		"success": true,
		"widget":  widgetResponse,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to Delete a Widget
func (handlers *FeedHandlers) DeleteWidget(c *gin.Context) {

	// fetch headers and url params
	widgetId := c.Param("widget_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// fetch widget using widget_id
	widget, err := fetchWidgetByID(handlers.widgetHelper, widgetId, false, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// delete widget using helper method
	err = handlers.widgetHelper.DeleteWidgetByIdHelper(widget.ID)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// delete widget data from elastic search
	err = handlers.esHelper.DeleteDocument(c, widget.ID.Hex(), constants.WidgetIndexName)
	if err != nil {
		fmt.Println(err.Error())
	}

	// reponse data
	response := gin.H{
		"success": true,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// method to delete widgets by Ids along with Indexed data
func deleteWidgetsByIds(widgetHelper interfaces.WidgetHelper, esHelper searchElastic.EsHelper, widgetIds []primitive.ObjectID,
) error {

	// Delete the documents from the collection
	filter := gin.H{
		"_id": gin.H{
			"$in": widgetIds,
		},
	}

	count, err := widgetHelper.DeleteWidgetsHelper(filter)
	if err != nil {
		return err
	}

	if int(count) != len(widgetIds) {
		logging.Error(fmt.Sprintf("Only %v Widgets deleted out of %v", count, widgetIds))
	}

	// Delete widget data from elastic search
	deleteWidgetsQuery := fmt.Sprintf(`{
		"query": {
			"terms": {
				"_id": %s
			}
		}
	}`, helpers.ParseObjectIdsToString(widgetIds))

	err = esHelper.DeleteByQuery(deleteWidgetsQuery, constants.WidgetIndexName)
	if err != nil {
		logging.Error(err.Error())
	}

	return nil
}
