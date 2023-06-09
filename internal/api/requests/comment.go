package requests

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Request Structure for Create Comment
type CreateCommentRequest struct {
	Text string `json:"text" binding:"required"`
}

// Request Structure for Edit Comment
type EditCommentRequest struct {
	Text     string `json:"text" binding:"required"`
	UserIsCm bool   `json:"user_is_cm"`
}

// Request Structure for Delete Comment
type DeleteCommentRequest struct {
	UserIsCm     bool   `json:"user_is_cm"`
	DeleteReason string `json:"delete_reason"`
}

// Response Structure for Comment
type CommentResponse struct {
	ID            primitive.ObjectID `json:"_id"`
	Text          string             `json:"text"`
	Level         int                `json:"level"`
	UserId        string             `json:"user_id"`
	UUID          string             `json:"uuid"`
	CommunityId   int                `json:"community_id,omitempty"`
	PostId        primitive.ObjectID `json:"post_id,omitempty"`
	IsLiked       bool               `json:"is_liked"`
	LikesCount    int                `json:"likes_count"`
	CommentsCount int                `json:"comments_count"`
	IsEdited      bool               `json:"is_edited"`
	IsDeleted     bool               `json:"is_deleted,omitempty"`
	DeletedBy     string             `json:"deleted_by,omitempty"`
	DeleteReason  string             `json:"delete_reason,omitempty"`
	MenuItems     []MenuResponse     `json:"menu_items"`
	CreatedAt     int                `json:"created_at"`
	UpdatedAt     int                `json:"updated_at"`
}

// Response Structure for Fetch Comment
type FetchCommentResponse struct {
	CommentResponse
	Post          interface{}       `json:"post_data,omitempty"`
	ParentComment *CommentResponse  `json:"parent_comment,omitempty"`
	Replies       []CommentResponse `json:"replies"`
}

type FetchCommentsResponse struct {
	CommentResponse
	ParentComment *CommentResponse `json:"parent_comment,omitempty"`
}
