package requests

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateCommentRequest struct {
	Text string `json:"text" binding:"required"`
}

type DeleteCommentRequest struct {
	UserIsCm     bool   `json:"user_is_cm"`
	DeleteReason string `json:"delete_reason"`
}

type CommentResponse struct {
	ID            primitive.ObjectID `json:"_id"`
	Text          string             `json:"text"`
	Level         int                `json:"level"`
	UserId        string             `json:"user_id"`
	LikesCount    int                `json:"likes_count"`
	CommentsCount int                `json:"comments_count"`
	IsDeleted     bool               `json:"is_deleted,omitempty"`
	DeletedBy     string             `json:"deleted_by,omitempty"`
	DeleteReason  string             `json:"delete_reason,omitempty"`
	CreatedAt     int                `json:"created_at"`
	UpdatedAt     int                `json:"updated_at"`
}

type FetchCommentResponse struct {
	CommentResponse
	Post          interface{}       `json:"post_data,omitempty"`
	ParentComment *CommentResponse  `json:"parent_comment,omitempty"`
	Replies       []CommentResponse `json:"replies"`
}
