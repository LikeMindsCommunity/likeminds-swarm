package constants

const FeedCategory = "Feed"

const (
	CommentPermissionRemovedSubCategory = "%s permission removed"
	CommentPermissionAddedSubCategory   = "%s permission added"
	PostPermissionRemovedSubCategory    = "%s permission removed"
	PostPermissionAddedSubCategory      = "%s permission added"
	ModerationPostDeleteSubCategory     = "Moderation delete %s"
	ModerationCommentDeleteSubCategory  = "Moderation delete %s"
	ModerationReplyDeleteSubCategory    = "Moderation delete reply"
	PostTagSubCategory                  = "%s Tag"
	CommentTagSubCategory               = "%s Tag"
	ReplyTagSubCategory                 = "Reply Tag"
	AlsoCommentSubCategory              = "Followed %s"
	PostCommentSubCategory              = "%s %s"
	CommentReplySubCategory             = "%s Reply"
	PostLikedSubCategory                = "%s Liked"
	CommentLikedSubCategory             = "%s Liked"
	RepostOnPostSubCategory             = "%s Reposted"
	PendingPostApprovedSubCategory      = "Pending %s Approved"
	PendingPostRejectedSubCategory      = "Pending %s Rejected"
)

const (
	PermissionUpdatedTitle   = "Permission updated"
	PostDeletedTitle         = "%s deleted"
	CommentDeletedTitle      = "%s deleted"
	ReplyDeletedTitle        = "Reply deleted"
	LikeTitle                = "New Like!"
	CommentTitle             = "New %s!"
	ReplyTitle               = "New Reply!"
	TagTitle                 = "You are tagged!"
	RepostTitle              = "New Repost!"
	PendingPostApprovedTitle = "%s approved!"
	PendingPostRejectedTitle = "%s rejected!"
)

const (
	CommentPermissionRemovedSubTitle = "Your permission to add %ss and replies to the %ss has been removed."
	CommentPermissionAddedSubTitle   = "You now have the permission to add %ss on the %ss. Start engaging now."
	PostPermissionRemovedSubTitle    = "Your permission to create %ss in the community has been removed."
	PostPermissionAddedSubTitle      = "You now have the permission to create %ss in the community. Start posting now."
	ModerationPostDeleteSubTitle     = "Your %s has been deleted as it violates community guidelines. Reason: %s"
	ModerationCommentDeleteSubTitle  = "Your %s has been deleted as it violates community guidelines. Reason: %s"
	ModerationReplyDeleteSubTitle    = "Your Reply has been deleted as it violates community guidelines. Reason: %s"
	PostTagSubTitle                  = "%s tagged you in a %s."
	CommentTagSubTitle               = "%s tagged you in a %s."
	ReplyTagSubTitle                 = "%s tagged you in a reply."
	AlsoCommentSubTitleLevelOne      = "%s also left a %s on %s's %s."
	AlsoCommentSubTitleLevelTwo      = "%s and 1 other also left %ss on %s's %s."
	AlsoCommentSubTitleLevelThree    = "%s and %d others also left %ss on %s's %s."
	PostCommentSubTitleLevelOne      = "%s left a %s on your %s."
	PostCommentSubTitleLevelTwo      = "%s and 1 other left %ss on your %s."
	PostCommentSubTitleLevelThree    = "%s and %d others left %ss on your %s."
	CommentReplySubTitleLevelOne     = "%s replied to your %s."
	CommentReplySubTitleLevelTwo     = "%s and 1 other replied to your %s."
	CommentReplySubTitleLevelThree   = "%s and %d others replied to your %s."
	PostLikedSubTitleLevelOne        = "%s liked your %s."
	PostLikedSubTitleLevelTwo        = "%s and 1 other liked your %s."
	PostLikedSubTitleLevelThree      = "%s and %d others liked your %s."
	CommentLikedSubTitleLevelOne     = "%s liked your %s."
	CommentLikedSubTitleLevelTwo     = "%s and 1 other liked your %s."
	CommentLikedSubTitleLevelThree   = "%s and %d others liked your %s."
	PostRepostedSubTitleLevelOne     = "%s reposted your %s."
	PostRepostedSubTitleLevelTwo     = "%s and 1 other reposted your %s."
	PostRepostedSubTitleLevelThree   = "%s and %d others reposted your %s."
	PendingPostApprovedSubTitle      = "Your %s has been approved."
	PendingPostRejectedSubTitle      = "Your %s has been rejected."
)

// routes
const (
	PlaceholderHomeRoute = "route://home"
)
