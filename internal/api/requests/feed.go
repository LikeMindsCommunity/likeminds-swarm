package requests

// Request Structure for Universal Feed
type FetchUniversalFeedRequest struct {
	IsCm      bool   `form:"user_is_cm"`
	TopicIds  string `form:"topic_ids"`
	WidgetIds string `form:"widget_ids"`
}

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

// Request Structure for Group Feed
type FetchGroupFeedRequest struct {
	IsCm       bool   `form:"user_is_cm"`
	FeedroomId string `form:"feedroom_id"`
	TopicIds   string `form:"topic_ids"`
}

// Request Structure for Delete User Data
type DeleteUserDataRequest struct {
	UserIsCm bool     `json:"user_is_cm"`
	UserIds  []string `json:"user_ids"`
}

// Request Structure for Follow Feed
type FetchConnectionFeedRequest struct {
	IsCm bool `form:"user_is_cm"`
}
