package requests

import "go.mongodb.org/mongo-driver/bson/primitive"

// Response Structure for Likes
type LikeResponse struct {
	ID        primitive.ObjectID `json:"_id"`
	UserId    string             `json:"user_id"`
	CreatedAt int                `json:"created_at"`
	UpdatedAt int                `json:"updated_at"`
	UUID      string             `json:"uuid"`
}

// Response Structure for Fetch Likes
type FetchLikesResponse struct {
	Success    bool           `json:"success"`
	TotalCount int            `json:"total_count"`
	Likes      []LikeResponse `json:"likes"`
}
