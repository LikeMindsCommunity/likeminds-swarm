package requests

// Request Structure for Create Poll Option
type CreatePollOptionRequest struct {
	Text     string `json:"text" binding:"required"`
	UserIsCm bool   `json:"user_is_cm"`
}

// Request Structure for Create Vote on Poll
type CreateVoteOnPollRequest struct {
	Votes []string `json:"votes" binding:"required"`
}

// Request Structure for Fetch Votes on Poll
type GetPollVotesRequest struct {
	Votes string `form:"votes"`
}
