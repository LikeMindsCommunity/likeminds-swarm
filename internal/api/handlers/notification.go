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
	postMetatadataValue := externalHelpers.GetFeedPostVariableOrDefault(handlers.cacheHelper, recieverUUID, communityId)

	receivers := recieverUUID
	category := constants.FeedCategory
	subCategory := constants.PendingPostApprovedSubCategory
	title := fmt.Sprintf(constants.PendingPostApprovedTitle, postMetatadataValue)
	subTitle := fmt.Sprintf(constants.PendingPostApprovedSubTitle, postMetatadataValue)
	route := fmt.Sprintf(utils.PostDetailRoute, postId)

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, communityId, category, subCategory, "", "")
}

// Internal Method to send notification on Pending post rejection
func sendPendingPostRejectionNotification(handlers FeedHandlers, recieverUUID string, communityId int) {

	// Fetch post variable value
	postMetatadataValue := externalHelpers.GetFeedPostVariableOrDefault(handlers.cacheHelper, recieverUUID, communityId)

	receivers := recieverUUID
	category := constants.FeedCategory
	subCategory := constants.PendingPostRejectedSubCategory
	title := fmt.Sprintf(constants.PendingPostRejectedTitle, postMetatadataValue)
	subTitle := fmt.Sprintf(constants.PendingPostRejectedSubTitle, postMetatadataValue)
	route := constants.PlaceholderHomeRoute // placeholder route

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, communityId, category, subCategory, "", "")
}

// Internal Method to send notification on Removal of Create Comment Permission for a user
func sendCreateCommentPermissionRemovedActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {

	// Fetch community configurations
	postMetatadataValue := externalHelpers.GetFeedPostVariableOrDefault(handlers.cacheHelper, activity.ActionOn, activity.CommunityID)

	receivers := activity.ActionOn
	category := constants.FeedCategory
	subCategory := constants.CommentPermissionRemovedSubCategory
	title := constants.PermissionUpdatedTitle
	subTitle := fmt.Sprintf(constants.CommentPermissionRemovedSubTitle, postMetatadataValue)
	route := constants.PlaceholderHomeRoute // placeholder route

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, activity.CommunityID,
		category, subCategory, platform_code, version_code)
}

// Internal Method to send notification on Addition of Create Comment Permission for a user
func sendCreateCommentPermissionAddedActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {

	// Fetch community configurations
	postMetatadataValue := externalHelpers.GetFeedPostVariableOrDefault(handlers.cacheHelper, activity.ActionOn, activity.CommunityID)

	receivers := activity.ActionOn
	category := constants.FeedCategory
	subCategory := constants.CommentPermissionAddedSubCategory
	title := constants.PermissionUpdatedTitle
	subTitle := fmt.Sprintf(constants.CommentPermissionAddedSubTitle, postMetatadataValue)
	route := activity.CTA

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, activity.CommunityID,
		category, subCategory, platform_code, version_code)
}

// Internal Method to send notification on Removal of Create Post Permission for a user
func sendCreatePostPermissionRemovedActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {

	// Fetch community configurations
	postMetatadataValue := externalHelpers.GetFeedPostVariableOrDefault(handlers.cacheHelper, activity.ActionOn, activity.CommunityID)

	receivers := activity.ActionOn
	category := constants.FeedCategory
	subCategory := constants.PostPermissionRemovedSubCategory
	title := constants.PermissionUpdatedTitle
	subTitle := fmt.Sprintf(constants.PostPermissionRemovedSubTitle, postMetatadataValue)
	route := "route://home" // placeholder route

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, activity.CommunityID,
		category, subCategory, platform_code, version_code)
}

// Internal Method to send notification on Addition of Create Post Permission for a user
func sendCreatePostPermissionAddedActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {

	// Fetch community configurations
	postMetatadataValue := externalHelpers.GetFeedPostVariableOrDefault(handlers.cacheHelper, activity.ActionOn, activity.CommunityID)

	receivers := activity.ActionOn
	category := constants.FeedCategory
	subCategory := constants.PostPermissionAddedSubCategory
	title := constants.PermissionUpdatedTitle
	subTitle := fmt.Sprintf(constants.PostPermissionAddedSubTitle, postMetatadataValue)
	route := activity.CTA

	// send notification
	externalHelpers.SendNotification([]string{receivers}, title, subTitle, route, activity.CommunityID,
		category, subCategory, platform_code, version_code)
}

// Internal Method to send notification on Deletion of a Post
func sendPostDeleteActionNotification(activity *entities.Activity, handlers FeedHandlers,
	platform_code string, version_code string) {
	// Fetch post data
	post_data, err := getPostByID(handlers.postHelper, activity.EntityID.Hex())
	if err != nil {
		return
	}

	// Fetch community configurations
	postMetatadataValue := externalHelpers.GetFeedPostVariableOrDefault(handlers.cacheHelper, activity.ActionOn, activity.CommunityID)

	receivers := activity.ActionOn
	route := activity.CTA
	category := constants.FeedCategory
	subCategory := constants.ModerationPostDeleteSubCategory
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

	receivers := activity.ActionOn
	route := activity.CTA
	category := constants.FeedCategory
	subCategory := ""
	title := ""
	subTitle := ""

	if comment_data.Level == 0 {
		subCategory = constants.ModerationCommentDeleteSubCategory
		title = constants.CommentDeletedTitle
		subTitle = fmt.Sprintf(constants.ModerationCommentDeleteSubTitle, comment_data.DeleteReason)
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
	postMetatadataValue := externalHelpers.GetFeedPostVariableOrDefault(handlers.cacheHelper, member.UserUniqueId, activity.CommunityID)

	// notification params
	receivers := activity.ActionOn
	title := constants.TagTitle
	route := activity.CTA
	category := constants.FeedCategory
	subCategory := constants.PostTagSubCategory
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

	if comment_data.Level == 0 {
		subCategory = constants.CommentTagSubCategory
		subTitle = fmt.Sprintf(constants.CommentTagSubTitle, member.Name)
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
		post_data, err := fetchPost(handlers.postHelper, activity.EntityID.Hex(), activity.CommunityID)
		if err != nil {
			return
		}

		latestCommentUserID := activity.ActionBy[len(activity.ActionBy)-1]

		// Fetch member details
		success, member_data := externalHelpers.FetchMemberMeta([](string){latestCommentUserID, activity.EntityOwnerID}, activity.ActionOn, activity.CommunityID)
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
		postMetatadataValue := externalHelpers.GetFeedPostVariableOrDefault(handlers.cacheHelper, latestCommentUserID, activity.CommunityID)

		// notification params
		receivers := activity.ActionOn
		title := constants.CommentTitle
		route := activity.CTA
		category := constants.FeedCategory
		subCategory := constants.AlsoCommentSubCategory
		subTitle := ""

		postCommentUsersCount := len(activity.ActionBy)

		if postCommentUsersCount == 1 {
			subTitle = fmt.Sprintf(constants.AlsoCommentSubTitleLevelOne, commentOwner, postOwner, postMetatadataValue)
		} else if postCommentUsersCount == 2 {
			subTitle = fmt.Sprintf(constants.AlsoCommentSubTitleLevelTwo, commentOwner, postOwner, postMetatadataValue)
		} else if postCommentUsersCount > 2 {
			subTitle = fmt.Sprintf(constants.AlsoCommentSubTitleLevelThree, commentOwner, len(activity.ActionBy)-1, postOwner, postMetatadataValue)
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
	postMetatadataValue := externalHelpers.GetFeedPostVariableOrDefault(handlers.cacheHelper, member.UUID, activity.CommunityID)

	// notification params
	receivers := activity.ActionOn
	title := constants.CommentTitle
	route := activity.CTA
	category := constants.FeedCategory
	subCategory := constants.PostCommentSubCategory
	subTitle := ""

	// Fetch comments count
	commentCount, err := fetchPostCommentsCount(handlers.commentHelper, activity.EntityID.Hex())
	if err != nil {
		return
	}

	// If comments count is not in fibonacci series
	if !checkIfFibonacciNumber(int(commentCount)) {
		return
	}

	postCommentUsersCount := len(activity.ActionBy)

	if postCommentUsersCount == 1 {
		subTitle = fmt.Sprintf(constants.PostCommentSubTitleLevelOne, member.Name, postMetatadataValue)
	} else if postCommentUsersCount == 2 {
		subTitle = fmt.Sprintf(constants.PostCommentSubTitleLevelTwo, member.Name, postMetatadataValue)
	} else if postCommentUsersCount > 2 {
		subTitle = fmt.Sprintf(constants.PostCommentSubTitleLevelThree, member.Name, commentCount-1, postMetatadataValue)
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

	// notification params
	receivers := activity.ActionOn
	title := constants.ReplyTitle
	route := activity.CTA
	category := constants.FeedCategory
	subCategory := constants.CommentReplySubCategory
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
		subTitle = fmt.Sprintf(constants.CommentReplySubTitleLevelOne, member.Name)
	} else if commentReplyUserCount == 2 {
		subTitle = fmt.Sprintf(constants.CommentReplySubTitleLevelTwo, member.Name)
	} else if commentReplyUserCount > 2 {
		subTitle = fmt.Sprintf(constants.CommentReplySubTitleLevelThree, member.Name, commentCount-1)
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
	likesCount, err := fetchEntityLikesCount(handlers.likeHelper, activity.EntityID.Hex(), entityType)
	if err != nil {
		return
	}

	// If likes count is not in fibonacci series
	if !checkIfFibonacciNumber(int(likesCount)) {
		return
	}

	// Fetch members details
	success, member_data := externalHelpers.FetchMemberMeta([]string{activity.ActionBy[len(activity.ActionBy)-1]}, activity.ActionOn, activity.CommunityID)
	if !success || len(member_data.Members) == 0 {
		return
	}

	member := member_data.Members[0]

	// Fetch community configurations
	postMetatadataValue := externalHelpers.GetFeedPostVariableOrDefault(handlers.cacheHelper, member.UUID, activity.CommunityID)

	// notification params
	receivers := activity.ActionOn
	title := constants.LikeTitle
	route := activity.CTA
	category := constants.FeedCategory
	subTitle := ""
	subCategory := constants.PostLikedSubCategory

	if likesCount == 1 {
		subTitle = fmt.Sprintf(constants.PostLikedSubTitleLevelOne, member.Name, postMetatadataValue)
	} else if likesCount == 2 {
		subTitle = fmt.Sprintf(constants.PostLikedSubTitleLevelTwo, member.Name, postMetatadataValue)
	} else if likesCount > 2 {
		subTitle = fmt.Sprintf(constants.PostLikedSubTitleLevelThree, member.Name, likesCount-1, postMetatadataValue)
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
	likesCount, err := fetchEntityLikesCount(handlers.likeHelper, activity.EntityID.Hex(), entityType)
	if err != nil {
		return
	}

	// If likes count is not in fibonacci series
	if !checkIfFibonacciNumber(int(likesCount)) {
		return
	}

	// Fetch members details
	success, member_data := externalHelpers.FetchMemberMeta([]string{activity.ActionBy[len(activity.ActionBy)-1]}, activity.ActionOn, activity.CommunityID)
	if !success || len(member_data.Members) == 0 {
		return
	}

	member := member_data.Members[0]

	// notification params
	receivers := activity.ActionOn
	title := constants.LikeTitle
	route := activity.CTA
	category := constants.FeedCategory
	subTitle := ""
	subCategory := constants.CommentLikedSubCategory

	if likesCount == 1 {
		subTitle = fmt.Sprintf(constants.CommentLikedSubTitleLevelOne, member.Name)
	} else if likesCount == 2 {
		subTitle = fmt.Sprintf(constants.CommentLikedSubTitleLevelTwo, member.Name)
	} else if likesCount > 2 {
		subTitle = fmt.Sprintf(constants.CommentLikedSubTitleLevelThree, member.Name, likesCount-1)
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

	postData, err := fetchPost(handlers.postHelper, activity.EntityID.Hex(), activity.CommunityID)
	if err != nil {
		return
	}
	repostCount := getPostRepostCount(handlers.widgetHelper, *postData)

	// If repost count is not in fibonacci series
	if !checkIfFibonacciNumber(int(repostCount)) {
		return
	}

	// Fetch members details
	success, member_data := externalHelpers.FetchMemberMeta([]string{activity.ActionBy[len(activity.ActionBy)-1]}, activity.ActionOn, activity.CommunityID)
	if !success || len(member_data.Members) == 0 {
		return
	}

	member := member_data.Members[0]

	// Fetch community configurations
	postMetatadataValue := externalHelpers.GetFeedPostVariableOrDefault(handlers.cacheHelper, member.UUID, activity.CommunityID)

	// notification params
	receivers := activity.ActionOn
	title := constants.RepostTitle
	route := activity.CTA
	category := constants.FeedCategory
	subTitle := ""
	subCategory := constants.RepostOnPostSubCategory

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
