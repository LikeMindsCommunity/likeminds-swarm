package requests

type FetchExploreFeedRequest struct {
	OrderType           int   `form:"order_type" binding:"required"`
	ChatroomIDs         []int `form:"chatroom_ids"`
	ExcludedChatroomIDs []int `form:"excluded_chatroom_ids"`
}

type FetchExploreFeedResponse struct {
	Success     bool        `json:"success"`
	ChatroomIDs []int       `json:"chatroom_ids"`
	PostCounts  map[int]int `json:"post_counts"`
}
