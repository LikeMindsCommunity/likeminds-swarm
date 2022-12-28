package handlers

import (
	"fmt"

	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/entities"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func sendCreateCommentPermissionRemovedActionNotification(activity *entities.Activity) {
	receivers := activity.ActionOn
	category := constants.FeedCategory
	subCategory := constants.CommentPermissionRemovedSubCategory
	title := constants.PermissionUpdatedTitle
	subTitle := constants.CommentPermissionRemovedSubTitle
	route := activity.CTA

	// send notification
	externalHelpers.SendNotification(receivers, title, subTitle, route, activity.CommunityId, category, subCategory)
}

func sendCreateCommentPermissionAddedActionNotification(activity *entities.Activity) {
	receivers := activity.ActionOn
	category := constants.FeedCategory
	subCategory := constants.CommentPermissionAddedSubCategory
	title := constants.PermissionUpdatedTitle
	subTitle := constants.CommentPermissionAddedSubTitle
	route := activity.CTA

	// send notification
	externalHelpers.SendNotification(receivers, title, subTitle, route, activity.CommunityId, category, subCategory)
}

func sendCreatePostPermissionRemovedActionNotification(activity *entities.Activity) {
	receivers := activity.ActionOn
	category := constants.FeedCategory
	subCategory := constants.PostPermissionRemovedSubCategory
	title := constants.PermissionUpdatedTitle
	subTitle := constants.PostPermissionRemovedSubTitle
	route := activity.CTA

	// send notification
	externalHelpers.SendNotification(receivers, title, subTitle, route, activity.CommunityId, category, subCategory)
}

func sendCreatePostPermissionAddedActionNotification(activity *entities.Activity) {
	receivers := activity.ActionOn
	category := constants.FeedCategory
	subCategory := constants.PostPermissionAddedSubCategory
	title := constants.PermissionUpdatedTitle
	subTitle := constants.PostPermissionAddedSubTitle
	route := activity.CTA

	// send notification
	externalHelpers.SendNotification(receivers, title, subTitle, route, activity.CommunityId, category, subCategory)
}

func sendPostDeleteActionNotification(activity *entities.Activity, handlers FeedHandlers) {
	// Fetch post data
	post_data, err := fetchPost(handlers.postHelper, activity.EntityId.Hex(), activity.CommunityId)
	if err != nil {
		return
	}

	receivers := activity.ActionOn
	route := activity.CTA
	category := constants.FeedCategory
	subCategory := constants.ModerationPostDeleteSubCategory
	title := constants.PostDeletedTitle
	subTitle := fmt.Sprintf(constants.ModerationPostDeleteSubTitle, post_data.DeleteReason)

	// send notification
	externalHelpers.SendNotification(receivers, title, subTitle, route, activity.CommunityId, category, subCategory)
}

func sendCommentDeleteActionNotification(activity *entities.Activity, handlers FeedHandlers) {
	// Fetch comment data
	comment_data, err := fetchCommentByIdInternal(handlers.commentHelper, activity.EntityId.Hex())
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
	externalHelpers.SendNotification(receivers, title, subTitle, route, activity.CommunityId, category, subCategory)
}

func sendDeleteActionNotification(activity *entities.Activity, handlers FeedHandlers) {
	switch activity.EntityType {
	case constants.PostEntityType:
		sendPostDeleteActionNotification(activity, handlers)

	case constants.CommentEntityType:
		sendCommentDeleteActionNotification(activity, handlers)
	}
}

func sendPostTagActionNotification(activity *entities.Activity, handlers FeedHandlers) {
	// Fetch member details
	success, member_data := externalHelpers.FetchMemberMeta([]string{activity.ActionBy}, activity.ActionBy, activity.CommunityId)
	if !success || len(member_data.Members) == 0 {
		return
	}

	member := member_data.Members[0]

	// notification params
	receivers := activity.ActionOn
	title := ""
	route := activity.CTA
	category := constants.FeedCategory
	subCategory := constants.PostTagSubCategory
	subTitle := fmt.Sprintf(constants.PostTagSubTitle, member.Name)

	// send notification
	externalHelpers.SendNotification(receivers, title, subTitle, route, activity.CommunityId, category, subCategory)
}

func sendCommentTagActionNotification(activity *entities.Activity, handlers FeedHandlers) {
	// Fetch member details
	success, member_data := externalHelpers.FetchMemberMeta([]string{activity.ActionBy}, activity.ActionBy, activity.CommunityId)
	if !success || len(member_data.Members) == 0 {
		return
	}

	member := member_data.Members[0]

	// notification params
	receivers := activity.ActionOn
	title := ""
	route := activity.CTA
	category := constants.FeedCategory
	subCategory := ""
	subTitle := ""

	// Fetch comment data
	comment_data, err := fetchCommentByIdInternal(handlers.commentHelper, activity.EntityId.Hex())
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
	externalHelpers.SendNotification(receivers, title, subTitle, route, activity.CommunityId, category, subCategory)
}

func sendTagActionNotification(activity *entities.Activity, handlers FeedHandlers) {
	switch activity.EntityType {
	case constants.PostEntityType:
		sendPostTagActionNotification(activity, handlers)

	case constants.CommentEntityType:
		sendCommentTagActionNotification(activity, handlers)
	}
}

func sendAlsoCommentActionNotification(activity *entities.Activity, handlers FeedHandlers) {
	switch activity.EntityType {
	case constants.PostEntityType:
		// Fetch post details
		post_data, err := fetchPost(handlers.postHelper, activity.EntityId.Hex(), activity.CommunityId)
		if err != nil {
			return
		}

		// Fetch member details
		success, member_data := externalHelpers.FetchMemberMeta([]string{activity.ActionBy, post_data.UserId}, activity.ActionBy, activity.CommunityId)
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

			if member.UserUniqueId == activity.ActionBy {
				commentOwner = member.Name
			}
		}

		// notification params
		receivers := activity.ActionOn
		title := ""
		route := activity.CTA
		category := constants.FeedCategory
		subCategory := constants.AlsoCommentSubCategory
		subTitle := ""

		if len(activity.ActionOn) == 1 {
			subTitle = fmt.Sprintf(constants.AlsoCommentSubTitleLevelOne, commentOwner, postOwner)
		} else if len(activity.ActionOn) == 2 {
			subTitle = fmt.Sprintf(constants.AlsoCommentSubTitleLevelTwo, commentOwner, postOwner)
		} else if len(activity.ActionOn) > 2 {
			subTitle = fmt.Sprintf(constants.AlsoCommentSubTitleLevelThree, commentOwner, len(activity.ActionOn)-1, postOwner)
		}

		// send notification
		externalHelpers.SendNotification(receivers, title, subTitle, route, activity.CommunityId, category, subCategory)
	}
}

func sendPostCommentActionNotification(activity *entities.Activity, handlers FeedHandlers) {
	// Fetch member details
	success, member_data := externalHelpers.FetchMemberMeta([]string{activity.ActionBy}, activity.ActionBy, activity.CommunityId)
	if !success || len(member_data.Members) == 0 {
		return
	}

	member := member_data.Members[0]

	// notification params
	receivers := activity.ActionOn
	title := ""
	route := activity.CTA
	category := constants.FeedCategory
	subCategory := ""
	subTitle := ""

	// Fetch comments count
	commentCount, err := fetchPostCommentsCount(handlers.commentHelper, activity.EntityId.Hex())
	if err != nil {
		return
	}

	// If comments count is not in fibonacci series
	if !checkIfFibonacciNumber(int(commentCount)) {
		return
	}

	subCategory = constants.PostCommentSubCategory

	if commentCount == 1 {
		subTitle = fmt.Sprintf(constants.PostCommentSubTitleLevelOne, member.Name)
	} else if commentCount == 2 {
		subTitle = fmt.Sprintf(constants.PostCommentSubTitleLevelTwo, member.Name)
	} else if commentCount > 2 {
		subTitle = fmt.Sprintf(constants.PostCommentSubTitleLevelThree, member.Name, commentCount-1)
	}

	// send notification
	externalHelpers.SendNotification(receivers, title, subTitle, route, activity.CommunityId, category, subCategory)
}

func sendCommentReplyActionNotification(activity *entities.Activity, handlers FeedHandlers) {
	// Fetch member details
	success, member_data := externalHelpers.FetchMemberMeta([]string{activity.ActionBy}, activity.ActionBy, activity.CommunityId)
	if !success || len(member_data.Members) == 0 {
		return
	}

	member := member_data.Members[0]

	// notification params
	receivers := activity.ActionOn
	title := ""
	route := activity.CTA
	category := constants.FeedCategory
	subCategory := ""
	subTitle := ""

	// Fetch comments count
	commentCount, err := fetchCommentRepliesCount(handlers.commentHelper, activity.EntityId.Hex())
	if err != nil {
		return
	}

	// If comments count is not in fibonacci series
	if !checkIfFibonacciNumber(int(commentCount)) {
		return
	}

	subCategory = constants.CommentReplySubCategory

	if commentCount == 1 {
		subTitle = fmt.Sprintf(constants.CommentReplySubTitleLevelOne, member.Name)
	} else if commentCount == 2 {
		subTitle = fmt.Sprintf(constants.CommentReplySubTitleLevelTwo, member.Name)
	} else if commentCount > 2 {
		subTitle = fmt.Sprintf(constants.CommentReplySubTitleLevelThree, member.Name, commentCount-1)
	}

	// send notification
	externalHelpers.SendNotification(receivers, title, subTitle, route, activity.CommunityId, category, subCategory)
}

func sendCommentActionNotification(activity *entities.Activity, handlers FeedHandlers) {
	switch activity.EntityType {
	case constants.PostEntityType:
		sendPostCommentActionNotification(activity, handlers)

	case constants.CommentEntityType:
		sendCommentReplyActionNotification(activity, handlers)
	}
}

func sendPostLikeActionNoitification(activity *entities.Activity, handlers FeedHandlers) {
	// Fetch likes count
	likesCount, err := fetchEntityLikesCount(handlers.likeHelper, activity.EntityId.Hex(), activity.EntityType)
	if err != nil {
		return
	}

	// If likes count is not in fibonacci series
	if !checkIfFibonacciNumber(int(likesCount)) {
		return
	}

	// Fetch members details
	success, member_data := externalHelpers.FetchMemberMeta([]string{activity.ActionBy}, activity.ActionBy, activity.CommunityId)
	if !success || len(member_data.Members) == 0 {
		return
	}

	member := member_data.Members[0]

	// notification params
	receivers := activity.ActionOn
	title := ""
	route := activity.CTA
	category := constants.FeedCategory
	subTitle := ""

	subCategory := constants.PostLikedSubCategory

	if likesCount == 1 {
		subTitle = fmt.Sprintf(constants.PostLikedSubTitleLevelOne, member.Name)
	} else if likesCount == 2 {
		subTitle = fmt.Sprintf(constants.PostLikedSubTitleLevelTwo, member.Name)
	} else if likesCount > 2 {
		subTitle = fmt.Sprintf(constants.PostLikedSubTitleLevelThree, member.Name, likesCount-1)
	}

	// send notification
	externalHelpers.SendNotification(receivers, title, subTitle, route, activity.CommunityId, category, subCategory)
}

func sendCommentLikeActionNotification(activity *entities.Activity, handlers FeedHandlers) {
	// Fetch likes count
	likesCount, err := fetchEntityLikesCount(handlers.likeHelper, activity.EntityId.Hex(), activity.EntityType)
	if err != nil {
		return
	}

	// If likes count is not in fibonacci series
	if !checkIfFibonacciNumber(int(likesCount)) {
		return
	}

	// Fetch members details
	success, member_data := externalHelpers.FetchMemberMeta([]string{activity.ActionBy}, activity.ActionBy, activity.CommunityId)
	if !success || len(member_data.Members) == 0 {
		return
	}

	member := member_data.Members[0]

	// notification params
	receivers := activity.ActionOn
	title := ""
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
	externalHelpers.SendNotification(receivers, title, subTitle, route, activity.CommunityId, category, subCategory)
}

func sendLikeActionNotification(activity *entities.Activity, handlers FeedHandlers) {
	switch activity.EntityType {
	case constants.PostEntityType:
		sendPostLikeActionNoitification(activity, handlers)

	case constants.CommentEntityType:
		sendCommentLikeActionNotification(activity, handlers)
	}
}

func SendNotification(activityId primitive.ObjectID, handlers FeedHandlers) {

	activity, err := fetchActivity(handlers.activityHelper, activityId.Hex())
	if err != nil {
		return
	}

	switch activity.Action {
	case constants.SaveAction:
		return

	case constants.LikeAction:
		sendLikeActionNotification(activity, handlers)

	case constants.CommentAction:
		sendCommentActionNotification(activity, handlers)

	case constants.AlsoCommentAction:
		sendAlsoCommentActionNotification(activity, handlers)

	case constants.TagAction:
		sendTagActionNotification(activity, handlers)

	case constants.DeleteAction:
		sendDeleteActionNotification(activity, handlers)

	case constants.CreatePostPermitAddedAction:
		sendCreatePostPermissionAddedActionNotification(activity)

	case constants.CreatePostPermitRemovedAction:
		sendCreatePostPermissionRemovedActionNotification(activity)

	case constants.CreateCommentPermissionAddedAction:
		sendCreateCommentPermissionAddedActionNotification(activity)

	case constants.CreateCommentPermitRemovedAction:
		sendCreateCommentPermissionRemovedActionNotification(activity)
	}
}
