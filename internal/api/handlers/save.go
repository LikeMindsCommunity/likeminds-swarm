package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to fetch list of posts saved by a User
func fetchPostIdsFromSave(saved_posts []entities.Save) []primitive.ObjectID {
	post_id_list := []primitive.ObjectID{}
	for _, save := range saved_posts {
		if save.EntityType == constants.PostEntityType {
			post_id_list = append(post_id_list, save.EntityId)
		}
	}

	return post_id_list
}

// Internal Method to fetch a Saved post of a User by post_id
func fetchUserSavedPostByPostId(helper interfaces.SaveHelper, post_id string,
	saved_by string) ([]entities.Save, error) {
	// save filter data
	save_filter_data := gin.H{
		"entity_id":   post_id,
		"entity_type": constants.PostEntityType,
		"saved_by":    saved_by,
		"is_deleted":  false,
	}

	// fetch save using helper method
	save_results, err := helper.FindSaveHelper(save_filter_data, gin.H{})
	if err != nil {
		return nil, err
	}

	return save_results, nil
}

// Internal Method to fetch the save status of a Post for a User
func fetchUserSavedStatusByPostId(helper interfaces.SaveHelper, post_id string, saved_by string) bool {

	save_results, err := fetchUserSavedPostByPostId(helper, post_id, saved_by)
	if err != nil {
		return false
	}

	if len(save_results) == 0 {
		return false
	}

	return true
}

func fetchUserSavedStatusByPostIds(helper interfaces.SaveHelper, postIds []primitive.ObjectID, savedBy string,
) map[primitive.ObjectID]bool {

	defer utils.Timer("fetchUserSavedStatusByPostIds")()

	userSavedStatusMap := make(map[primitive.ObjectID]bool, len(postIds))

	saveFilterData := gin.H{
		"entity_id": gin.H{
			"$in": postIds,
		},
		"entity_type": constants.PostEntityType,
		"saved_by":    savedBy,
		"is_deleted":  false,
	}

	saveResults, err := helper.FindSaveHelper(saveFilterData, gin.H{})
	if err != nil {
		logging.Error("Error while fetching saved posts", err)
		return userSavedStatusMap
	}

	for _, save := range saveResults {
		userSavedStatusMap[save.EntityId] = true
	}

	return userSavedStatusMap
}

// Exposed Method to Save a Post for a User
func (handlers *FeedHandlers) SavePost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")

	userId := headers[utils.HeadersMemberId]
	memberRole := headers[utils.HeadersMemberRole]

	isCm := utils.IsCMRole(memberRole)

	// validation of api_key
	community_id := externalHelpers.GetCommunityId(c)
	if community_id == externalHelpers.DefaultCommunityId {
		return
	}

	// fetch post using helper method
	postData, err := FetchPostData(handlers.postHelper, post_id, community_id, true, []string{})
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// If post is hidden and user is not cm or creator, then throw error
	if !isCm && postData.IsHidden && userId != postData.UserId {
		utils.GeneralAPIValidationError(c, utils.PostIsHiddenError)
		return
	}

	// fetch save using helper method
	save_results, err := fetchUserSavedPostByPostId(handlers.saveHelper, post_id,
		headers[utils.HeadersMemberId])
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	if len(save_results) == 0 {
		// create save using the helper method
		_, err := handlers.saveHelper.CreateSaveHelper(constants.PostEntityType, postData.ID,
			headers[utils.HeadersMemberId], community_id)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	} else {
		save_data := save_results[0]

		// save update data
		save_update_data := gin.H{
			"$set": gin.H{
				"is_deleted": !save_data.IsDeleted,
			},
		}

		// update save using the helper method
		err = handlers.saveHelper.UpdateSaveByIdHelper(save_data.ID, save_update_data)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// Exposed Method to fetch Posts saved by a User
func (handlers *FeedHandlers) FetchUserSavedPosts(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	userId := c.Param("user_id")
	paramIsCm := c.Query("user_is_cm")
	isCm := false

	apiRevampV1Check := utils.ApiRevampCheckV1(headers[utils.HeadersAcceptVersion])
	versionCode := headers[utils.HeadersAcceptVersion]
	platformCode := headers[utils.HeadersPlatformCode]

	if paramIsCm == "true" {
		isCm = true
	}

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	if userId != headers[utils.HeadersMemberId] {
		utils.GeneralAPIValidationError(c, "You are not authorized to perform this operation.")
		return
	}

	// save filter data
	saveFilterData := gin.H{
		"entity_type":  constants.PostEntityType,
		"saved_by":     headers[utils.HeadersMemberId],
		"community_id": communityId,
		"is_deleted":   false,
	}

	// fetch save count using helper method
	saveCount, err := handlers.saveHelper.CountSaveHelper(saveFilterData)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// filter options
	saveFilterOptions, err := generatePageFilterOptions(c, "", OrderTypeDefault)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch save using helper method
	saveResults, err := handlers.saveHelper.FindSaveHelper(saveFilterData, saveFilterOptions)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// posts filter data
	postFilterData := gin.H{
		"_id": gin.H{
			"$in": fetchPostIdsFromSave(saveResults),
		},
		"is_deleted": false,
	}

	// fetch posts using helper method
	postResults, err := handlers.postHelper.FindPostHelper(postFilterData, saveFilterOptions)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	parsedPosts := parseMultiplePostResponse(handlers, postResults, userId, isCm,
		headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check, utils.DefaultRole)

	// final response data
	finalResponse := gin.H{
		"success":     true,
		"posts":       parsedPosts,
		"total_count": saveCount,
	}

	finalResponse["topics"] = getTopicDataFromPosts(handlers.topicHelper, finalResponse, communityId)
	finalResponse["reposted_posts"] = getOriginalPostForReposts(handlers, finalResponse, communityId, headers[utils.HeadersMemberId], isCm, headers[utils.HeadersVersionCode], headers[utils.HeadersPlatformCode], apiRevampV1Check)
	finalResponse["widgets"] = getWidgetDataFromFeedResponse(handlers, finalResponse, communityId, isCm, headers[utils.HeadersMemberId])

	// Get community configurations
	universalFeedConfig := externalHelpers.GetUniversalFeedConfigurationsData(handlers.cacheHelper, userId, communityId)

	var commentSortOrderVal int
	filtered_comments := map[string]responses.CommentWithParentResponse{}

	if universalFeedConfig.CommentSortOrder == enums.DescendingSortOrder {
		commentSortOrderVal = -1
	} else {
		commentSortOrderVal = 1
	}

	if universalFeedConfig.CommentSortOn == enums.UniversalFeedTopLikedComments {
		var updatedPostsWithComments []responses.PostResponse
		updatedPostsWithComments, filtered_comments, err = getTopCommentsAgainstPostsSortOnLikes(handlers,
			parsedPosts, userId, isCm, communityId, commentSortOrderVal,
			universalFeedConfig.CommentCount, versionCode, platformCode, apiRevampV1Check, utils.DefaultRole)

		if err != nil {
			utils.GeneralAPIValidationError(c, err.Error())
			return
		}

		if len(updatedPostsWithComments) > 0 {
			finalResponse["posts"] = updatedPostsWithComments
		}

	}

	finalResponse["filtered_comments"] = filtered_comments

	// return final response
	c.JSON(http.StatusOK, finalResponse)
}
