package requests

// Request Structure for Update Connection
type UpdateConnectionRequest struct {
	Status         string `json:"status"`
	ConnectionType string `json:"connection_type"`
}
