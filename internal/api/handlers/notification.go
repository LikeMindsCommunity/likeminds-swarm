package handlers

import (
	"fmt"

	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SendNotification(activityId primitive.ObjectID, handlers FeedHandlers) {

	activity, err := fetchActivity(handlers.activityHelper, activityId.Hex())
	if err != nil {
		return
	}

	var receivers []string
	var category string
	var subCategory string
	var communityId int
	var title string
	var subTitle string
	var route string

	category = "Feed"
	communityId = activity.CommunityId

	switch activity.Action {
	case constants.LikeAction:
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
		receivers = activity.ActionOn
		title = ""
		route = activity.CTA

		switch activity.EntityType {
		case constants.PostEntityType:
			subCategory = "Post Liked"

			if likesCount == 1 {
				subTitle = fmt.Sprintf("%s liked your post.", member.Name)
			} else if likesCount == 2 {
				subTitle = fmt.Sprintf("%s and 1 other liked your post.", member.Name)
			} else if likesCount > 2 {
				subTitle = fmt.Sprintf("%s and %d others liked your post.", member.Name, likesCount-1)
			}

		case constants.CommentEntityType:
			subCategory = "Comment Liked"

			if likesCount == 1 {
				subTitle = fmt.Sprintf("%s liked your post.", member.Name)
			} else if likesCount == 2 {
				subTitle = fmt.Sprintf("%s and 1 other liked your post.", member.Name)
			} else if likesCount > 2 {
				subTitle = fmt.Sprintf("%s and %d others liked your post.", member.Name, likesCount-1)
			}
		}

	case constants.CommentAction:
		// Fetch member details
		success, member_data := externalHelpers.FetchMemberMeta([]string{activity.ActionBy}, activity.ActionBy, activity.CommunityId)
		if !success || len(member_data.Members) == 0 {
			return
		}

		member := member_data.Members[0]

		// notification params
		receivers = activity.ActionOn
		title = ""
		route = activity.CTA

		switch activity.EntityType {
		case constants.PostEntityType:
			// Fetch comments count
			commentCount, err := fetchPostCommentsCount(handlers.commentHelper, activity.EntityId.Hex())
			if err != nil {
				return
			}

			// If comments count is not in fibonacci series
			if !checkIfFibonacciNumber(int(commentCount)) {
				return
			}

			subCategory = "Post Comment"

			if commentCount == 1 {
				subTitle = fmt.Sprintf("%s commented on your post.", member.Name)
			} else if commentCount == 2 {
				subTitle = fmt.Sprintf("%s and 1 other commented on your post.", member.Name)
			} else if commentCount > 2 {
				subTitle = fmt.Sprintf("%s and %d others commented on your post.", member.Name, commentCount-1)
			}

		case constants.CommentEntityType:
			// Fetch comments count
			commentCount, err := fetchCommentRepliesCount(handlers.commentHelper, activity.EntityId.Hex())
			if err != nil {
				return
			}

			// If comments count is not in fibonacci series
			if !checkIfFibonacciNumber(int(commentCount)) {
				return
			}

			subCategory = "Comment Reply"

			if commentCount == 1 {
				subTitle = fmt.Sprintf("%s replied to your comment.", member.Name)
			} else if commentCount == 2 {
				subTitle = fmt.Sprintf("%s and 1 other replied to your comment.", member.Name)
			} else if commentCount > 2 {
				subTitle = fmt.Sprintf("%s and %d others replied to your comment.", member.Name, commentCount-1)
			}
		}

	case constants.AlsoCommentAction:
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
			receivers = activity.ActionOn
			title = ""
			route = activity.CTA
			subCategory = "Followed Post"

			if len(activity.ActionOn) == 1 {
				subTitle = fmt.Sprintf("%s also commented on the %s's post.", commentOwner, postOwner)
			} else if len(activity.ActionOn) == 2 {
				subTitle = fmt.Sprintf("%s and 1 other also commented on the %s's post.", commentOwner, postOwner)
			} else if len(activity.ActionOn) > 2 {
				subTitle = fmt.Sprintf("%s and %d others also commented on the %s's post.", commentOwner, len(activity.ActionOn)-1, postOwner)
			}
		}

	case constants.TagAction:
		// Fetch member details
		success, member_data := externalHelpers.FetchMemberMeta([]string{activity.ActionBy}, activity.ActionBy, activity.CommunityId)
		if !success || len(member_data.Members) == 0 {
			return
		}

		member := member_data.Members[0]

		// notification params
		receivers = activity.ActionOn
		title = ""
		route = activity.CTA

		switch activity.EntityType {
		case constants.PostEntityType:
			subCategory = "Post Tag"
			subTitle = fmt.Sprintf("%s tagged you in a post.", member.Name)

		case constants.CommentEntityType:
			// Fetch comment data
			comment_data, err := fetchCommentByIdInternal(handlers.commentHelper, activity.EntityId.Hex())
			if err != nil {
				return
			}

			if comment_data.Level == 0 {
				subCategory = "Comment Tag"
				subTitle = fmt.Sprintf("%s tagged you in a comment.", member.Name)
			}

			if comment_data.Level > 0 {
				subCategory = "Reply Tag"
				subTitle = fmt.Sprintf("%s tagged you in a reply.", member.Name)
			}
		}

	case constants.DeleteAction:
		receivers = activity.ActionOn
		route = activity.CTA

		switch activity.EntityType {
		case constants.PostEntityType:
			// Fetch post data
			post_data, err := fetchPost(handlers.postHelper, activity.EntityId.Hex(), activity.CommunityId)
			if err != nil {
				return
			}

			subCategory = "Moderation delete post"
			title = "Post deleted"
			subTitle = fmt.Sprintf("Your post has been deleted as it violates community guidelines. Reason: %s", post_data.DeleteReason)

		case constants.CommentEntityType:
			// Fetch comment data
			comment_data, err := fetchCommentByIdInternal(handlers.commentHelper, activity.EntityId.Hex())
			if err != nil {
				return
			}

			if comment_data.Level == 0 {
				subCategory = "Moderation delete comment"
				title = "Comment deleted"
				subTitle = fmt.Sprintf("Your comment has been deleted as it violates community guidelines. Reason: %s", comment_data.DeleteReason)
			}

			if comment_data.Level > 0 {
				subCategory = "Moderation delete reply"
				title = "Reply deleted"
				subTitle = fmt.Sprintf("Your Reply has been deleted as it violates community guidelines. Reason: %s", comment_data.DeleteReason)
			}
		}

	case constants.CreatePostPermitAddedAction:
		receivers = activity.ActionOn
		subCategory = "Post permission added"
		title = "Permission updated"
		subTitle = "You now have the permission to create posts in the community. Start posting now."
		route = activity.CTA

	case constants.CreatePostPermitRemovedAction:
		receivers = activity.ActionOn
		subCategory = "Post permission removed"
		title = "Permission updated"
		subTitle = "Your permission to create posts in the community has been removed."
		route = "community_home_route"

	case constants.CreateCommentPermissionAddedAction:
		receivers = activity.ActionOn
		subCategory = "Comment permission added"
		title = "Permission updated"
		subTitle = "You now have the permission to add come comments on the posts. Start engaging now."
		route = activity.CTA

	case constants.CreateCommentPermitRemovedAction:
		receivers = activity.ActionOn
		subCategory = "Comment permission removed"
		title = "Permission updated"
		subTitle = "Your permission to add comments and replies to the posts has been removed."
		route = "community_home_route"
	}

	// send notification
	externalHelpers.SendNotification(receivers, title, subTitle, route, communityId, category, subCategory)
}
