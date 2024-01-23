package requests

// Request structure for approve/reject a pending post
type ApproveRejectPendingPostRequest struct {
	Status string `json:"status"`
}
