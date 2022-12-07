package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/interfaces"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

func parseLikeResponse(like entities.Like) requests.LikeResponse {
	var response requests.LikeResponse

	response.ID = like.ID
	response.UserId = like.LikedBy
	response.CreatedAt = int(like.CreatedAt.UnixMilli())
	response.UpdatedAt = int(like.UpdatedAt.UnixMilli())

	return response
}

func parseMultipleLikeResponse(likes []entities.Like) []requests.LikeResponse {
	response := []requests.LikeResponse{}

	for _, like := range likes {
		response = append(response, parseLikeResponse(like))
	}

	return response
}

func parseFetchLikeResponse(likes []entities.Like, total_count int) requests.FetchLikesResponse {
	var response requests.FetchLikesResponse

	response.Success = true
	response.TotalCount = total_count
	response.Likes = parseMultipleLikeResponse(likes)

	return response
}

func fetchEntityLikesCount(helper interfaces.LikeHelper, entity_id string, entity_type string) (int64, error) {
	// like filter data
	like_filter_data := gin.H{
		"entity_id":   entity_id,
		"entity_type": entity_type,
		"is_deleted":  false,
	}

	// fetch likes count using helper method
	likes_count, err := helper.CountLikeHelper(like_filter_data)
	if err != nil {
		return 0, err
	}

	return likes_count, nil
}

func fetchEntityLikes(helper interfaces.LikeHelper, entity_id string, entity_type string,
	filterOpts map[string]interface{}) ([]entities.Like, error) {
	// like filter data
	like_filter_data := gin.H{
		"entity_id":   entity_id,
		"entity_type": entity_type,
		"is_deleted":  false,
	}

	// fetch like using helper method
	like_results, err := helper.FindLikeHelper(like_filter_data, filterOpts)
	if err != nil {
		return nil, err
	}

	return like_results, nil
}

func fetchSpecificMemberLikesOnEntity(helper interfaces.LikeHelper, entity_id string, entity_type string,
	member_id string) ([]entities.Like, error) {
	// like filter data
	like_filter_data := gin.H{
		"entity_id":   entity_id,
		"entity_type": entity_type,
		"liked_by":    member_id,
	}

	// fetch like using helper method
	like_results, err := helper.FindLikeHelper(like_filter_data, gin.H{})
	if err != nil {
		return nil, err
	}

	return like_results, nil
}

func (handlers *likeHandlers) LikePost(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")

	// fetch post using helper method
	post_data, err := fetchPost(handlers.postHelper, post_id, headers[utils.HeadersApiKey])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch member like on entity
	like_results, err := fetchSpecificMemberLikesOnEntity(handlers.likeHelper, post_id, constants.PostEntityType,
		headers[utils.HeadersMemberId])
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	if len(like_results) == 0 {
		// create like using the helper method
		_, err := handlers.likeHelper.CreateLikeHelper(constants.PostEntityType, post_data.ID,
			headers[utils.HeadersMemberId])
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	} else {
		like_data := like_results[0]

		// like update data
		like_update_data := gin.H{
			"$set": gin.H{
				"is_deleted": !like_data.IsDeleted,
			},
		}

		// update like using the helper method
		err = handlers.likeHelper.UpdateLikeByIdHelper(like_data.ID, like_update_data)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	}

	// create like activity
	err = createActivity(handlers.activityHelper, constants.LikeAction, post_data.ID, constants.PostEntityType,
		post_data.ApiKey, headers[utils.HeadersMemberId], post_data.UserId, gin.H{
			"entity_type": constants.PostEntityType,
			"post_id":     post_id,
		})
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func (handlers *likeHandlers) FetchPostLikes(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")

	// fetch post data
	_, err := fetchPost(handlers.postHelper, post_id, headers[utils.HeadersApiKey])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch likes count using helper method
	likes_count, err := fetchEntityLikesCount(handlers.likeHelper, post_id, constants.PostEntityType)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// filter options
	like_filter_options, err := generatePageFilterOptions(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch like using helper method
	like_results, err := fetchEntityLikes(handlers.likeHelper, post_id, constants.PostEntityType, like_filter_options)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// return final response
	c.JSON(http.StatusOK, parseFetchLikeResponse(like_results, int(likes_count)))
}

func (handlers *likeHandlers) LikeComment(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")
	comment_id := c.Param("comment_id")

	// fetch post using helper method
	post_data, err := fetchPost(handlers.postHelper, post_id, headers[utils.HeadersApiKey])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment using helper method
	comment_data, err := fetchComment(handlers.commentHelper, comment_id, post_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch member like on entity
	like_results, err := fetchSpecificMemberLikesOnEntity(handlers.likeHelper, comment_id, constants.CommentEntityType,
		headers[utils.HeadersMemberId])
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	if len(like_results) == 0 {
		// create like using the helper method
		_, err := handlers.likeHelper.CreateLikeHelper(constants.CommentEntityType, comment_data.ID,
			headers[utils.HeadersMemberId])
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	} else {
		like_data := like_results[0]

		// like update data
		like_update_data := gin.H{
			"$set": gin.H{
				"is_deleted": !like_data.IsDeleted,
			},
		}

		// update like using the helper method
		err = handlers.likeHelper.UpdateLikeByIdHelper(like_data.ID, like_update_data)
		if err != nil {
			utils.GeneralAPIInternalError(c, err.Error())
			return
		}
	}

	// create like activity
	err = createActivity(handlers.activityHelper, constants.LikeAction, comment_data.ID, constants.CommentEntityType,
		post_data.ApiKey, headers[utils.HeadersMemberId], comment_data.UserId, gin.H{
			"entity_type": constants.CommentEntityType,
			"post_id":     post_id,
			"comment_id":  comment_id,
		})
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func (handlers *likeHandlers) FetchCommentLikes(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	post_id := c.Param("post_id")
	comment_id := c.Param("comment_id")

	// fetch post data
	_, err := fetchPost(handlers.postHelper, post_id, headers[utils.HeadersApiKey])
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch comment using helper method
	_, err = fetchComment(handlers.commentHelper, comment_id, post_id)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch likes count using helper method
	likes_count, err := fetchEntityLikesCount(handlers.likeHelper, comment_id, constants.CommentEntityType)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// filter options
	like_filter_options, err := generatePageFilterOptions(c)
	if err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	// fetch like using helper method
	like_results, err := fetchEntityLikes(handlers.likeHelper, comment_id, constants.CommentEntityType, like_filter_options)
	if err != nil {
		utils.GeneralAPIInternalError(c, err.Error())
		return
	}

	// return final response
	c.JSON(http.StatusOK, parseFetchLikeResponse(like_results, int(likes_count)))
}

type likeHandlers struct {
	likeHelper     interfaces.LikeHelper
	commentHelper  interfaces.CommentHelper
	postHelper     interfaces.PostHelper
	activityHelper interfaces.ActivityHelper
}

func NewLikeHandlers(postHelper interfaces.PostHelper, likeHelper interfaces.LikeHelper, commentHelper interfaces.CommentHelper,
	activityHelper interfaces.ActivityHelper) *likeHandlers {
	return &likeHandlers{
		likeHelper:     likeHelper,
		commentHelper:  commentHelper,
		postHelper:     postHelper,
		activityHelper: activityHelper,
	}
}
