package requests

// Request structure for approve/reject a pending post
type ApproveRejectPendingPostRequest struct {
	PendingPostIds []string `json:"pending_post_ids"`
	Status         string   `json:"status"`
}
