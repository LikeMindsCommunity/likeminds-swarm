package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Response Structure for Comment
type CommentResponse struct {
	ID            primitive.ObjectID `json:"_id"`
	TempID        *string            `json:"temp_id"`
	Text          string             `json:"text"`
	Level         int                `json:"level"`
	UserId        string             `json:"user_id,omitempty"`
	UUID          string             `json:"uuid"`
	CommunityId   int                `json:"community_id,omitempty"`
	PostId        primitive.ObjectID `json:"post_id,omitempty"`
	IsLiked       bool               `json:"is_liked"`
	Attachments   []Attachment       `json:"attachments"`
	LikesCount    int                `json:"likes_count"`
	CommentsCount int                `json:"comments_count"`
	IsEdited      bool               `json:"is_edited"`
	IsDeleted     bool               `json:"is_deleted,omitempty"`
	DeletedBy     string             `json:"deleted_by,omitempty"`
	DeletedByUUID string             `json:"deleted_by_uuid,omitempty"`
	DeleteReason  string             `json:"delete_reason,omitempty"`
	MenuItems     []MenuResponse     `json:"menu_items"`
	CreatedAt     int                `json:"created_at"`
	UpdatedAt     int                `json:"updated_at"`
}

// Response Structure for Comment with Parent
type CommentWithParentResponse struct {
	CommentResponse
	ParentComment *CommentResponse `json:"parent_comment,omitempty"`
}

// Response Structure for Fetch Comment
type FetchCommentResponse struct {
	CommentResponse
	Post          *PostResponse     `json:"post_data,omitempty"`
	ParentComment *CommentResponse  `json:"parent_comment,omitempty"`
	Replies       []CommentResponse `json:"replies"`
}

type TopCommentIDResponse struct {
	CommentID primitive.ObjectID `json:"comment_id" bson:"comment_id"`
}

type TopCommentsAggregationQueryResponse struct {
	PostID      primitive.ObjectID     `json:"_id" bson:"_id"`
	TopComments []TopCommentIDResponse `json:"top_comments" bson:"top_comments"`
}
