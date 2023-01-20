package requests

type FetchExploreFeedRequest struct {
	OrderType           int    `form:"order_type"`
	ChatroomIDs         string `form:"chatroom_ids"`
	ExcludedChatroomIDs string `form:"excluded_chatroom_ids"`
}

type FetchExploreFeedResponse struct {
	Success     bool        `json:"success"`
	ChatroomIDs []int       `json:"chatroom_ids"`
	PostCounts  map[int]int `json:"post_counts"`
}
