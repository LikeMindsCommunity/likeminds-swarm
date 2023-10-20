package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	log "github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to create a new widget with indexing
func createWidget(c *gin.Context, handlers *FeedHandlers, createdByLM bool, parentEntityID string, parentEntityType string,
	metaData map[string]interface{}, lmMeta map[string]interface{}, communityId int) (*entities.Widget, bool) {
	// create Widget using the helper method
	widgetId, err := handlers.widgetHelper.CreateWidgetHelper(createdByLM, parentEntityID, parentEntityType, metaData, lmMeta,
		communityId)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return nil, false
	}

	widgetData, err := fetchWidgetByID(handlers.widgetHelper, widgetId.(primitive.ObjectID).Hex(), createdByLM, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return nil, false
	}

	// insert widget data in elastic search
	err = handlers.esHelper.InsertDocument(c, ParseWidgetIndexData(widgetData), widgetData.ID.Hex(),
		constants.WidgetIndexName)
	if err != nil {
		log.Error(err.Error())
	}

	return widgetData, true
}

// Internal Method to edit an existing widget with index update
func editWidget(c *gin.Context, handlers *FeedHandlers, widgetId string, createdByLM bool, metaData map[string]interface{},
	lmMeta map[string]interface{}, communityId int) (*entities.Widget, bool) {
	// fetch Widget using widget_id
	widget, err := fetchWidgetByID(handlers.widgetHelper, widgetId, createdByLM, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return widget, false
	}

	// widget update data
	widgetUpdateData := gin.H{
		"$set": gin.H{},
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
			utils.GeneralAPIInternalError(c, err.Error())
			return nil, false
		}

		widget, err = fetchWidgetByID(handlers.widgetHelper, widget.ID.Hex(), createdByLM, communityId)
		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return nil, false
		}

		// update Widget data in elastic search
		err = handlers.esHelper.UpdateDocument(c, ParseWidgetIndexData(widget), widget.ID.Hex(),
			constants.WidgetIndexName)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	return widget, true
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
	widgetData, ok := createWidget(c, handlers, false, createWidgetRequest.ParentEntityID, createWidgetRequest.ParentEntityType,
		createWidgetRequest.MetaData, nil, communityId)
	if !ok {
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

	// dsl query to search topics
	widgetQuery := GetWidgetFilterQuery(page, pageSize, communityId, fetchWidgetRequest.SearchKey,
		fetchWidgetRequest.SearchValue)

	// Check for JSON errors
	isValid := json.Valid([]byte(widgetQuery)) // returns bool
	if !isValid {
		utils.GeneralAPIValidationError(c, "Invalid search params sent")
		return
	}

	response := handlers.esHelper.ExecuteQuery(widgetQuery, constants.WidgetIndexName)

	finalResponse := processWidgetSearchData(handlers, response, communityId, headers[utils.HeadersMemberId])

	// reponse data
	finalParsedResponse := gin.H{
		"success": true,
		"widgets": finalResponse,
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
	widget, ok := editWidget(c, handlers, widgetId, false, editWidgetRequest.MetaData, nil, communityId)
	if !ok {
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
