package handlers

import (
	"encoding/json"
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
	"go.mongodb.org/mongo-driver/bson"
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
	lmMetaOptions := []interface{}{}

	// Fetch existing LM Meta Data
	if pollWidget.LMMeta != nil {
		lmMeta = pollWidget.LMMeta
	}

	if _, exists := lmMeta["options"]; exists {
		// option data conversion to desired type
		options := []interface{}{}
		convertedOptions, _ := json.Marshal(lmMeta["options"])
		_ = json.Unmarshal(convertedOptions, &options)
		lmMetaOptions = options
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

	lmMetaOptions = append(lmMetaOptions, pollOption)
	lmMeta["options"] = lmMetaOptions

	// update widget from given metadata
	pollWidget, ok := editWidget(c, handlers, pollId, true, pollWidget.MetaData, lmMeta, communityId)
	if !ok {
		return
	}

	widgetResponse := parseWidgetResponse(handlers, pollWidget, communityId, headers[utils.HeadersMemberId])

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

	if expiryTime.(float64) <= float64(time.Now().UnixMilli()) {
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
	if (multipleSelectState == enums.ExactlySelectStateType && multipleSelectNumber.(int32) != int32(votesLength)) ||
		(multipleSelectState == enums.AtMaxSelectStateType && multipleSelectNumber.(float64) < float64(votesLength)) ||
		(multipleSelectState == enums.AtLeastSelectStateType && multipleSelectNumber.(float64) > float64(votesLength)) {
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

	// option data conversion to desired type
	options := []map[string]interface{}{}
	convertedOptions, _ := json.Marshal(pollOptions)
	_ = json.Unmarshal(convertedOptions, &options)

	// process option Ids
	optionIdsMap := map[string]bool{}
	for _, option := range options {
		optionIdsMap[option["_id"].(string)] = true
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

// Internal Method to fetch unique users on a Poll
func getUniqueVotersOnPoll(handlers *FeedHandlers, pollId string, communityId int) (int64, error) {
	pollUniqueVotersFilterData := gin.H{
		"poll_id":      pollId,
		"community_id": communityId,
		"votes.0": gin.H{
			"$exists": true,
		},
	}

	uniqueVotersCount, err := handlers.pollVotesHelper.CountPollVotesHelper(pollUniqueVotersFilterData)
	if err != nil {
		return 0, err
	}

	return uniqueVotersCount, nil
}

// Internal Method to fetch poll votes data using aggregation
func getPollVotesDataUsingAggregation(handlers *FeedHandlers, pollId string, communityId int, uuid string) ([]gin.H, error) {
	uniqueVotersOnPoll, err := getUniqueVotersOnPoll(handlers, pollId, communityId)
	if err != nil {
		return nil, err
	}

	pollVotesDataFilterData := []map[string]interface{}{}

	// Add match logic
	pollVotesDataFilterData = append(pollVotesDataFilterData, bson.M{
		"$match": bson.M{
			"poll_id":      pollId,
			"community_id": communityId,
		},
	})

	// Add project logic
	pollVotesDataFilterData = append(pollVotesDataFilterData, bson.M{
		"$project": bson.M{
			"_id":   0,
			"uuid":  1,
			"votes": 1,
		},
	})

	// Add unwind logic
	pollVotesDataFilterData = append(pollVotesDataFilterData, bson.M{
		"$unwind": bson.M{
			"path": "$votes",
		},
	})

	// Add group logic
	pollVotesDataFilterData = append(pollVotesDataFilterData, bson.M{
		"$group": bson.M{
			"_id": "$votes",
			"users": bson.M{
				"$addToSet": "$uuid",
			},
		},
	})

	// Add fields logic
	pollVotesDataFilterData = append(pollVotesDataFilterData, bson.M{
		"$addFields": bson.M{
			"vote_count": bson.M{
				"$size": "$users",
			},
			"is_selected": bson.M{
				"$cond": []interface{}{
					bson.M{
						"$gt": []interface{}{
							bson.M{
								"$size": gin.H{
									"$setIntersection": []interface{}{"$users", []string{uuid}},
								},
							}, 0,
						},
					}, true, false,
				},
			},
		},
	})

	// Add fields logic
	pollVotesDataFilterData = append(pollVotesDataFilterData, bson.M{
		"$addFields": bson.M{
			"percentage": bson.M{
				"$multiply": []interface{}{
					bson.M{
						"$divide": []interface{}{"$vote_count", uniqueVotersOnPoll},
					}, 100,
				},
			},
		},
	})

	// Add project logic
	pollVotesDataFilterData = append(pollVotesDataFilterData, bson.M{
		"$project": bson.M{
			"_id":         1,
			"vote_count":  1,
			"is_selected": 1,
			"percentage":  1,
		},
	})

	// fetch pollVotes using helper method
	pollVotesResults, err := handlers.pollVotesHelper.AggregatePollVotesHelper(pollVotesDataFilterData)
	if err != nil {
		return nil, err
	}

	return pollVotesResults, nil
}

// Internal Method to Fetch Votes on a Poll using aggregation
func getPollVotesUsingAggregation(handlers *FeedHandlers, pollId string, communityId int,
	optionIds []string) ([]gin.H, error) {
	// Run the aggregate query here
	pollVotesFilterData := []map[string]interface{}{}

	// Add match logic
	pollVotesFilterData = append(pollVotesFilterData, gin.H{
		"$match": gin.H{
			"poll_id":      pollId,
			"community_id": communityId,
		},
	})

	if len(optionIds) > 0 {
		// Add match logic
		pollVotesFilterData = append(pollVotesFilterData, gin.H{
			"$match": gin.H{
				"votes": gin.H{
					"$exists": true,
					"$in":     optionIds,
				},
			},
		})
	}

	// Add project logic
	pollVotesFilterData = append(pollVotesFilterData, gin.H{
		"$project": gin.H{
			"_id":   0,
			"uuid":  1,
			"votes": 1,
		},
	})

	// Add unwind logic
	pollVotesFilterData = append(pollVotesFilterData, gin.H{
		"$unwind": gin.H{
			"path": "$votes",
		},
	})

	if len(optionIds) > 0 {
		// Add match logic
		pollVotesFilterData = append(pollVotesFilterData, gin.H{
			"$match": gin.H{
				"votes": gin.H{
					"$exists": true,
					"$in":     optionIds,
				},
			},
		})
	}

	// Add group logic
	pollVotesFilterData = append(pollVotesFilterData, gin.H{
		"$group": gin.H{
			"_id": "$votes",
			"users": gin.H{
				"$addToSet": "$uuid",
			},
		},
	})

	// fetch pollVotes using helper method
	pollVotesResults, err := handlers.pollVotesHelper.AggregatePollVotesHelper(pollVotesFilterData)
	if err != nil {
		return nil, err
	}

	return pollVotesResults, nil
}

// Exposed Method to Fetch Votes on a Poll
func (handlers *FeedHandlers) GetPollVotes(c *gin.Context) {
	// fetch headers and url params
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
		_, exists := pollWidget.LMMeta["options"]
		if !exists {
			utils.GeneralAPIValidationError(c, "Invalid poll_id sent")
			return
		}

		// fetch pollVotes using internal method
		pollVotesResults, err := getPollVotesUsingAggregation(handlers, pollId, communityId, voteIds)
		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}

		response["votes"] = pollVotesResults
	}

	// return final response
	c.JSON(http.StatusOK, response)
}
