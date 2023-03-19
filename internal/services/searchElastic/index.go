package searchElastic

import "time"

// Struct for Elasticsearch Post Index fields
type PostIndex struct {
	Id          string      `json:"id"`
	Text        string      `json:"text"`
	Heading     string      `json:"heading"`
	ChatroomId  int         `json:"chatroom_id"`
	CommunityId int         `json:"community_id"`
	IsPinned    bool        `json:"is_pinned"`
	UserId      string      `json:"user_id"`
	Attachments interface{} `json:"attachments"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
