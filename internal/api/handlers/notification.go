package handlers

import (
	"fmt"

	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to send notification on Pending post approval
func sendPendingPostApprovalNotification(handlers FeedHandlers, recieverUUID string, communityId int, postId string) {

	// Fetch post variable value
	postMetatadataValue := externalHelpers.GetPostVariableOrDefault(handlers.cacheHelper, recieverUUID, communityId)

	receivers := recieverUUID
	category := constants.FeedCategory
	subCategory := fmt.Sprintf(constants.PendingPostApprovedSubCategory, utils.CapitalizeFirstLetter(postMetatadataValue))
	title := fmt.Sprintf(constants.PendingPostApprovedTitle, utils.CapitalizeFirstLetter(postMetatadataValue))
	subTitle := fmt.Sprintf(constants.PendingPostApprovedSubTitle, postMetatadataValue)
	route := fmt.Sprintf(utils.PostDetailRoute, postId)

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, communityId, category, subCategory, "", "")
}

// Internal Method to send notification on Pending post rejection
func sendPendingPostRejectionNotification(handlers FeedHandlers, recieverUUID string, communityId int, pendingPostId string) {

	// Fetch post variable value
	postMetatadataValue := externalHelpers.GetPostVariableOrDefault(handlers.cacheHelper, recieverUUID, communityId)

	receivers := recieverUUID
	category := constants.FeedCategory
	subCategory := fmt.Sprintf(constants.PendingPostRejectedSubCategory, postMetatadataValue)
	title := fmt.Sprintf(constants.PendingPostRejectedTitle, utils.CapitalizeFirstLetter(postMetatadataValue))
	subTitle := fmt.Sprintf(constants.PendingPostRejectedSubTitle, postMetatadataValue)
	route := fmt.Sprintf(utils.PendingPostDetailRoute, pendingPostId)

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, communityId, category, subCategory, "", "")
}

// Internal Method to send notification on Removal of Create Comment Permission for a user
func sendCreateCommentPermissionRemovedActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {

	// Fetch community configurations
	postMetatadataValue := externalHelpers.GetPostVariableOrDefault(handlers.cacheHelper, activity.ActionOn, activity.CommunityID)
	commentMetatadataValue := externalHelpers.GetCommentVariableOrDefault(handlers.cacheHelper, activity.ActionOn, activity.CommunityID)

	receivers := activity.ActionOn
	category := constants.FeedCategory
	subCategory := fmt.Sprintf(constants.CommentPermissionRemovedSubCategory, utils.CapitalizeFirstLetter(commentMetatadataValue))
	title := constants.PermissionUpdatedTitle
	subTitle := fmt.Sprintf(constants.CommentPermissionRemovedSubTitle, utils.GetPluralOfString(commentMetatadataValue), utils.GetPluralOfString(postMetatadataValue))
	route := constants.PlaceholderHomeRoute // placeholder route

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, activity.CommunityID,
		category, subCategory, platform_code, version_code)
}

// Internal Method to send notification on Addition of Create Comment Permission for a user
func sendCreateCommentPermissionAddedActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {

	// Fetch community configurations
	postMetatadataValue := externalHelpers.GetPostVariableOrDefault(handlers.cacheHelper, activity.ActionOn, activity.CommunityID)
	commentMetatadataValue := externalHelpers.GetCommentVariableOrDefault(handlers.cacheHelper, activity.ActionOn, activity.CommunityID)

	receivers := activity.ActionOn
	category := constants.FeedCategory
	subCategory := fmt.Sprintf(constants.CommentPermissionAddedSubCategory, utils.CapitalizeFirstLetter(commentMetatadataValue))
	title := constants.PermissionUpdatedTitle
	subTitle := fmt.Sprintf(constants.CommentPermissionAddedSubTitle, utils.GetPluralOfString(commentMetatadataValue), utils.GetPluralOfString(postMetatadataValue))
	route := activity.CTA

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, activity.CommunityID,
		category, subCategory, platform_code, version_code)
}

// Internal Method to send notification on Removal of Create Post Permission for a user
func sendCreatePostPermissionRemovedActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {

	// Fetch community configurations
	postMetatadataValue := externalHelpers.GetPostVariableOrDefault(handlers.cacheHelper, activity.ActionOn, activity.CommunityID)

	receivers := activity.ActionOn
	category := constants.FeedCategory
	subCategory := fmt.Sprintf(constants.PostPermissionRemovedSubCategory, utils.CapitalizeFirstLetter(postMetatadataValue))
	title := constants.PermissionUpdatedTitle
	subTitle := fmt.Sprintf(constants.PostPermissionRemovedSubTitle, utils.GetPluralOfString(postMetatadataValue))
	route := "route://home" // placeholder route

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, activity.CommunityID,
		category, subCategory, platform_code, version_code)
}

// Internal Method to send notification on Addition of Create Post Permission for a user
func sendCreatePostPermissionAddedActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {

	// Fetch community configurations
	postMetatadataValue := externalHelpers.GetPostVariableOrDefault(handlers.cacheHelper, activity.ActionOn, activity.CommunityID)

	receivers := activity.ActionOn
	category := constants.FeedCategory
	subCategory := fmt.Sprintf(constants.PostPermissionAddedSubCategory, utils.CapitalizeFirstLetter(postMetatadataValue))
	title := constants.PermissionUpdatedTitle
	subTitle := fmt.Sprintf(constants.PostPermissionAddedSubTitle, postMetatadataValue)
	route := activity.CTA

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, activity.CommunityID,
		category, subCategory, platform_code, version_code)
}

// Internal Method to send notification on Deletion of a Post
func sendPostDeleteActionNotification(activity *entities.Activity, handlers FeedHandlers, platform_code string, version_code string,
) {
	// Fetch post data
	post_data, err := FetchPostData(handlers.postHelper, activity.EntityID.Hex(), activity.CommunityID, false)
	if err != nil {
		return
	}

	// Fetch community configurations
	postMetatadataValue := externalHelpers.GetPostVariableOrDefault(handlers.cacheHelper, activity.ActionOn, activity.CommunityID)

	receivers := activity.ActionOn
	route := activity.CTA
	category := constants.FeedCategory
	subCategory := fmt.Sprintf(constants.ModerationPostDeleteSubCategory, postMetatadataValue)
	title := fmt.Sprintf(constants.PostDeletedTitle, utils.CapitalizeFirstLetter(postMetatadataValue))
	subTitle := fmt.Sprintf(constants.ModerationPostDeleteSubTitle, postMetatadataValue, post_data.DeleteReason)

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, activity.CommunityID,
		category, subCategory, platform_code, version_code)
}

// Internal Method to send notification on Deletion of a Comment
func sendCommentDeleteActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {
	// Fetch comment data
	comment_data, err := fetchCommentByID(handlers.commentHelper, activity.EntityID.Hex())
	if err != nil {
		return
	}

	// Fetch community configurations
	commentMetatadataValue := externalHelpers.GetCommentVariableOrDefault(handlers.cacheHelper, activity.ActionOn, activity.CommunityID)

	receivers := activity.ActionOn
	route := activity.CTA
	category := constants.FeedCategory
	subCategory := ""
	title := ""
	subTitle := ""

	if comment_data.Level == 0 {
		subCategory = fmt.Sprintf(constants.ModerationCommentDeleteSubCategory, commentMetatadataValue)
		title = fmt.Sprintf(constants.CommentDeletedTitle, utils.CapitalizeFirstLetter(commentMetatadataValue))
		subTitle = fmt.Sprintf(constants.ModerationCommentDeleteSubTitle, commentMetatadataValue, comment_data.DeleteReason)
	}

	if comment_data.Level > 0 {
		subCategory = constants.ModerationReplyDeleteSubCategory
		title = constants.ReplyDeletedTitle
		subTitle = fmt.Sprintf(constants.ModerationReplyDeleteSubTitle, comment_data.DeleteReason)
	}

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, activity.CommunityID,
		category, subCategory, platform_code, version_code)
}

// Internal General Method to send notification on deletion of an Entity
func sendDeleteActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {
	switch activity.EntityType {
	case constants.Post:
		sendPostDeleteActionNotification(activity, handlers, platform_code, version_code)

	case constants.Comment:
		sendCommentDeleteActionNotification(activity, handlers, platform_code, version_code)
	}
}

// Internal Method to send notification on tagging of a user on a Post
func sendPostTagActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platformCode string, versionCode string) {
	// Fetch member details
	success, memberData := externalHelpers.FetchMemberMeta([]string{activity.ActionBy[len(activity.ActionBy)-1]}, activity.ActionOn, activity.CommunityID)
	if !success || len(memberData.Members) == 0 {
		return
	}

	member := memberData.Members[0]

	// Fetch community configurations
	postMetatadataValue := externalHelpers.GetPostVariableOrDefault(handlers.cacheHelper, member.UserUniqueId, activity.CommunityID)

	// notification params
	receivers := activity.ActionOn
	title := constants.TagTitle
	route := activity.CTA
	category := constants.FeedCategory
	subCategory := fmt.Sprintf(constants.PostTagSubCategory, utils.CapitalizeFirstLetter(postMetatadataValue))
	subTitle := fmt.Sprintf(constants.PostTagSubTitle, member.Name, postMetatadataValue)

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, activity.CommunityID,
		category, subCategory, platformCode, versionCode)
}

// Internal Method to send notification on tagging of a user on a Comment
func sendCommentTagActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {
	// Fetch member details
	success, member_data := externalHelpers.FetchMemberMeta([]string{activity.ActionBy[len(activity.ActionBy)-1]}, activity.ActionOn, activity.CommunityID)
	if !success || len(member_data.Members) == 0 {
		return
	}

	member := member_data.Members[0]

	// notification params
	receivers := activity.ActionOn
	title := constants.TagTitle
	route := activity.CTA
	category := constants.FeedCategory
	subCategory := ""
	subTitle := ""

	// Fetch comment data
	comment_data, err := fetchCommentByIdInternal(handlers.commentHelper, activity.EntityID.Hex())
	if err != nil {
		return
	}

	// Fetch community configurations
	commentMetatadataValue := externalHelpers.GetCommentVariableOrDefault(handlers.cacheHelper, activity.ActionOn, activity.CommunityID)

	if comment_data.Level == 0 {
		subCategory = fmt.Sprintf(constants.CommentTagSubCategory, utils.CapitalizeFirstLetter(commentMetatadataValue))
		subTitle = fmt.Sprintf(constants.CommentTagSubTitle, member.Name, commentMetatadataValue)
	}

	if comment_data.Level > 0 {
		subCategory = constants.ReplyTagSubCategory
		subTitle = fmt.Sprintf(constants.ReplyTagSubTitle, member.Name)
	}

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, activity.CommunityID,
		category, subCategory, platform_code, version_code)
}

// Internal General Method to send notification on tagging of a user on an Entity
func sendTagActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {
	switch activity.EntityType {
	case constants.Post:
		sendPostTagActionNotification(activity, handlers, platform_code, version_code)

	case constants.Comment:
		sendCommentTagActionNotification(activity, handlers, platform_code, version_code)
	}
}

// Internal Method to send notification of also comment action on a Post
func sendAlsoCommentActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {
	switch activity.EntityType {
	case constants.Post:
		// Fetch post details
		post_data, err := FetchPostData(handlers.postHelper, activity.EntityID.Hex(), activity.CommunityID, true)
		if err != nil {
			return
		}

		latestCommentUserID := activity.ActionBy[len(activity.ActionBy)-1]

		// Fetch member details
		success, member_data := externalHelpers.FetchMemberMeta([]string{latestCommentUserID, activity.EntityOwnerID}, activity.ActionOn, activity.CommunityID)
		if !success || len(member_data.Members) == 0 {
			return
		}

		members := member_data.Members
		var postOwner string
		var commentOwner string

		// member names parsing
		for _, member := range members {
			if member.UserUniqueId == post_data.UserId {
				postOwner = member.Name
			}

			if member.UserUniqueId == latestCommentUserID {
				commentOwner = member.Name
			}
		}

		// Fetch community configurations
		postMetatadataValue := externalHelpers.GetPostVariableOrDefault(handlers.cacheHelper, latestCommentUserID, activity.CommunityID)
		commentMetatadataValue := externalHelpers.GetCommentVariableOrDefault(handlers.cacheHelper, latestCommentUserID, activity.CommunityID)

		// notification params
		receivers := activity.ActionOn
		title := fmt.Sprintf(constants.CommentTitle, utils.CapitalizeFirstLetter(commentMetatadataValue))
		route := activity.CTA
		category := constants.FeedCategory
		subCategory := fmt.Sprintf(constants.AlsoCommentSubCategory, utils.CapitalizeFirstLetter(postMetatadataValue))
		subTitle := ""

		postCommentUsersCount := len(activity.ActionBy)

		if postCommentUsersCount == 1 {
			subTitle = fmt.Sprintf(constants.AlsoCommentSubTitleLevelOne, commentOwner, commentMetatadataValue, postOwner, postMetatadataValue)
		} else if postCommentUsersCount == 2 {
			subTitle = fmt.Sprintf(constants.AlsoCommentSubTitleLevelTwo, commentOwner, utils.GetPluralOfString(commentMetatadataValue),
				postOwner, postMetatadataValue)
		} else if postCommentUsersCount > 2 {
			subTitle = fmt.Sprintf(constants.AlsoCommentSubTitleLevelThree, commentOwner, len(activity.ActionBy)-1, utils.GetPluralOfString(commentMetatadataValue),
				postOwner, postMetatadataValue)
		}

		// send notification
		externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, activity.CommunityID,
			category, subCategory, platform_code, version_code)
	}
}

// Internal Method to send notification on comment action on a Post
func sendPostCommentActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {
	// Fetch member details
	success, member_data := externalHelpers.FetchMemberMeta([]string{activity.ActionBy[len(activity.ActionBy)-1]}, activity.ActionOn, activity.CommunityID)
	if !success || len(member_data.Members) == 0 {
		return
	}

	member := member_data.Members[0]

	// Fetch community configurations
	postMetatadataValue := externalHelpers.GetPostVariableOrDefault(handlers.cacheHelper, member.UUID, activity.CommunityID)
	commentMetatadataValue := externalHelpers.GetCommentVariableOrDefault(handlers.cacheHelper, member.UUID, activity.CommunityID)

	// notification params
	receivers := activity.ActionOn
	title := fmt.Sprintf(constants.CommentTitle, utils.CapitalizeFirstLetter(commentMetatadataValue))
	route := activity.CTA
	category := constants.FeedCategory
	subCategory := fmt.Sprintf(constants.PostCommentSubCategory, utils.CapitalizeFirstLetter(postMetatadataValue), utils.CapitalizeFirstLetter(commentMetatadataValue))
	subTitle := ""

	// Fetch comments count
	commentCount := fetchPostCommentsCount(handlers.commentHelper, activity.EntityID.Hex())

	// If comments count is not in fibonacci series
	if !checkIfFibonacciNumber(commentCount) {
		return
	}

	postCommentUsersCount := len(activity.ActionBy)

	if postCommentUsersCount == 1 {
		subTitle = fmt.Sprintf(constants.PostCommentSubTitleLevelOne, member.Name, commentMetatadataValue, postMetatadataValue)
	} else if postCommentUsersCount == 2 {
		subTitle = fmt.Sprintf(constants.PostCommentSubTitleLevelTwo, member.Name, utils.GetPluralOfString(commentMetatadataValue),
			postMetatadataValue)
	} else if postCommentUsersCount > 2 {
		subTitle = fmt.Sprintf(constants.PostCommentSubTitleLevelThree, member.Name, commentCount-1, utils.GetPluralOfString(commentMetatadataValue),
			postMetatadataValue)
	}

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, activity.CommunityID,
		category, subCategory, platform_code, version_code)
}

// Internal Method to send notification on reply action on a Comment
func sendCommentReplyActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {
	// Fetch member details
	success, member_data := externalHelpers.FetchMemberMeta([]string{activity.ActionBy[len(activity.ActionBy)-1]}, activity.ActionOn, activity.CommunityID)
	if !success || len(member_data.Members) == 0 {
		return
	}

	member := member_data.Members[0]

	// Fetch community configurations
	commentMetatadataValue := externalHelpers.GetCommentVariableOrDefault(handlers.cacheHelper, member.UUID, activity.CommunityID)

	// notification params
	receivers := activity.ActionOn
	title := constants.ReplyTitle
	route := activity.CTA
	category := constants.FeedCategory
	subCategory := fmt.Sprintf(constants.CommentReplySubCategory, utils.CapitalizeFirstLetter(commentMetatadataValue))
	subTitle := ""

	// Fetch comments count
	commentCount, err := fetchCommentRepliesCount(handlers.commentHelper, activity.EntityID.Hex())
	if err != nil {
		return
	}

	// If comments count is not in fibonacci series
	if !checkIfFibonacciNumber(int(commentCount)) {
		return
	}

	commentReplyUserCount := len(activity.ActionBy)

	if commentReplyUserCount == 1 {
		subTitle = fmt.Sprintf(constants.CommentReplySubTitleLevelOne, member.Name, commentMetatadataValue)
	} else if commentReplyUserCount == 2 {
		subTitle = fmt.Sprintf(constants.CommentReplySubTitleLevelTwo, member.Name, commentMetatadataValue)
	} else if commentReplyUserCount > 2 {
		subTitle = fmt.Sprintf(constants.CommentReplySubTitleLevelThree, member.Name, commentCount-1, commentMetatadataValue)
	}

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, activity.CommunityID,
		category, subCategory, platform_code, version_code)
}

// Internal General Method to send notification on comment action on an Entity
func sendCommentActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {
	switch activity.EntityType {
	case constants.Post:
		sendPostCommentActionNotification(activity, handlers, platform_code, version_code)

	case constants.Comment:
		sendCommentReplyActionNotification(activity, handlers, platform_code, version_code)
	}
}

// Internal Method to send notification on like action on a Post
func sendPostLikeActionNoitification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {

	entityType := ""
	switch activity.EntityType {
	case constants.Post:
		entityType = "post"
	case constants.Comment:
		entityType = "comment"
	case constants.User:
		entityType = "user"
	}

	// Fetch likes count
	likesCount := fetchEntityLikesCount(handlers.likeHelper, activity.EntityID.Hex(), entityType)

	// If likes count is not in fibonacci series
	if !checkIfFibonacciNumber(likesCount) {
		return
	}

	// Fetch members details
	success, member_data := externalHelpers.FetchMemberMeta([]string{activity.ActionBy[len(activity.ActionBy)-1]}, activity.ActionOn, activity.CommunityID)
	if !success || len(member_data.Members) == 0 {
		return
	}

	member := member_data.Members[0]

	// Fetch community configurations
	postMetatadataValue := externalHelpers.GetPostVariableOrDefault(handlers.cacheHelper, member.UUID, activity.CommunityID)
	likePresentValue, likePastValue := externalHelpers.GetLikeVariablesOrDefault(handlers.cacheHelper, member.UUID, activity.CommunityID)

	// notification params
	receivers := activity.ActionOn
	title := fmt.Sprintf(constants.LikeTitle, likePresentValue)
	route := activity.CTA
	category := constants.FeedCategory
	subTitle := ""
	subCategory := fmt.Sprintf(constants.PostLikedSubCategory, utils.CapitalizeFirstLetter(postMetatadataValue), likePastValue)

	if likesCount == 1 {
		subTitle = fmt.Sprintf(constants.PostLikedSubTitleLevelOne, member.Name, likePastValue, postMetatadataValue)
	} else if likesCount == 2 {
		subTitle = fmt.Sprintf(constants.PostLikedSubTitleLevelTwo, member.Name, likePastValue, postMetatadataValue)
	} else if likesCount > 2 {
		subTitle = fmt.Sprintf(constants.PostLikedSubTitleLevelThree, member.Name, likesCount-1, likePastValue, postMetatadataValue)
	}

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, activity.CommunityID,
		category, subCategory, platform_code, version_code)
}

// Internal Method to send notification on like action on a Comment
func sendCommentLikeActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {

	entityType := ""
	switch activity.EntityType {
	case constants.Post:
		entityType = "post"
	case constants.Comment:
		entityType = "comment"
	case constants.User:
		entityType = "user"
	}
	// Fetch likes count
	likesCount := fetchEntityLikesCount(handlers.likeHelper, activity.EntityID.Hex(), entityType)

	// If likes count is not in fibonacci series
	if !checkIfFibonacciNumber(likesCount) {
		return
	}

	// Fetch members details
	success, member_data := externalHelpers.FetchMemberMeta([]string{activity.ActionBy[len(activity.ActionBy)-1]}, activity.ActionOn, activity.CommunityID)
	if !success || len(member_data.Members) == 0 {
		return
	}

	member := member_data.Members[0]

	// Fetch community configurations
	commentMetatadataValue := externalHelpers.GetCommentVariableOrDefault(handlers.cacheHelper, member.UUID, activity.CommunityID)
	likePresentValue, likePastValue := externalHelpers.GetLikeVariablesOrDefault(handlers.cacheHelper, member.UUID, activity.CommunityID)

	// notification params
	receivers := activity.ActionOn
	title := fmt.Sprintf(constants.LikeTitle, likePresentValue)
	route := activity.CTA
	category := constants.FeedCategory
	subTitle := ""
	subCategory := fmt.Sprintf(constants.CommentLikedSubCategory, utils.CapitalizeFirstLetter(commentMetatadataValue), likePastValue)

	if likesCount == 1 {
		subTitle = fmt.Sprintf(constants.CommentLikedSubTitleLevelOne, member.Name, likePastValue, commentMetatadataValue)
	} else if likesCount == 2 {
		subTitle = fmt.Sprintf(constants.CommentLikedSubTitleLevelTwo, member.Name, likePastValue, commentMetatadataValue)
	} else if likesCount > 2 {
		subTitle = fmt.Sprintf(constants.CommentLikedSubTitleLevelThree, member.Name, likesCount-1, likePastValue, commentMetatadataValue)
	}

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, activity.CommunityID,
		category, subCategory, platform_code, version_code)
}

// Internal General Method to send notification on like action on an Entity
func sendLikeActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platformCode string, versionCode string) {
	switch activity.EntityType {
	case constants.Post:
		sendPostLikeActionNoitification(activity, handlers, platformCode, versionCode)

	case constants.Comment:
		sendCommentLikeActionNotification(activity, handlers, platformCode, versionCode)
	}
}

// Internal General Method to send notification on repost action on a post
func sendRepostPostActionNotification(activity *entities.Activity, handlers FeedHandlers, platformCode string, versionCode string) {

	postData, err := FetchPostData(handlers.postHelper, activity.EntityID.Hex(), activity.CommunityID, true)
	if err != nil {
		return
	}
	repostCount := getPostRepostCount(handlers.widgetHelper, postData)

	// If repost count is not in fibonacci series
	if !checkIfFibonacciNumber(repostCount) {
		return
	}

	// Fetch members details
	success, member_data := externalHelpers.FetchMemberMeta([]string{activity.ActionBy[len(activity.ActionBy)-1]}, activity.ActionOn, activity.CommunityID)
	if !success || len(member_data.Members) == 0 {
		return
	}

	member := member_data.Members[0]

	// Fetch community configurations
	postMetatadataValue := externalHelpers.GetPostVariableOrDefault(handlers.cacheHelper, member.UUID, activity.CommunityID)

	// notification params
	receivers := activity.ActionOn
	title := constants.RepostTitle
	route := activity.CTA
	category := constants.FeedCategory
	subTitle := ""
	subCategory := fmt.Sprintf(constants.RepostOnPostSubCategory, utils.CapitalizeFirstLetter(postMetatadataValue))

	if repostCount == 1 {
		subTitle = fmt.Sprintf(constants.PostRepostedSubTitleLevelOne, member.Name, postMetatadataValue)
	} else if repostCount == 2 {
		subTitle = fmt.Sprintf(constants.PostRepostedSubTitleLevelTwo, member.Name, postMetatadataValue)
	} else if repostCount > 2 {
		subTitle = fmt.Sprintf(constants.PostRepostedSubTitleLevelThree, member.Name, repostCount-1, postMetatadataValue)
	}

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, activity.CommunityID,
		category, subCategory, platformCode, versionCode)
}

// Internal Method to validate notification receivers
func validateReceivers(activity *entities.Activity) *entities.Activity {
	receivers := []string{}

	for _, receiver := range activity.ActionBy {
		if receiver != activity.ActionOn {
			receivers = append(receivers, receiver)
		}
	}

	newActivity := &entities.Activity{
		ID:            activity.ID,
		CommunityID:   activity.CommunityID,
		ActionBy:      receivers,
		ActionOn:      activity.ActionOn,
		EntityType:    activity.EntityType,
		EntityID:      activity.EntityID,
		EntityOwnerID: activity.EntityOwnerID,
		Action:        activity.Action,
		CTA:           activity.CTA,
		IsRead:        activity.IsRead,
		IsDeleted:     activity.IsDeleted,
		CreatedAt:     activity.CreatedAt,
		UpdatedAt:     activity.UpdatedAt,
	}

	return newActivity
}

// SendNotification | method to send notification for activity
func SendNotification(handlers FeedHandlers, activityID primitive.ObjectID, platformCode string, versionCode string) error {

	activity, err := fetchActivity(handlers.activityHelper, activityID.Hex())
	if err != nil {
		return fmt.Errorf("failed to fetch activity: %w", err)
	}

	// Don't send notification when action done by the entity creator
	activity = validateReceivers(activity)
	if len(activity.ActionBy) == 0 {
		return fmt.Errorf("no valid receivers found")
	}

	switch activity.Action {

	case constants.LikeOnPost, constants.LikeOnComment:
		sendLikeActionNotification(activity, handlers, platformCode, versionCode)

	case constants.CommentOnPost, constants.CommentOnComment:
		sendCommentActionNotification(activity, handlers, platformCode, versionCode)

	case constants.AlsoCommentOnPost:
		sendAlsoCommentActionNotification(activity, handlers, platformCode, versionCode)

	case constants.RepostOnPost:
		sendRepostPostActionNotification(activity, handlers, platformCode, versionCode)

	case constants.TaggedInPost, constants.TaggedInPostComment:
		sendTagActionNotification(activity, handlers, platformCode, versionCode)

	case constants.CMDeletedPost, constants.CMDeletedComment:
		sendDeleteActionNotification(activity, handlers, platformCode, versionCode)

	case constants.CreatePostPermitAdded:
		sendCreatePostPermissionAddedActionNotification(activity, handlers, platformCode, versionCode)

	case constants.CreatePostPermitRemoved:
		sendCreatePostPermissionRemovedActionNotification(activity, handlers, platformCode, versionCode)

	case constants.CreateCommentPermitAdded:
		sendCreateCommentPermissionAddedActionNotification(activity, handlers, platformCode, versionCode)

	case constants.CreateCommentPermitRemoved:
		sendCreateCommentPermissionRemovedActionNotification(activity, handlers, platformCode, versionCode)
	}

	return nil
}
