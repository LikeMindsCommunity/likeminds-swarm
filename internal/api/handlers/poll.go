package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
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

// Internal Method to create a poll option object using text
func createPollOptionObjects(optionTexts []string, userUUID string) ([]interface{}, error) {
	pollOptions := []interface{}{}

	for _, optionText := range optionTexts {
		// Generating poll option
		pollOptionId, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}

		pollOption := gin.H{
			"_id":  pollOptionId.String(),
			"text": optionText,
			"uuid": userUUID,
		}

		pollOptions = append(pollOptions, pollOption)
	}

	return pollOptions, nil
}

// Exposed Method to Add a New Poll Option
func (handlers *FeedHandlers) AddPollOption(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	pollId := c.Param("poll_id")

	isCM := utils.IsCMRole(headers[utils.HeaderMemberRole])

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

	// Generating poll options
	pollOptionObjects, err := createPollOptionObjects([]string{createPollOptionRequest.Text},
		headers[utils.HeadersMemberId])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	lmMetaOptions = append(lmMetaOptions, pollOptionObjects...)
	lmMeta["options"] = lmMetaOptions

	// update widget from given metadata
	pollWidget, err = editWidget(handlers, pollId, "", "", true, pollWidget.MetaData, lmMeta, communityId)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	widgetResponse := parseWidgetResponse(handlers, pollWidget, communityId, isCM, headers[utils.HeadersMemberId])

	// response data
	response := gin.H{
		"success": true,
		"widget":  widgetResponse,
	}

	// return final response
	c.JSON(http.StatusOK, response)
}

// Internal Method to perform validation on pollWidget
func pollWidgetValidation(handlers *FeedHandlers, pollId string, communityId int) (*entities.Widget, error) {
	// Fetch poll widget using poll Id
	pollWidget, err := fetchWidgetByID(handlers.widgetHelper, pollId, true, communityId)
	if err != nil {
		return nil, fmt.Errorf("invalid poll_id sent")
	}

	// Check if expiry_time key exists
	expiryTime, exists := pollWidget.MetaData["expiry_time"]
	if !exists {
		return nil, fmt.Errorf("invalid poll_id sent")
	}

	if expiryTime.(float64) <= float64(time.Now().UnixMilli()) {
		return nil, fmt.Errorf("poll expired")
	}

	// Check if poll_type key exists
	_, exists = pollWidget.MetaData["poll_type"]
	if !exists {
		return nil, fmt.Errorf("invalid poll_id sent")
	}

	return pollWidget, nil
}

// Internal Method to perform validation on VoteOnPoll
func voteOnPollValidation(createVoteOnPollRequest requests.CreateVoteOnPollRequest, handlers *FeedHandlers,
	pollId string, communityId int, uuid string) (*entities.Widget, []entities.PollVotes, error) {
	if len(createVoteOnPollRequest.Votes) < 1 {
		return nil, nil, fmt.Errorf("invalid votes sent")
	}

	// perform poll widget validation
	pollWidget, err := pollWidgetValidation(handlers, pollId, communityId)
	if err != nil {
		return nil, nil, err
	}

	// Fetch votes of the user, if exists
	pollVotes, err := fetchPollVoteByUUID(handlers.pollVotesHelper, pollId,
		uuid, communityId)
	if err != nil {
		return pollWidget, nil, err
	}

	// Check if user is trying to vote again on instant poll
	if pollWidget.MetaData["poll_type"] == enums.InstantPollType && len(pollVotes) > 0 {
		return pollWidget, pollVotes, fmt.Errorf("can't Vote again")
	}

	// Check if multiple_select_state and multiple_select_number key exists
	multipleSelectState, exists := pollWidget.MetaData["multiple_select_state"]
	if !exists {
		return pollWidget, pollVotes, fmt.Errorf("invalid poll_id sent")
	}

	multipleSelectNumber, exists := pollWidget.MetaData["multiple_select_number"]
	if !exists {
		return pollWidget, pollVotes, fmt.Errorf("invalid poll_id sent")
	}

	votesLength := len(createVoteOnPollRequest.Votes)

	finalMultipleSelectNumber, ok := multipleSelectNumber.(int32)

	if !ok {
		finalMultipleSelectNumber = int32(multipleSelectNumber.(float64))
	}

	// Check if invalid number of options are selected while voting
	if (multipleSelectState == enums.ExactlySelectStateType && finalMultipleSelectNumber != int32(votesLength)) ||
		(multipleSelectState == enums.AtMaxSelectStateType && finalMultipleSelectNumber < int32(votesLength)) ||
		(multipleSelectState == enums.AtLeastSelectStateType && finalMultipleSelectNumber > int32(votesLength)) {
		return pollWidget, pollVotes, fmt.Errorf("invalid number of options selected")
	}

	// Check if options are present in poll
	if pollWidget.LMMeta == nil {
		return pollWidget, pollVotes, fmt.Errorf("invalid votes sent")
	}

	// Check if options are present in poll
	pollOptions, exists := pollWidget.LMMeta["options"]
	if !exists {
		return pollWidget, pollVotes, fmt.Errorf("invalid votes sent")
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
			return pollWidget, pollVotes, fmt.Errorf("invalid votes sent")
		}
	}

	return pollWidget, pollVotes, nil
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

	pollWidget, pollVotes, err := voteOnPollValidation(createVoteOnPollRequest, handlers,
		pollId, communityId, headers[utils.HeadersMemberId])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
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

// Internal Method to fetch a users vote on a poll
func GetPollVoteOfUUID(handlers *FeedHandlers, pollId string, communityId int, uuid string) (*entities.PollVotes, error) {
	pollVoteFilterData := gin.H{
		"poll_id":      pollId,
		"community_id": communityId,
		"votes.0": gin.H{
			"$exists": true,
		},
		"uuid": uuid,
	}

	pollVotes, err := handlers.pollVotesHelper.FindPollVotesHelper(pollVoteFilterData, gin.H{})
	if err != nil {
		return nil, err
	}

	// validation of post_id
	if len(pollVotes) == 0 {
		return nil, fmt.Errorf("invalid poll_id sent")
	}

	return &pollVotes[0], nil
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
func getPollVotesDataUsingAggregation(handlers *FeedHandlers, pollId string, communityId int, uniqueVotersOnPoll int64, uuid string) ([]gin.H, error) {
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
				"$round": []interface{}{
					bson.M{
						"$multiply": []interface{}{
							bson.M{
								"$divide": []interface{}{"$vote_count", uniqueVotersOnPoll},
							}, 100,
						},
					}, 2,
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
	optionIds []string, page int, page_size int) ([]gin.H, error) {
	// Run the aggregate query here
	pollVotesFilterData := []map[string]interface{}{}

	pollIdObject := helpers.ConvertIdsToObjectIds([]string{pollId})

	if len(pollIdObject) == 0 {
		return nil, errors.New("Invalid poll id")
	}

	// Add match logic
	pollVotesFilterData = append(pollVotesFilterData, bson.M{
		"$match": bson.M{
			"poll_id":      pollIdObject[0],
			"community_id": communityId,
		},
	})

	if len(optionIds) > 0 {
		// Add match logic
		pollVotesFilterData = append(pollVotesFilterData, bson.M{
			"$match": bson.M{
				"votes": bson.M{
					"$exists": true,
					"$in":     optionIds,
				},
			},
		})
	}

	// Add project logic
	pollVotesFilterData = append(pollVotesFilterData, bson.M{
		"$project": bson.M{
			"_id":        0,
			"uuid":       1,
			"votes":      1,
			"created_at": 1,
			"updated_at": 1,
		},
	})

	// Add unwind logic
	pollVotesFilterData = append(pollVotesFilterData, bson.M{
		"$unwind": bson.M{
			"path": "$votes",
		},
	})

	if len(optionIds) > 0 {
		// Add match logic
		pollVotesFilterData = append(pollVotesFilterData, bson.M{
			"$match": bson.M{
				"votes": bson.M{
					"$exists": true,
					"$in":     optionIds,
				},
			},
		})
	}

	// Add sort logic
	pollVotesFilterData = append(pollVotesFilterData, bson.M{
		"$sort": gin.H{
			"updated_at": -1,
			"created_at": 1,
		},
	})

	// Add skip logic
	pollVotesFilterData = append(pollVotesFilterData, gin.H{
		"$skip": page_size * (page - 1),
	})
	// Add limit logic
	pollVotesFilterData = append(pollVotesFilterData, gin.H{
		"$limit": page_size,
	})

	// Add group logic
	pollVotesFilterData = append(pollVotesFilterData, bson.M{
		"$group": bson.M{
			"_id": "$votes",
			"users": bson.M{
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
	voteIds := utils.ParseStringArrayParam(getPollVotesRequest.Votes)

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

	// fetch pagination query params
	page, page_size, err := fetchPaginationParams(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
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
		pollVotesResults, err := getPollVotesUsingAggregation(handlers, pollId, communityId, voteIds, page, page_size)
		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}

		response["votes"] = pollVotesResults
	}

	// return final response
	c.JSON(http.StatusOK, response)
}
