package requests

// Request Structure for Fetch Explore Feed
type FetchExploreFeedRequest struct {
	OrderType           int    `form:"order_type"`
	ChatroomIDs         string `form:"chatroom_ids"`
	ExcludedChatroomIDs string `form:"excluded_chatroom_ids"`
}

// Response Structure for Fetch Explore Feed
type FetchExploreFeedResponse struct {
	Success     bool        `json:"success"`
	ChatroomIDs []int       `json:"chatroom_ids"`
	PostCounts  map[int]int `json:"post_counts"`
}
