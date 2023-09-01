package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	log "github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to parse custom Widget for response
func parseCustomWidgetResponse(customWidget *entities.CustomWidget) requests.CustomWidgetResponse {
	var response requests.CustomWidgetResponse

	response.ID = customWidget.ID
	response.ParentEntityID = customWidget.ParentEntityID
	response.ParentEntityType = customWidget.ParentEntityType
	response.MetaData = customWidget.MetaData

	response.CreatedAt = int(customWidget.CreatedAt.UnixMilli())
	response.UpdatedAt = int(customWidget.UpdatedAt.UnixMilli())
	return response
}

// Internal Method to fetch custom widget using custom widget id and communityId
func fetchCustomWidgetByID(helper interfaces.CustomWidgetHelper, customWidgetId string, communityId int, isLMCreated bool) (*entities.CustomWidget, error) {
	// custom widget filter data
	customWidgetFilterData := gin.H{
		"_id":           customWidgetId,
		"community_id":  communityId,
		"created_by_lm": isLMCreated,
	}

	// fetch custom widget using helper method
	customWidgetResults, err := helper.FindCustomWidgetHelper(customWidgetFilterData, gin.H{})
	if err != nil {
		return nil, err
	}

	// validation of customWidgetId
	if len(customWidgetResults) == 0 {
		return nil, fmt.Errorf("invalid widget id sent")
	}

	return &customWidgetResults[0], nil
}

// Exposed Method to Create a Custom Widget
func (handlers *FeedHandlers) CreateCustomWidget(c *gin.Context) {
	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var createCustomWidgetRequest requests.CreateCustomWidgetRequest
	if err := c.ShouldBindJSON(&createCustomWidgetRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// create custom Widget using the helper method
	customWidgetId, err := handlers.customWidgetHelper.CreateCustomWidgetHelper(false, createCustomWidgetRequest.ParentEntityID,
		createCustomWidgetRequest.ParentEntityType, createCustomWidgetRequest.MetaData, communityId)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	customWidgetData, err := fetchCustomWidgetByID(handlers.customWidgetHelper, customWidgetId.(primitive.ObjectID).Hex(), communityId, false)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// insert custom widget data in elastic search
	err = handlers.esHelper.InsertDocument(c, ParseCustomWidgetIndexData(customWidgetData), customWidgetData.ID.Hex(),
		constants.CustomWidgetIndexName)
	if err != nil {
		log.Error(err.Error())
	}

	customWidgetResponse := parseCustomWidgetResponse(customWidgetData)

	// response data
	response := gin.H{
		"success": true,
		"widget":  customWidgetResponse,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

func processCustomWidgetSearchData(handlers *FeedHandlers, data map[string]interface{}) []requests.CustomWidgetResponse {
	customWidgetDetails := data["hits"].(map[string]interface{})["hits"].([]interface{})
	var customWidgetList []entities.CustomWidget

	for _, data := range customWidgetDetails {
		customWidgetData := data.(map[string]interface{})["_source"].(map[string]interface{})
		customWidgetData["_id"] = customWidgetData["id"]

		// convert the data to custom widget entity
		var customWidget entities.CustomWidget
		b, _ := json.Marshal(customWidgetData)
		json.Unmarshal(b, &customWidget)

		customWidgetList = append(customWidgetList, customWidget)
	}

	customWidgetResponse := []requests.CustomWidgetResponse{}

	// Parse all fetched custom Widget Data
	for _, customWidget := range customWidgetList {
		customWidgetResponse = append(customWidgetResponse, parseCustomWidgetResponse(&customWidget))
	}

	return customWidgetResponse
}

// Exposed Method to Fetch CustomWidgets based on given params
func (handlers *FeedHandlers) FetchCustomWidget(c *gin.Context) {
	// parse fetch custom Widget request
	var fetchCustomWidgetRequest requests.FetchCustomWidgetRequest
	err := c.BindQuery(&fetchCustomWidgetRequest)
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
	customWidgetQuery := GetCustomWidgetFilterQuery(page, pageSize, communityId, fetchCustomWidgetRequest.SearchKey,
		fetchCustomWidgetRequest.SearchValue)

	// Check for JSON errors
	isValid := json.Valid([]byte(customWidgetQuery)) // returns bool
	if !isValid {
		utils.GeneralAPIValidationError(c, "Invalid search params sent")
		return
	}

	response := handlers.esHelper.ExecuteQuery(customWidgetQuery, constants.CustomWidgetIndexName)

	finalResponse := processCustomWidgetSearchData(handlers, response)

	// reponse data
	finalParsedResponse := gin.H{
		"success": true,
		"widgets": finalResponse,
	}

	// return final response
	c.JSON(http.StatusOK, finalParsedResponse)
}

// Exposed Method to Edit a Custom Widget
func (handlers *FeedHandlers) EditCustomWidget(c *gin.Context) {
	// fetch url params
	widgetId := c.Param("widget_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var editCustomWidgetRequest requests.EditCustomWidgetRequest
	if err := c.ShouldBindJSON(&editCustomWidgetRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch custom Widget using widget_id
	customWidget, err := fetchCustomWidgetByID(handlers.customWidgetHelper, widgetId, communityId, false)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// custom widget update data
	customWidgetUpdateData := gin.H{
		"$set": gin.H{},
	}

	// Update set object with name field, if changed
	if editCustomWidgetRequest.MetaData != nil {
		customWidgetUpdateData["$set"].(gin.H)["metadata"] = editCustomWidgetRequest.MetaData
	}

	// Validation of data change
	if len(customWidgetUpdateData["$set"].(gin.H)) > 0 {
		// update custom widget using the helper method
		err = handlers.customWidgetHelper.UpdateCustomWidgetByIdHelper(customWidget.ID, customWidgetUpdateData)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}

		customWidget, err = fetchCustomWidgetByID(handlers.customWidgetHelper, customWidget.ID.Hex(), communityId, false)
		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}

		// update cusotm Widget data in elastic search
		err = handlers.esHelper.UpdateDocument(c, ParseCustomWidgetIndexData(customWidget), customWidget.ID.Hex(),
			constants.CustomWidgetIndexName)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	// Fetch Updated custom Widget Response
	customWidgetResponse := parseCustomWidgetResponse(customWidget)

	// reponse data
	response := gin.H{
		"success": true,
		"widget":  customWidgetResponse,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}
