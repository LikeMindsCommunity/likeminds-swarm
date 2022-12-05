package requests

type CreateActivityRequest struct {
	Action string `json:"action" binding:"required"`
}
