package constants

const ActivityEntityType string = "activity"
const (
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
)

// EntityType | defines enum for activity entities
type EntityType uint8

const (

	// Post | post entity
	Post EntityType = 0

	// Comment | comment entity
	Comment EntityType = 1

	// User | user entity
	User EntityType = 2
)

// ActivityAction | defines enum for activity actions
type ActivityAction uint8

const (

	// DefaultAction | placeholder value
	DefaultAction ActivityAction = 99

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
)
