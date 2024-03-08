package responses

import "go.mongodb.org/mongo-driver/bson/primitive"

type TopCommentIDResponse struct {
	CommentID primitive.ObjectID `json:"comment_id" bson:"comment_id"`
}

type TopCommentsAggregationQueryResponse struct {
	PostID      primitive.ObjectID     `json:"_id" bson:"_id"`
	TopComments []TopCommentIDResponse `json:"top_comments" bson:"top_comments"`
}
