package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Response Structure for Menu Item of an Entity
type MenuResponse struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Route string `json:"route,omitempty"`
}

// Resonse Structure for Post
type PostResponse struct {
	ID                 primitive.ObjectID   `json:"_id"`
	TempID             *string              `json:"temp_id"`
	Topics             []primitive.ObjectID `json:"topics"`
	Text               string               `json:"text"`
	Heading            string               `json:"heading"`
	CommunityId        int                  `json:"community_id,omitempty"`
	ChatroomId         int                  `json:"feedroom_id,omitempty"`
	IsPinned           bool                 `json:"is_pinned"`
	UserId             string               `json:"user_id,omitempty"`
	UUID               string               `json:"uuid,omitempty"`
	Attachments        []AttachmentResponse `json:"attachments"`
	LikesCount         int                  `json:"likes_count"`
	CommentsCount      int                  `json:"comments_count"`
	RepostCount        int                  `json:"repost_count"`
	IsDeleted          bool                 `json:"is_deleted,omitempty"`
	IsEdited           bool                 `json:"is_edited"`
	IsRepost           bool                 `json:"is_repost"`
	IsAnonymous        bool                 `json:"is_anonymous"`
	IsRepostedByUser   bool                 `json:"is_reposted_by_user"`
	OriginalAuthorUUID string               `json:"original_author_uuid,omitempty"`
	DeletedBy          string               `json:"deleted_by,omitempty"`
	DeletedByUUID      string               `json:"deleted_by_uuid,omitempty"`
	DeleteReason       string               `json:"delete_reason,omitempty"`
	IsLiked            bool                 `json:"is_liked"`
	IsSaved            bool                 `json:"is_saved"`
	MenuItems          []MenuResponse       `json:"menu_items"`
	CreatedAt          int                  `json:"created_at"`
	UpdatedAt          int                  `json:"updated_at"`
	CommentIDs         []string             `json:"comment_ids"`
	IsPendingPost      bool                 `json:"is_pending_post"`
	PostStatus         string               `json:"post_status"`
}

// Response Structure for Fetch Post
type PostWithRepliesResponse struct {
	PostResponse
	Replies []CommentResponse `json:"replies"`
}

// Response Structure for Fetch Multiple Post - Deprecated
type FetchUserMultiplePostResponse struct {
	Success    bool           `json:"success"`
	TotalCount int            `json:"total_count,omitempty"`
	Posts      []PostResponse `json:"posts"`
}
