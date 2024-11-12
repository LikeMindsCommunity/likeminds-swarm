package constants

// Not being used anywhere | can be deprecated
const ActivityEntityType string = "activity"
const (
	CustomActivityAction               string = "custom_activity"
	LikeAction                         string = "like"
	CommentAction                      string = "comment"
	AlsoCommentAction                  string = "also_comment"
	TagAction                          string = "tag"
	DeleteAction                       string = "delete"
	SaveAction                         string = "save"
	CreatePostPermitAddedAction        string = "create_post_permit_added"
	CreatePostPermitRemovedAction      string = "create_post_permit_removed"
	CreateCommentPermissionAddedAction string = "create_comment_permit_added"
	CreateCommentPermitRemovedAction   string = "create_comment_permit_removed"
	RepostedPostAction                 string = "reposted_post"
)

// EntityType | constants for activity entities
type EntityType uint8

const (
	DefaultEntity EntityType = 99

	// PostEntity | post entity
	PostEntity EntityType = 0

	// CommentEntity | comment entity
	CommentEntity EntityType = 1

	// UserEntity | user entity
	UserEntity EntityType = 2

	// PendingPostEntity | Pending post entity
	PendingPostEntity EntityType = 3
)

// ActivityAction | constants for activity actions
type ActivityAction uint8

const (

	// DefaultAction | placeholder value
	CustomActivity ActivityAction = 99

	// CreatePostPermitAdded | create post permission added
	CreatePostPermitAdded ActivityAction = 0

	// CreatePostPermitRemoved | create post permission removed
	CreatePostPermitRemoved ActivityAction = 1

	// CreateCommentPermitAdded | create comment permission added
	CreateCommentPermitAdded ActivityAction = 2

	// CreateCommentPermitRemoved | create comment permission removed
	CreateCommentPermitRemoved ActivityAction = 3

	// CMDeletedPost | community manager deleted post
	CMDeletedPost ActivityAction = 4

	// CMDeletedComment | community manager deleted comment
	CMDeletedComment ActivityAction = 5

	// LikeOnPost | like added on post
	LikeOnPost ActivityAction = 6

	// CommentOnPost | comment added on post
	CommentOnPost ActivityAction = 7

	// LikeOnComment | like added on comment
	LikeOnComment ActivityAction = 8

	// CommentOnComment | comment added on comment
	CommentOnComment ActivityAction = 9

	// TaggedInPost | tagged in a post text
	TaggedInPost ActivityAction = 10

	// TaggedInPostComment | tagged in a post comment text
	TaggedInPostComment ActivityAction = 11

	// AlsoCommentOnPost | level 0 comment added on a commented post
	AlsoCommentOnPost ActivityAction = 12

	// RepostOnPost | repost on post
	RepostOnPost ActivityAction = 13

	// PendingPostAccepted | Accepted the pending post
	PendingPostAccepted ActivityAction = 14

	// PendingPostRejected | Rejected the pending post
	PendingPostRejected ActivityAction = 15
)

// ActivityCacheKey | cache key for activity instance
const ActivityCacheKey = "activity_%s"

// UserActivityFeedCacheKey | cache key for user activity feed
const UserActivityFeedCacheKey = "user_%s_activity_feed"
