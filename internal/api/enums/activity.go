package enums

import "github.com/LikeMindsCommunity/likeminds-swarm/internal/api/constants"

// EntityType | defines enum for activity entities
type EntityType string

const (
	Post        EntityType = "post"
	Comment     EntityType = "comment"
	User        EntityType = "user"
	PendingPost EntityType = "pending_post"
)

// constructor method to creates EntityType from int
func NewEntityTypeFromInt(entity_type int) EntityType {
	switch entity_type {
	case int(constants.PostEntity):
		return Post
	case int(constants.CommentEntity):
		return Comment
	case int(constants.UserEntity):
		return User
	case int(constants.PendingPostEntity):
		return PendingPost
	}

	return ""
}

// constructor method to creates EntityType from string
func NewIntEntityTypeFromString(entityType string) constants.EntityType {
	switch entityType {
	case Post.ToString():
		return constants.PostEntity
	case Comment.ToString():
		return constants.CommentEntity
	case User.ToString():
		return constants.UserEntity
	case PendingPost.ToString():
		return constants.PendingPostEntity
	}

	return constants.DefaultEntity
}

// To string method for EntityType
func (et EntityType) ToString() string {
	return string(et)
}

// ActivityAction | defines enum for activity actions
type ActivityAction string

const (
	CustomActivity             ActivityAction = "custom_activity"
	CreatePostPermitAdded      ActivityAction = "create_post_permit_added"
	CreatePostPermitRemoved    ActivityAction = "create_post_permit_removed"
	CreateCommentPermitAdded   ActivityAction = "create_comment_permit_added"
	CreateCommentPermitRemoved ActivityAction = "create_comment_permit_removed"
	CMDeletedPost              ActivityAction = "cm_deleted_your_post"
	CMDeletedComment           ActivityAction = "cm_deleted_your_comment"
	LikeOnPost                 ActivityAction = "like_on_your_post"
	CommentOnPost              ActivityAction = "comment_on_your_post"
	LikeOnComment              ActivityAction = "like_on_your_comment"
	CommentOnComment           ActivityAction = "comment_on_your_comment"
	TaggedYouOnPost            ActivityAction = "tagged_you_on_post"
	TaggedYouOnComment         ActivityAction = "tagged_you_on_comment_post"
	AlsoCommentedOnPost        ActivityAction = "also_commented_on_post_you_commented"
	RepostOnPost               ActivityAction = "reposted_your_post"

	// User Profile Activity Actions
	UserLikeOnPost       ActivityAction = "like_on_post"
	UserLikeOnComment    ActivityAction = "like_on_comment"
	UserCommentOnPost    ActivityAction = "comment_on_post"
	UserCommentOnComment ActivityAction = "comment_on_comment"
)

// constructor method to creates ActivityAction from int
func NewActivityActionFromInt(activity_action int, userProfileActivity bool) ActivityAction {
	switch activity_action {
	case int(constants.CustomActivity):
		return CustomActivity
	case int(constants.CreatePostPermitAdded):
		return CreatePostPermitAdded
	case int(constants.CreatePostPermitRemoved):
		return CreatePostPermitRemoved
	case int(constants.CreateCommentPermitAdded):
		return CreateCommentPermitAdded
	case int(constants.CreateCommentPermitRemoved):
		return CreateCommentPermitRemoved
	case int(constants.CMDeletedPost):
		return CMDeletedPost
	case int(constants.CMDeletedComment):
		return CMDeletedComment
	case int(constants.LikeOnPost):
		if userProfileActivity {
			return UserLikeOnPost
		} else {
			return LikeOnPost
		}
	case int(constants.CommentOnPost):
		if userProfileActivity {
			return UserCommentOnPost
		} else {
			return CommentOnPost
		}
	case int(constants.LikeOnComment):
		if userProfileActivity {
			return UserLikeOnComment
		} else {
			return LikeOnComment
		}
	case int(constants.CommentOnComment):
		if userProfileActivity {
			return UserCommentOnComment
		} else {
			return CommentOnComment
		}
	case int(constants.TaggedInPost):
		return TaggedYouOnPost
	case int(constants.TaggedInPostComment):
		return TaggedYouOnComment
	case int(constants.AlsoCommentOnPost):
		return AlsoCommentedOnPost
	case int(constants.RepostOnPost):
		return RepostOnPost
	}

	return ""
}

// To string method for ActivityAction
func (aa ActivityAction) ToString() string {
	return string(aa)
}
