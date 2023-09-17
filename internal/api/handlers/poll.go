package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

// Internal Method to fetch poll votes using uuid id and poll_id
func fetchPollVoteByUUID(helper interfaces.PollVotesHelper, pollId string, uuid string,
	communityId int) ([]entities.PollVotes, error) {
	// pollVotes filter data
	pollVotesFilterData := gin.H{
		"poll_id":      pollId,
		"community_id": communityId,
		"uuid":         uuid,
	}

	// fetch pollVotes using helper method
	pollVotesResults, err := helper.FindPollVotesHelper(pollVotesFilterData, gin.H{})
	if err != nil {
		return nil, err
	}

	return pollVotesResults, nil
}

// Exposed Method to Add a New Poll Option
func (handlers *FeedHandlers) AddPollOption(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	pollId := c.Param("poll_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var createPollOptionRequest requests.CreatePollOptionRequest
	if err := c.ShouldBindJSON(&createPollOptionRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Fetch poll widget using poll Id
	pollWidget, err := fetchWidgetByID(handlers.widgetHelper, pollId, true, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, "Invalid poll_id sent")
		return
	}

	// Check if allow_add_option key exists
	allowAddOption, exists := pollWidget.MetaData["allow_add_option"]
	if !exists {
		utils.GeneralAPIValidationError(c, "Invalid poll_id sent")
		return
	}

	if !allowAddOption.(bool) {
		utils.GeneralAPIValidationError(c, "Option can't be added")
		return
	}

	lmMeta := map[string]interface{}{}

	// Fetch existing LM Meta Data
	if pollWidget.LMMeta != nil {
		lmMeta = pollWidget.LMMeta
	}

	if _, exists := lmMeta["options"]; !exists {
		lmMeta["options"] = []interface{}{}
	}

	// Generating poll option
	pollOptionId, err := uuid.NewRandom()
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	pollOption := gin.H{
		"_id":  pollOptionId.String(),
		"text": createPollOptionRequest.Text,
		"uuid": headers[utils.HeadersMemberId],
	}

	lmMeta["options"] = append(lmMeta["options"].([]interface{}), pollOption)

	// update widget from given metadata
	pollWidget, ok := editWidget(c, handlers, pollId, true, pollWidget.MetaData, lmMeta, communityId)
	if !ok {
		return
	}

	widgetResponse := parseWidgetResponse(pollWidget)

	// response data
	response := gin.H{
		"success": true,
		"widget":  widgetResponse,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to Vote on a Poll
func (handlers *FeedHandlers) VoteOnPoll(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	pollId := c.Param("poll_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var createVoteOnPollRequest requests.CreateVoteOnPollRequest
	if err := c.ShouldBindJSON(&createVoteOnPollRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	if len(createVoteOnPollRequest.Votes) < 1 {
		utils.GeneralAPIValidationError(c, "Invalid votes sent")
		return
	}

	// Fetch poll widget using poll Id
	pollWidget, err := fetchWidgetByID(handlers.widgetHelper, pollId, true, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, "Invalid poll_id sent")
		return
	}

	// Check if expiry_time key exists
	expiryTime, exists := pollWidget.MetaData["expiry_time"]
	if !exists {
		utils.GeneralAPIValidationError(c, "Invalid poll_id sent")
		return
	}

	if expiryTime.(int64) <= int64(time.Now().UnixMilli()) {
		utils.GeneralAPIValidationError(c, "Poll Expired")
		return
	}

	// Check if poll_type key exists
	pollType, exists := pollWidget.MetaData["poll_type"]
	if !exists {
		utils.GeneralAPIValidationError(c, "Invalid poll_id sent")
		return
	}

	// Fetch votes of the user, if exists
	pollVotes, err := fetchPollVoteByUUID(handlers.pollVotesHelper, pollId,
		headers[utils.HeadersMemberId], communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Check if user is trying to vote again on instant poll
	if pollType == enums.InstantPollType && len(pollVotes) > 0 {
		utils.GeneralAPIValidationError(c, "Can't Vote again")
		return
	}

	// Check if multiple_select_state and multiple_select_number key exists
	multipleSelectState, exists := pollWidget.MetaData["multiple_select_state"]
	if !exists {
		utils.GeneralAPIValidationError(c, "Invalid poll_id sent")
		return
	}

	multipleSelectNumber, exists := pollWidget.MetaData["multiple_select_number"]
	if !exists {
		utils.GeneralAPIValidationError(c, "Invalid poll_id sent")
		return
	}

	votesLength := len(createVoteOnPollRequest.Votes)

	// Check if invalid number of options are selected while voting
	if (multipleSelectState == enums.ExactlySelectStateType && multipleSelectNumber.(int) != votesLength) ||
		(multipleSelectState == enums.AtMaxSelectStateType && multipleSelectNumber.(int) < votesLength) ||
		(multipleSelectState == enums.AtLeastSelectStateType && multipleSelectNumber.(int) > votesLength) {
		utils.GeneralAPIValidationError(c, "Invalid number of options selected")
		return
	}

	// Check if options are present in poll
	if pollWidget.LMMeta == nil {
		utils.GeneralAPIValidationError(c, "Invalid votes sent")
		return
	}

	// Check if options are present in poll
	pollOptions, exists := pollWidget.LMMeta["options"]
	if !exists {
		utils.GeneralAPIValidationError(c, "Invalid votes sent")
		return
	}

	// process option Ids
	optionIdsMap := map[string]bool{}
	for _, option := range pollOptions.([]interface{}) {
		optionIdsMap[option.(map[string]string)["_id"]] = true
	}

	// Check for valid vote Ids
	for _, voteId := range createVoteOnPollRequest.Votes {
		if !optionIdsMap[voteId] {
			utils.GeneralAPIValidationError(c, "Invalid votes sent")
			return
		}
	}

	if len(pollVotes) == 0 {
		// Create a new PollVotes Instance
		_, err := handlers.pollVotesHelper.CreatePollVotesHelper(pollWidget.ID,
			headers[utils.HeadersMemberId], createVoteOnPollRequest.Votes, communityId)
		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}
	} else {
		// Update the existing PollVotes Instance
		pollVote := pollVotes[0]

		updateBody := gin.H{
			"$set": gin.H{
				"votes": createVoteOnPollRequest.Votes,
			},
		}

		err := handlers.pollVotesHelper.UpdatePollVotesByIdHelper(pollVote.ID, updateBody)
		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}
	}

	// response data
	response := gin.H{
		"success": true,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// Exposed Method to Fetch Votes on a Poll
func (handlers *FeedHandlers) GetPollVotes(c *gin.Context) {
	// fetch headers and url params
	// headers := utils.GetHeaders(c)
	pollId := c.Param("poll_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var getPollVotesRequest requests.GetPollVotesRequest
	err := c.BindQuery(&getPollVotesRequest)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// Parse votes Ids string array
	voteIds := parseStringArrayParam(getPollVotesRequest.Votes)

	// Fetch poll widget using poll Id
	pollWidget, err := fetchWidgetByID(handlers.widgetHelper, pollId, true, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, "Invalid poll_id sent")
		return
	}

	// Check if is_anonymous key exists
	isAnonymous, exists := pollWidget.MetaData["is_anonymous"]
	if !exists {
		utils.GeneralAPIValidationError(c, "Invalid poll_id sent")
		return
	}

	// response data
	response := gin.H{
		"success": true,
	}

	if isAnonymous.(bool) {
		response["votes"] = []interface{}{}
	} else {
		// Check if options are present in poll
		if pollWidget.LMMeta == nil {
			utils.GeneralAPIValidationError(c, "Invalid poll_id sent")
			return
		}

		// Check if options are present in poll
		pollOptions, exists := pollWidget.LMMeta["options"]
		if !exists {
			utils.GeneralAPIValidationError(c, "Invalid poll_id sent")
			return
		}

		// process option Ids
		optionIdsMap := map[string]bool{}
		for _, option := range pollOptions.([]interface{}) {
			optionIdsMap[option.(map[string]string)["_id"]] = true
		}

		// Check for valid vote Ids
		for _, voteId := range voteIds {
			if !optionIdsMap[voteId] {
				utils.GeneralAPIValidationError(c, "Invalid votes sent")
				return
			}
		}

		// Run the aggregate query here
	}

	// return final response
	c.JSON(http.StatusOK, response)
}
